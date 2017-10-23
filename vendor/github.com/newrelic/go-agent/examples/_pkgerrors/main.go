package main

import (
	"fmt"
	"os"
	"time"

	newrelic "github.com/newrelic/go-agent"
	"github.com/newrelic/go-agent/_integrations/nrpkgerrors"
	"github.com/pkg/errors"
)

type sampleError string

func (e sampleError) Error() string {
	return string(e)
}

func alpha() error {
	return errors.WithStack(sampleError("alpha is the cause"))
}

func beta() error {
	return errors.WithStack(alpha())
}

func gamma() error {
	return errors.Wrap(beta(), "gamma was involved")
}

func mustGetEnv(key string) string {
	if val := os.Getenv(key); "" != val {
		return val
	}
	panic(fmt.Sprintf("environment variable %s unset", key))
}

func main() {
	cfg := newrelic.NewConfig("pkg/errors app", mustGetEnv("NEW_RELIC_LICENSE_KEY"))
	cfg.Logger = newrelic.NewDebugLogger(os.Stdout)
	app, err := newrelic.NewApplication(cfg)
	if nil != err {
		fmt.Println(err)
		os.Exit(1)
	}

	if err := app.WaitForConnection(5 * time.Second); nil != err {
		fmt.Println(err)
	}

	txn := app.StartTransaction("has-error", nil, nil)
	e := gamma()
	txn.NoticeError(nrpkgerrors.Wrap(e))
	txn.End()

	app.Shutdown(10 * time.Second)
}
