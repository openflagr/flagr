package negroninewrelic

import (
	"net/http"

	"github.com/newrelic/go-agent"
)

type Newrelic struct {
	Application *newrelic.Application
	Transaction *newrelic.Transaction
}

func NewConfig(applicationName string, licenseKey string) newrelic.Config {
	return newrelic.NewConfig(applicationName, licenseKey)
}
func New(config newrelic.Config) (*Newrelic, error) {
	app, err := newrelic.NewApplication(config)
	return &Newrelic{Application: &app}, err
}

func (n *Newrelic) ServeHTTP(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc) {

	txn := ((*n.Application).StartTransaction(r.URL.Path, rw, r)).(newrelic.Transaction)
	n.Transaction = &txn
	defer (*n.Transaction).End()

	// Use if required
	//	(*n.Transaction).AddAttribute("query", r.URL.RawQuery)

	next(rw, r)
}
