package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/codegangsta/negroni"
	newrelic "github.com/yadvendar/negroni-newrelic-go-agent"
)

func main() {
	r := http.NewServeMux()
	r.HandleFunc(`/`, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "success!\n")
	})

	n := negroni.New()
	config := newrelic.NewConfig("APP_SERVER_NAME", "NEWRELIC_LICENSE_KEY")
	config.BetaToken = "BETA_TOKEN" // this is valid only till go-agent is in beta
	config.Enabled = true
	newRelicMiddleware, err := newrelic.New(config)
	if err != nil {
		fmt.Println("Unable to initialize newrelic. Error=" + err.Error())
		os.Exit(0)
	}
	n.Use(newRelicMiddleware)
	n.UseHandler(r)

	n.Run(":3000")
}
