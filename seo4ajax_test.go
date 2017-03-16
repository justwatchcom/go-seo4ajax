package seo4ajax

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

var (
	port      = ":8080"
	appAdress = "127.0.0.1:3000"
)

func TestIsPrerender(t *testing.T) {
	serverIP := "127.0.0.1"

	Convey("_escaped_fragment_ urls properly proxified", t, func() {
		Convey("without _escaped_fragment_", func() {
			req, err := http.NewRequest("GET", "http://"+appAdress+"/path?withQuery=parameter", nil)
			So(err, ShouldBeNil)
			So(req, ShouldNotBeNil)
			So(IsPrerender(req), ShouldBeFalse)
		})

		Convey("with _escaped_fragment_", func() {
			req, err := http.NewRequest("GET", "http://"+appAdress+"/path?_escaped_fragment_=fragment", nil)
			So(err, ShouldBeNil)
			So(req, ShouldNotBeNil)
			So(IsPrerender(req), ShouldBeTrue)
		})

		Convey("_escaped_fragment_ without value", func() {
			req, err := http.NewRequest("GET", "http://"+appAdress+"/path?_escaped_fragment_=", nil)
			So(err, ShouldBeNil)
			So(req, ShouldNotBeNil)
			So(IsPrerender(req), ShouldBeTrue)
		})

		Convey("_escaped_fragment_ as a second parameter", func() {
			req, err := http.NewRequest("GET", "http://"+appAdress+"/path?param1=val1&_escaped_fragment_=", nil)
			So(err, ShouldBeNil)
			So(req, ShouldNotBeNil)
			So(IsPrerender(req), ShouldBeTrue)
		})

		Convey("with a HEAD request", func() {
			req, err := http.NewRequest("HEAD", "http://"+appAdress+"/path?param1=val1&_escaped_fragment_=", nil)
			So(err, ShouldBeNil)
			So(req, ShouldNotBeNil)
			So(IsPrerender(req), ShouldBeTrue)
		})

		Convey("with a POST request", func() {
			req, err := http.NewRequest("POST", "http://"+appAdress+"/path?param1=val1&_escaped_fragment_=", nil)
			So(err, ShouldBeNil)
			So(req, ShouldNotBeNil)
			So(IsPrerender(req), ShouldBeFalse)
		})
	})

	Convey("urls filtered by user-agent properly proxified", t, func() {
		Convey("Google bot", func() {
			req, err := http.NewRequest("GET", "http://"+appAdress+"/path/subpath", nil)
			So(err, ShouldBeNil)
			So(req, ShouldNotBeNil)
			req.Header.Add("User-Agent", "Mozilla/5.0 (compatible; Googlebot/2.1; +http://www.google.com/bot.html)")
			So(IsPrerender(req), ShouldBeFalse)
		})

		Convey("Bing bot", func() {
			req, err := http.NewRequest("GET", "http://"+appAdress+"/path/subpath", nil)
			So(err, ShouldBeNil)
			So(req, ShouldNotBeNil)
			req.Header.Add("User-Agent", "Mozilla/5.0 (compatible; bingbot/2.0; +http://www.bing.com/bingbot.htm)")
			So(IsPrerender(req), ShouldBeFalse)
		})

		Convey("Google bot mobile", func() {
			req, err := http.NewRequest("GET", "http://"+appAdress+"/path/subpath", nil)
			So(err, ShouldBeNil)
			So(req, ShouldNotBeNil)
			req.Header.Add("User-Agent", "Mozilla/5.0 (iPhone; CPU iPhone OS 6_0 like Mac OS X) AppleWebKit/536.26 (KHTML, like Gecko) Version/6.0 Mobile/10A5376e Safari/8536.25 (compatible; Googlebot/2.1; +http://www.google.com/bot.html)")
			So(IsPrerender(req), ShouldBeFalse)
		})

		Convey("Yandex bot", func() {
			req, err := http.NewRequest("GET", "http://"+appAdress+"/path/subpath", nil)
			So(err, ShouldBeNil)
			So(req, ShouldNotBeNil)
			req.Header.Add("User-Agent", "Mozilla/5.0 (compatible; YandexBot/3.0; +http://yandex.com/bots)")
			So(IsPrerender(req), ShouldBeFalse)
		})

		Convey("Mail.RU bot", func() {
			req, err := http.NewRequest("GET", "http://"+appAdress+"/path/subpath", nil)
			So(err, ShouldBeNil)
			So(req, ShouldNotBeNil)
			req.Header.Add("User-Agent", "Mozilla/5.0 (compatible; Linux x86_64; Mail.RU_Bot/2.0; +http://go.mail.ru/help/robots)")
			So(IsPrerender(req), ShouldBeFalse)
		})

		Convey("Pinterest iOS App", func() {
			req, err := http.NewRequest("GET", "http://"+appAdress+"/path/subpath", nil)
			So(err, ShouldBeNil)
			So(req, ShouldNotBeNil)
			req.Header.Add("User-Agent", "Mozilla/5.0 (iPad; CPU OS 7_0 like Mac OS X) AppleWebKit/537.51.1 (KHTML, like Gecko) Mobile/11A465 [Pinterest/iOS]")
			So(IsPrerender(req), ShouldBeFalse)
		})

		Convey("Flipboard Android App", func() {
			req, err := http.NewRequest("GET", "http://"+appAdress+"/path/subpath", nil)
			So(err, ShouldBeNil)
			So(req, ShouldNotBeNil)
			req.Header.Add("User-Agent", "Mozilla/5.0 (Linux; U; Android 4.3; en-us; SAMSUNG-SGH-I337 Build/JSS15J) AppleWebKit/534.30 (KHTML, like Gecko) Version/4.0 Mobile Safari/534.30 Flipboard/2.2.3/2094,2.2.3.2094,2014-01-29 16:51, -0500, us")
			So(IsPrerender(req), ShouldBeFalse)
		})

		Convey("Twitter bot", func() {
			req, err := http.NewRequest("GET", "http://"+appAdress+"/path/subpath", nil)
			So(err, ShouldBeNil)
			So(req, ShouldNotBeNil)
			req.Header.Add("User-Agent", "Twitterbot/1.0")
			So(IsPrerender(req), ShouldBeTrue)
		})

		Convey("Facebook bot", func() {
			req, err := http.NewRequest("GET", "http://"+appAdress+"/path/subpath", nil)
			So(err, ShouldBeNil)
			So(req, ShouldNotBeNil)
			req.Header.Add("User-Agent", "facebookexternalhit/1.1 (+http://www.facebook.com/externalhit_uatext.php)")
			So(IsPrerender(req), ShouldBeTrue)
		})

		Convey("Pinterest bot", func() {
			req, err := http.NewRequest("GET", "http://"+appAdress+"/path/subpath", nil)
			So(err, ShouldBeNil)
			So(req, ShouldNotBeNil)
			req.Header.Add("User-Agent", "Pinterest/0.2 (+http://www.pinterest.com/)")
			So(IsPrerender(req), ShouldBeTrue)
		})

		Convey("Flipboard bot", func() {
			req, err := http.NewRequest("GET", "http://"+appAdress+"/path/subpath", nil)
			So(err, ShouldBeNil)
			So(req, ShouldNotBeNil)
			req.Header.Add("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10.9; rv:28.0) Gecko/20100101 Firefox/28.0 (FlipboardProxy/1.1; +http://flipboard.com/browserproxy)")
			So(IsPrerender(req), ShouldBeTrue)
		})

		Convey("Generic bot", func() {
			req, err := http.NewRequest("GET", "http://"+appAdress+"/path/subpath", nil)
			So(err, ShouldBeNil)
			So(req, ShouldNotBeNil)
			req.Header.Add("User-Agent", "A string that contain the word bot ....")
			So(IsPrerender(req), ShouldBeTrue)
		})

		Convey("Generic spider", func() {
			req, err := http.NewRequest("GET", "http://"+appAdress+"/path/subpath", nil)
			So(err, ShouldBeNil)
			So(req, ShouldNotBeNil)
			req.Header.Add("User-Agent", "A string that contain the word spider ....")
			So(IsPrerender(req), ShouldBeTrue)
		})

		Convey("Generic crawler", func() {
			req, err := http.NewRequest("GET", "http://"+appAdress+"/path/subpath", nil)
			So(err, ShouldBeNil)
			So(req, ShouldNotBeNil)
			req.Header.Add("User-Agent", "A string that contain the word crawler ....")
			So(IsPrerender(req), ShouldBeTrue)
		})

		Convey("Generic archiver", func() {
			req, err := http.NewRequest("GET", "http://"+appAdress+"/path/subpath", nil)
			So(err, ShouldBeNil)
			So(req, ShouldNotBeNil)
			req.Header.Add("User-Agent", "A string that contain the word archiver ....")
			So(IsPrerender(req), ShouldBeTrue)
		})

		Convey("Static resources with 2 letters extension", func() {
			req, err := http.NewRequest("GET", "http://"+appAdress+"/path/subpath.js", nil)
			So(err, ShouldBeNil)
			So(req, ShouldNotBeNil)
			req.Header.Add("User-Agent", "Any bot that gets filtered by its user-agent.")
			So(IsPrerender(req), ShouldBeFalse)
		})

		Convey("Static resources with 3 letters extension", func() {
			req, err := http.NewRequest("GET", "http://"+appAdress+"/path/subpath.png", nil)
			So(err, ShouldBeNil)
			So(req, ShouldNotBeNil)
			req.Header.Add("User-Agent", "Any bot that gets filtered by its user-agent.")
			So(IsPrerender(req), ShouldBeFalse)
		})

		Convey("Static resources with 4 letters extension", func() {
			req, err := http.NewRequest("GET", "http://"+appAdress+"/path/subpath.html", nil)
			So(err, ShouldBeNil)
			So(req, ShouldNotBeNil)
			req.Header.Add("User-Agent", "Any bot that gets filtered by its user-agent.")
			So(IsPrerender(req), ShouldBeFalse)
		})

		Convey("Static resources with 2 letters extension and a query parameter", func() {
			req, err := http.NewRequest("GET", "http://"+appAdress+"/path/subpath.js?query=something", nil)
			So(err, ShouldBeNil)
			So(req, ShouldNotBeNil)
			req.Header.Add("User-Agent", "Any bot that gets filtered by its user-agent.")
			So(IsPrerender(req), ShouldBeFalse)
		})

		Convey("Static resources with 3 letters extension and a query parameter", func() {
			req, err := http.NewRequest("GET", "http://"+appAdress+"/path/subpath.png?query=something", nil)
			So(err, ShouldBeNil)
			So(req, ShouldNotBeNil)
			req.Header.Add("User-Agent", "Any bot that gets filtered by its user-agent.")
			So(IsPrerender(req), ShouldBeFalse)
		})

		Convey("Static resources with 4 letters extension and a query parameter", func() {
			req, err := http.NewRequest("GET", "http://"+appAdress+"/path/subpath.html?query=something", nil)
			So(err, ShouldBeNil)
			So(req, ShouldNotBeNil)
			req.Header.Add("User-Agent", "Any bot that gets filtered by its user-agent.")
			So(IsPrerender(req), ShouldBeFalse)
		})
	})

	Convey("with mock server", t, func() {
		token := "123"

		Convey("_escaped_fragment_ urls properly proxified", func() {
			Convey("token well added", func(c C) {
				ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					c.Convey("expected request in (mock) server", func() {
						parts := strings.Split(r.URL.Path, "/")
						So(len(parts), ShouldBeGreaterThanOrEqualTo, 2)
						So(parts[1], ShouldEqual, token)
					})
				}))
				defer ts.Close()

				seo4ajaxClient, err := New(Config{
					IP:     serverIP,
					Token:  token,
					Server: ts.URL,
				})
				So(err, ShouldBeNil)
				So(seo4ajaxClient, ShouldNotBeNil)

				req, err := http.NewRequest("GET", "http://"+appAdress+"/path?param1=val1&_escaped_fragment_=", nil)
				So(err, ShouldBeNil)
				So(req, ShouldNotBeNil)
				So(IsPrerender(req), ShouldBeTrue)
				recorder := httptest.NewRecorder()
				seo4ajaxClient.ServeHTTP(recorder, req)
				So(err, ShouldBeNil)
			})
		})

		Convey("header properly sent", func() {
			Convey("headers from origin request", func(c C) {
				ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					c.Convey("expected request in (mock) server", func() {
						So(r.Header.Get("content-type"), ShouldEqual, "content-type")
						So(r.Header.Get("user-agent"), ShouldEqual, "user-agent")
					})
				}))
				defer ts.Close()

				seo4ajaxClient, err := New(Config{
					IP:     serverIP,
					Token:  token,
					Server: ts.URL,
				})
				So(err, ShouldBeNil)
				So(seo4ajaxClient, ShouldNotBeNil)

				req, err := http.NewRequest("GET", "http://"+appAdress+"/path?param1=val1&_escaped_fragment_=", nil)
				req.Header.Add("content-type", "content-type")
				req.Header.Add("user-agent", "user-agent")
				So(err, ShouldBeNil)
				So(req, ShouldNotBeNil)
				So(IsPrerender(req), ShouldBeTrue)
				recorder := httptest.NewRecorder()
				seo4ajaxClient.ServeHTTP(recorder, req)
				So(err, ShouldBeNil)
			})
		})

		Convey("x-forwarded-for added", func(c C) {
			ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				c.Convey("expected request in (mock) server", func() {
					So(r.Header.Get("x-forwarded-for"), ShouldEqual, serverIP)
				})
			}))
			defer ts.Close()

			seo4ajaxClient, err := New(Config{
				IP:     serverIP,
				Token:  token,
				Server: ts.URL,
			})
			So(err, ShouldBeNil)
			So(seo4ajaxClient, ShouldNotBeNil)

			req, err := http.NewRequest("GET", "http://"+appAdress+"/?_escaped_fragment_=", nil)
			So(err, ShouldBeNil)
			So(req, ShouldNotBeNil)
			So(IsPrerender(req), ShouldBeTrue)
			recorder := httptest.NewRecorder()
			seo4ajaxClient.ServeHTTP(recorder, req)
			So(err, ShouldBeNil)
		})

		Convey("x-forwarded-for already present", func(c C) {
			ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				c.Convey("expected request in (mock) server", func() {
					So(r.Header.Get("X-Forwarded-For"), ShouldResemble, serverIP+", 10.0.0.2, 10.0.0.1")
				})
			}))
			defer ts.Close()

			seo4ajaxClient, err := New(Config{
				IP:     serverIP,
				Token:  token,
				Server: ts.URL,
			})
			So(err, ShouldBeNil)
			So(seo4ajaxClient, ShouldNotBeNil)

			req, err := http.NewRequest("GET", "http://"+appAdress+"/?_escaped_fragment_=", nil)
			req.Header.Add("x-forwarded-for", "10.0.0.2, 10.0.0.1")
			So(err, ShouldBeNil)
			So(req, ShouldNotBeNil)
			So(IsPrerender(req), ShouldBeTrue)
			recorder := httptest.NewRecorder()
			seo4ajaxClient.ServeHTTP(recorder, req)
			So(err, ShouldBeNil)
		})
	})

	Convey("not follow redirect", t, func() {
		token := "123"

		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			http.Redirect(w, r, "http://example.com/", 302)
		}))
		defer ts.Close()

		seo4ajaxClient, err := New(Config{
			IP:     serverIP,
			Token:  token,
			Server: ts.URL,
		})
		So(err, ShouldBeNil)
		So(seo4ajaxClient, ShouldNotBeNil)

		req, err := http.NewRequest("GET", "http://"+appAdress+"/?_escaped_fragment_=", nil)
		So(err, ShouldBeNil)
		So(req, ShouldNotBeNil)
		So(IsPrerender(req), ShouldBeTrue)

		recorder := httptest.NewRecorder()
		seo4ajaxClient.ServeHTTP(recorder, req)

		So(err, ShouldBeNil)
		So(recorder.Header().Get("Location"), ShouldEqual, "http://example.com/")
		So(recorder.Code, ShouldEqual, 302)
	})

	Convey("returns error if no token", t, func() {
		seo4ajaxClient, err := New(Config{
			IP: serverIP,
		})
		So(seo4ajaxClient, ShouldBeNil)
		So(err, ShouldNotBeNil)
		So(err, ShouldEqual, ErrNoToken)
	})
}
