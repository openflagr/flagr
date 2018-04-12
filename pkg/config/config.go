package config

import (
	"fmt"
	"os"

	"github.com/DataDog/datadog-go/statsd"
	"github.com/caarlos0/env"
	"github.com/evalphobia/logrus_sentry"
	raven "github.com/getsentry/raven-go"
	newrelic "github.com/newrelic/go-agent"
	"github.com/sirupsen/logrus"
)

// Global is the global dependency we can use, such as the new relic app instance
var Global = struct {
	NewrelicApp  newrelic.Application
	StatsdClient *statsd.Client
}{}

func init() {
	env.Parse(&Config)

	setupSentry()
	setupLogrus()
	setupStatsd()
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

func setupStatsd() {
	if Config.StatsdEnabled {
		client, err := statsd.New(fmt.Sprintf("%s:%s", Config.StatsdHost, Config.StatsdPort))
		if err != nil {
			panic(fmt.Sprintf("unable to initialize statsd. %s", err))
		}
		client.Namespace = Config.StatsdPrefix

		Global.StatsdClient = client
	}
}

func setupNewrelic() {
	if Config.NewRelicEnabled {
		nCfg := newrelic.NewConfig(Config.NewRelicAppName, Config.NewRelicKey)
		nCfg.Enabled = true
		app, err := newrelic.NewApplication(nCfg)
		if err != nil {
			panic(fmt.Sprintf("unable to initialize newrelic. %s", err))
		}
		Global.NewrelicApp = app
	}
}
