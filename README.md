# go-seo4ajax
Golang connector for Seo4Ajax [GoDoc](http://godoc.org/github.com/justwatchcom/go-seo4ajax).

Code and tests is highly adapted from [connect middleware](https://www.npmjs.com/package/connect-s4a).

```
http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
    if seo4ajax.IsPrerender(r) {
        err := seo4ajax.GetPrerenderedPage(w, r)
        if err != nil {
            log.Print(err.Error())
        }
    } else {
        fmt.Fprintf(w, "Hello, %q", html.EscapeString(r.URL.Path))
    }
})
```
