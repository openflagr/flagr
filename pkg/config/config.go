package config

import (
	"fmt"
	"os"

	"github.com/DataDog/datadog-go/statsd"
	"github.com/caarlos0/env"
	"github.com/evalphobia/logrus_sentry"
	raven "github.com/getsentry/raven-go"
	newrelic "github.com/newrelic/go-agent"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/sirupsen/logrus"
)

// EvalOnlyModeDBDrivers is a list of DBDrivers that we should only run in EvalOnlyMode.
var EvalOnlyModeDBDrivers = map[string]struct{}{
	"json_file": {},
	"json_http": {},
}

// Global is the global dependency we can use, such as the new relic app instance
var Global = struct {
	NewrelicApp  newrelic.Application
	StatsdClient *statsd.Client
	Prometheus   prometheusMetrics
}{}

func init() {
	env.Parse(&Config)

	setupEvalOnlyMode()
	setupSentry()
	setupLogrus()
	setupStatsd()
	setupNewrelic()
	setupPrometheus()
}

func setupEvalOnlyMode() {
	if _, ok := EvalOnlyModeDBDrivers[Config.DBDriver]; ok {
		Config.EvalOnlyMode = true
	}
}

func setupLogrus() {
	l, err := logrus.ParseLevel(Config.LogrusLevel)
	if err != nil {
		logrus.WithField("err", err).Fatalf("failed to set logrus level:%s", Config.LogrusLevel)
	}
	logrus.SetLevel(l)
	logrus.SetOutput(os.Stdout)
	switch Config.LogrusFormat {
	case "text":
		logrus.SetFormatter(&logrus.TextFormatter{})
	case "json":
		logrus.SetFormatter(&logrus.JSONFormatter{})
	default:
		logrus.Warnf("unexpected logrus format: %s, should be one of: text, json", Config.LogrusFormat)
	}
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
		// These two cannot be enabled at the same time and cross application is enabled by default
		nCfg.DistributedTracer.Enabled = Config.NewRelicDistributedTracingEnabled
		nCfg.CrossApplicationTracer.Enabled = !Config.NewRelicDistributedTracingEnabled
		app, err := newrelic.NewApplication(nCfg)
		if err != nil {
			panic(fmt.Sprintf("unable to initialize newrelic. %s", err))
		}
		Global.NewrelicApp = app
	}
}

type prometheusMetrics struct {
	ScrapePath       string
	EvalCounter      *prometheus.CounterVec
	RequestCounter   *prometheus.CounterVec
	RequestHistogram *prometheus.HistogramVec
}

func setupPrometheus() {
	if Config.PrometheusEnabled {
		Global.Prometheus.ScrapePath = Config.PrometheusPath
		Global.Prometheus.EvalCounter = promauto.NewCounterVec(prometheus.CounterOpts{
			Name: "flagr_eval_results",
			Help: "A counter of eval results",
		}, []string{"EntityType", "FlagID", "VariantID", "VariantKey"})
		Global.Prometheus.RequestCounter = promauto.NewCounterVec(prometheus.CounterOpts{
			Name: "flagr_requests_total",
			Help: "The total http requests received",
		}, []string{"status", "path", "method"})

		if Config.PrometheusIncludeLatencyHistogram {
			Global.Prometheus.RequestHistogram = promauto.NewHistogramVec(prometheus.HistogramOpts{
				Name: "flagr_requests_buckets",
				Help: "A histogram of latencies for requests received",
			}, []string{"status", "path", "method"})
		}
	}
}
