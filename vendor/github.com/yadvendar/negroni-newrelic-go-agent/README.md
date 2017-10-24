# negroni-newrelic-go-agent
[New Relic Go Agent](https://github.com/newrelic/go-agent) middleware for [negroni](https://github.com/codegangsta/negroni)

[New Relic](https://newrelic.com) has recently released a Go Agent in their APM module. Its currently in beta and in order to get started you can request for a beta token by filling [the beta agreement form](http://goo.gl/forms/Rcv1b10Qvt1ENLlr1).
 
If you have microservice architecture using negroni and are looking for a middleware to attach new relic go agent to your existing stack of middlewares, then you can use [negroni-newrelic-go-agent](https://github.com/yadvendar/negroni-newrelic-go-agent) to achieve the same.

Usage
-----

```go
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

```

See a running [example](https://github.com/yadvendar/negroni-newrelic-go-agent/blob/master/example/example.go).

Credits
-------

[New Relic Go Agent](https://github.com/newrelic/go-agent)

License
-------

See [LICENSE.txt](https://github.com/newrelic/go-agent/blob/master/LICENSE.txt)
