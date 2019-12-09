/*
Package seo4ajax provides a library for accessing the SEO4Ajax prerender service.
Before using, you need to set ServerIp to a valid IP address.
*/
package seo4ajax

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/cenkalti/backoff"
	"github.com/go-kit/kit/log"
)

var (
	// ErrNoToken is returned when the client isn't provided a API token
	ErrNoToken = errors.New("no token given")
	// ErrCacheMiss happens if seo4ajax responded with a cache miss
	ErrCacheMiss = errors.New("cache miss from seo4ajax")
	// ErrUnknownStatus represents an unknown status code
	ErrUnknownStatus = errors.New("Unknown Status Code")
	errRedirect      = errors.New("SEO4AJAX: do not follow redirect")

	regexInvalidUserAgent = regexp.MustCompile(`(?i:bing|msnbot|yandexbot|pinterest.*ios|mail\.ru)`)
	regexValidUserAgent   = regexp.MustCompile(`(?i:bot|google|crawler|spider|archiver|pinterest|facebookexternalhit|flipboardproxy)`)
	regexPath             = regexp.MustCompile(`.*(\.[^?]{2,4}$|\.[^?]{2,4}?.*)`)
)

// Config is the Seo4Ajax Client config
type Config struct {
	Log       log.Logger
	Next      http.Handler
	Transport http.RoundTripper
	Server    string        // seo4ajax api server, defaults to http://api.seo4ajax.com
	Token     string        // seo4ajax token, must be set
	IP        string        // server IP, defaults to 127.0.0.1
	Timeout   time.Duration // retry timeout, defaults to 30s
	// s4a supports client side caching and returns an empty 304 if the content hasn't changed.
	// If UnconditionalFetch set to true the client side caching headers (If-Modified-Since and If-None-Match)
	// are removed
	UnconditionalFetch bool
	// FetchErrorStatus is the http status code returned if the fetch from seo4ajax fails
	FetchErrorStatus int
	// FetchTimeout is the http timeout for a single fetch attempt
	FetchTimeout time.Duration
	// RetryUnavailable advises the retry loop to retry a fetch on 503 upstream results until success or Timeout
	RetryUnavailable bool
}

// Client is the Seo4Ajax Client
type Client struct {
	log                log.Logger
	next               http.Handler
	server             string
	token              string
	ip                 string
	timeout            time.Duration
	http               *http.Client
	unconditionalFetch bool
	fetchErrorStatus   int
	retryUnavailable   bool
}

// New creates a new Seo4Ajax client. Returns an error if no token is provided
func New(cfg Config) (*Client, error) {
	if cfg.Log == nil {
		cfg.Log = log.NewNopLogger()
	}
	if cfg.Server == "" {
		cfg.Server = "http://api.seo4ajax.com"
	}
	if cfg.Token == "" {
		return nil, ErrNoToken
	}
	if cfg.IP == "" {
		cfg.IP = "127.0.0.1"
	}
	if cfg.Transport == nil {
		cfg.Transport = http.DefaultTransport
	}
	if cfg.FetchErrorStatus == 0 {
		cfg.FetchErrorStatus = http.StatusServiceUnavailable
	}

	c := &Client{
		log:                cfg.Log,
		server:             cfg.Server,
		token:              cfg.Token,
		ip:                 cfg.IP,
		timeout:            cfg.Timeout,
		next:               cfg.Next,
		unconditionalFetch: cfg.UnconditionalFetch,
		fetchErrorStatus:   cfg.FetchErrorStatus,
		retryUnavailable:   cfg.RetryUnavailable,
	}
	c.http = &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return errRedirect
		},
		Transport: cfg.Transport,
	}
	if cfg.FetchTimeout > 0 {
		c.http.Timeout = cfg.FetchTimeout
	}
	return c, nil
}

