/*
	Package seo4ajax provides a library for accessing the SEO4Ajax prerender service.
	Before using, you need to set ServerIp to a valid IP address.
*/
package seo4ajax

import (
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"regexp"
	"strings"
)

var (
	ErrNoToken = errors.New("no token given")
	// seo4ajax responded with a cache miss
	ErrCacheMiss = errors.New("cache miss from seo4ajax")
	errRedirect  = errors.New("SEO4AJAX: do not follow redirect")

	regexInvalidUserAgent = regexp.MustCompile(`(?i:google.*bot|bing|msnbot|yandexbot|pinterest.*ios|mail\.ru)`)
	regexValidUserAgent   = regexp.MustCompile(`(?i:bot|crawler|spider|archiver|pinterest|facebookexternalhit|flipboardproxy)`)
	regexPath             = regexp.MustCompile(`.*(\.[^?]{2,4}$|\.[^?]{2,4}?.*)`)

	client = &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error { return errRedirect },
	}
)

type Client struct {
	APIHost         string
	serverIP, token string
}

func New(serverIP, token string) (*Client, error) {
	if serverIP == "" || net.ParseIP(serverIP) == nil {
		return nil, errors.New("no ip address")
	}

	if token == "" {
		return nil, ErrNoToken
	}

	return &Client{
		APIHost:  "http://api.seo4ajax.com",
		serverIP: serverIP,
		token:    token,
	}, nil
}

// IsPrerender returns true, when Seo4Ajax shall be used for the given http Request.
// The logic is taken from https://github.com/seo4ajax/connect-s4a/blob/master/lib/connect-s4a.js
func IsPrerender(req *http.Request) (usePrerender bool) {
	if req.Method != "GET" && req.Method != "HEAD" {
		return false
	}

	if strings.Contains(req.URL.RawQuery, "_escaped_fragment_") {
		return true
	}

	if regexInvalidUserAgent.MatchString(req.Header.Get("User-Agent")) {
		return false
	}

	if regexPath.MatchString(req.URL.Path) {
		return false
	}

	return regexValidUserAgent.MatchString(req.Header.Get("User-Agent"))
}

// GetPrerenderedPage returns the prerendered html from the seo4ajax api. If no
// token is given, it will return an error.
func (c *Client) GetPrerenderedPage(w http.ResponseWriter, req *http.Request) (err error) {
	var prerenderRequest *http.Request
	prerenderRequest, err = http.NewRequest("GET", fmt.Sprintf("%s/%s%s", c.APIHost, c.token, cleanPath(req.URL)), nil)
	if err != nil {
		return
	}

	prerenderRequest.Header = req.Header

	xForwardedFor := req.Header.Get("X-Forwarded-For")
	if xForwardedFor != "" {
		xForwardedFor = fmt.Sprintf("%s, %s", c.serverIP, xForwardedFor)
	} else {
		xForwardedFor = c.serverIP
	}
	prerenderRequest.Header.Set("X-Forwarded-For", xForwardedFor)

	var resp *http.Response
	resp, err = client.Do(prerenderRequest)
	if err != nil && !strings.HasSuffix(err.Error(), errRedirect.Error()) {
		return
	} else {
		err = nil
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case 200:
		for header, val := range resp.Header {
			w.Header()[header] = val
		}

		_, err = io.Copy(w, resp.Body)
	case 302:
		http.Redirect(w, req, resp.Header.Get("Location"), resp.StatusCode)
	case 401, 403:
		err = ErrNoToken
	case 503:
		err = ErrCacheMiss
	}

	if err != nil {
		http.Error(w, err.Error(), 503)
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
