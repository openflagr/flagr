package config

import (
	"net/http"
	"os"

	"github.com/caarlos0/env"
	"github.com/evalphobia/logrus_sentry"
	raven "github.com/getsentry/raven-go"
	"github.com/gohttp/pprof"
	negronilogrus "github.com/meatballhat/negroni-logrus"
	newrelic "github.com/newrelic/go-agent"
	"github.com/rs/cors"
	"github.com/sirupsen/logrus"
	"github.com/urfave/negroni"
	negroninewrelic "github.com/yadvendar/negroni-newrelic-go-agent"
)

// Global is the global dependency we can use, such as the new relic app instance
var Global = struct {
	NewrelicApp newrelic.Application
}{}

func init() {
	env.Parse(&Config)

	setupSentry()
	setupLogrus()
	setupNewrelic()
}

func setupLogrus() {
	logrus.SetOutput(os.Stdout)
	logrus.SetLevel(logrus.InfoLevel)
}

func setupSentry() {
	if Config.SentryEnabled {
		raven.SetDSN(Config.SentryDSN)
		hook, err := logrus_sentry.NewSentryHook(Config.SentryDSN, []logrus.Level{
			logrus.PanicLevel,
			logrus.FatalLevel,
			logrus.ErrorLevel,
		})
		if err != nil {
			logrus.WithField("err", err).Error("failed to hook logurs to sentry")
			return
		}
		logrus.StandardLogger().Hooks.Add(hook)
	}
}

func setupNewrelic() {
	if Config.NewRelicEnabled {
		nCfg := newrelic.NewConfig(Config.NewRelicAppName, Config.NewRelicKey)
		nCfg.Enabled = true
		app, err := newrelic.NewApplication(nCfg)
		if err != nil {
			logrus.Fatalf("unable to initialize newrelic. %s", err)
		}
		Global.NewrelicApp = app
	}
}

// SetupGlobalMiddleware setup the global middleware
func SetupGlobalMiddleware(handler http.Handler) http.Handler {
	pwd, _ := os.Getwd()
	n := negroni.New()

	if Config.CORSEnabled {
		c := cors.New(cors.Options{
			AllowedOrigins: []string{"*"},
			AllowedHeaders: []string{"Content-Type", "Accepts"},
			AllowedMethods: []string{"GET", "POST", "PUT", "DELETE"},
		})
		n.Use(c)
	}

	if Config.NewRelicEnabled {
		newRelicMiddleware := &negroninewrelic.Newrelic{Application: &Global.NewrelicApp}
		n.Use(newRelicMiddleware)
	}

	n.Use(negronilogrus.NewMiddlewareFromLogger(logrus.StandardLogger(), "flagr"))
	n.Use(negroni.NewRecovery())
	n.Use(negroni.NewStatic(http.Dir(pwd + "/browser/flagr-ui/dist/")))

	if Config.PProfEnabled {
		n.UseHandler(pprof.New()(handler))
	} else {
		n.UseHandler(handler)
	}

	return n
}
