package seo4ajax

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strings"
)

var (
	apiHost = "http://api.seo4ajax.com"

	regexInvalidUserAgent = regexp.MustCompile(`(?i:google.*bot|bing|msnbot|yandexbot|pinterest.*ios|mail\.ru)`)
	regexValidUserAgent   = regexp.MustCompile(`(?i:bot|crawler|spider|archiver|pinterest|facebookexternalhit|flipboardproxy)`)
	regexPath             = regexp.MustCompile(`.*(\.[^?]{2,4}$|\.[^?]{2,4}?.*)`)
	token                 = os.Getenv("SEO4AJAX_TOKEN")

	ErrNoToken   = errors.New("no token given")
	ErrCacheMiss = errors.New("cache miss from seo4ajax")
	errRedirect  = errors.New("SEO4AJAX: do not follow redirect")

	client = &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return errRedirect
		},
	}
)

// IsPrerender returns true, when Seo4Ajax shall be used for the given http Request.
// The logic is taken from https://github.com/seo4ajax/connect-s4a/blob/master/lib/connect-s4a.js
func IsPrerender(req *http.Request) (usePrerender bool) {
	if req.Method != "GET" && req.Method != "HEAD" {
		return
	}

	usePrerender = strings.Contains(req.URL.RawQuery, "_escaped_fragment_")
	if usePrerender {
		return
	}

	if regexInvalidUserAgent.MatchString(req.Header.Get("User-Agent")) {
		return
	}

	if regexPath.MatchString(req.URL.Path) {
		return
	}

	usePrerender = regexValidUserAgent.MatchString(req.Header.Get("User-Agent"))
	return
}

// GetPrerenderedPage returns the prerendered html from the seo4ajax api. If no
// token is given, it will return an error.
func GetPrerenderedPage(w http.ResponseWriter, req *http.Request) (err error) {
	if token == "" {
		err = ErrNoToken
		return
	}

	var prerenderRequest *http.Request
	prerenderRequest, err = http.NewRequest("GET", fmt.Sprintf("%s/%s%s", apiHost, token, cleanPath(req.URL)), nil)
	if err != nil {
		return
	}

	prerenderRequest.Header = req.Header

	xForwardedFor := req.Header.Get("X-Forwarded-For")
	if xForwardedFor != "" {
		xForwardedFor = fmt.Sprintf("%s, %s", req.URL.Host, xForwardedFor)
	} else {
		xForwardedFor = req.URL.Host
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

	for header, val := range resp.Header {
		w.Header()[header] = val
	}

	switch resp.StatusCode {
	case 200:
		_, err = io.Copy(w, resp.Body)
	case 302:
		http.Redirect(w, req, resp.Header.Get("Location"), resp.StatusCode)
	case 401, 403:
		err = ErrNoToken
	case 503:
		err = ErrCacheMiss
	}

	if err != nil {
		code := resp.StatusCode
		if code == 200 {
			code = http.StatusInternalServerError
		}
		http.Error(w, err.Error(), code)
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