// IsPrerender returns true, when Seo4Ajax shall be used for the given http Request.
// The logic is taken from https://github.com/seo4ajax/connect-s4a/blob/master/lib/connect-s4a.js
func IsPrerender(r *http.Request) bool {
	if r.Method != "GET" && r.Method != "HEAD" {
		return false
	}

	if strings.Contains(r.URL.RawQuery, "_escaped_fragment_") {
		return true
	}

	if regexInvalidUserAgent.MatchString(r.Header.Get("User-Agent")) {
		return false
	}

	if regexPath.MatchString(r.URL.Path) {
		return false
	}

	return regexValidUserAgent.MatchString(r.Header.Get("User-Agent"))
}

// ServeHTTP will serve the prerendered page if this is a prerender request.
// If no upstream handler is set it will return an error. Otherwise it will
// just invoke the upstream handler. This way it can be either used as an
// HTTP middleware intercepting any prerender requests or an regular HTTP
// handler (if next is nil) to serve only prerender request
func (c *Client) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if IsPrerender(r) {
		c.GetPrerenderedPage(w, r)
		return
	}

	if c.next == nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	c.next.ServeHTTP(w, r)
	return
}

// GetPrerenderedPage returns the prerendered html from the seo4ajax api
func (c *Client) GetPrerenderedPage(w http.ResponseWriter, r *http.Request) {
	var outputStarted bool
	opFunc := func() error {
		req, err := http.NewRequest("GET", fmt.Sprintf("%s/%s%s", c.server, c.token, cleanPath(r.URL)), nil)
		if err != nil {
			return err
		}

		req.Header = r.Header
		ips := []string{c.ip}
		if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
			ips = append(ips, xff)
		}
		req.Header.Set("X-Forwarded-For", strings.Join(ips, ", "))

		if c.unconditionalFetch {
			req.Header.Del("If-Modified-Since")
			req.Header.Del("If-None-Match")
		}

		resp, err := c.http.Do(req)
		if err != nil && !strings.HasSuffix(err.Error(), errRedirect.Error()) {
			return err
		}
		defer resp.Body.Close()

		if resp.StatusCode == http.StatusFound {
			http.Redirect(w, r, resp.Header.Get("Location"), http.StatusFound)
			return nil
		}

		// conditionally terminate retry loop if the status code is 503 or 404
		if !c.retryUnavailable {
			if resp.StatusCode == http.StatusServiceUnavailable {
				return backoff.Permanent(errors.New("page not yet rendered"))
			}
			if resp.StatusCode == http.StatusNotFound {
				return backoff.Permanent(errors.New("page not found"))
			}
		}

		if resp.StatusCode != http.StatusOK {
			// retry
			return fmt.Errorf("expected 200 status code, got %d", resp.StatusCode)
		}

		for header, val := range resp.Header {
			w.Header()[header] = val
		}

		outputStarted = true
		// as soon as we start writing the body we must return nil, otherwise we'll
		// mess up the HTTP response by calling response.WriteHeader multiple times
		_, err = io.Copy(w, resp.Body)
		return err
	}

	bo := backoff.NewExponentialBackOff()
	bo.InitialInterval = 50 * time.Millisecond
	bo.MaxInterval = 30 * time.Second
	if c.timeout > 0 {
		bo.MaxElapsedTime = c.timeout
	}
	err := backoff.Retry(opFunc, bo)
	if err != nil {
		c.log.Log("level", "warn", "msg", "Upstream request failed", "err", err, "path", r.URL.Path)
		if !outputStarted {
			http.Error(w, "Upstream error", c.fetchErrorStatus)
			return
		}
	}
	return
}

func cleanPath(u *url.URL) string {
	cpy := *u
	if len(cpy.Path) == 0 {
		cpy.Path = "/"
	} else if cpy.Path[0] != '/' {
		cpy.Path = "/" + cpy.Path
	}

	cpy.Scheme = ""
	cpy.Opaque = ""
	cpy.User = nil
	cpy.Host = ""

	return cpy.String()
}
