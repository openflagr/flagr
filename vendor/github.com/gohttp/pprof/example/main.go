package main

import "github.com/gohttp/pprof"
import "github.com/gohttp/app"
import "net/http"

func main() {
	app := app.New()
	srv := &http.Server{Addr: ":3000", Handler: app}
	app.Use(pprof.New())
	app.Get("/", respond("hai\n"))
	println("listening on 0.0.0.0:3000")
	srv.ListenAndServe()
}

func respond(s string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(s))
	}
}
