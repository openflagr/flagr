
# pprof

 Wrapper middleware for `net/http/pprof`.

 View the [docs](http://godoc.org/github.com/gohttp/pprof).

```go
app := app.New()
srv := &http.Server{Addr: ":3000", Handler: app}
app.Use(pprof.New())
srv.ListenAndServe()
```

# License

 MIT