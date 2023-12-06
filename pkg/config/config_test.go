package config

import (
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
)

func TestSetupSentry(t *testing.T) {
	Config.SentryEnabled = true
	Config.SentryEnvironment = "test"
	defer func() {
		Config.SentryEnabled = false
		Config.SentryEnvironment = ""
	}()

	assert.NotPanics(t, func() { setupSentry() })
}

func TestSetupNewRelic(t *testing.T) {
	Config.NewRelicEnabled = true
	defer func() {
		Config.NewRelicEnabled = false
	}()

	assert.Panics(t, func() { setupNewrelic() })
}

func TestSetupStatsd(t *testing.T) {
	Config.StatsdEnabled = true
	defer func() {
		Config.StatsdEnabled = false
	}()

	assert.NotPanics(t, func() { setupStatsd() })
}

func TestSetupPrometheus(t *testing.T) {
	prometheus.DefaultRegisterer = prometheus.NewRegistry()
	setupPrometheus()
	assert.Nil(t, Global.Prometheus.EvalCounter)

	Config.PrometheusEnabled = true
	defer func() { Config.PrometheusEnabled = false }()
	setupPrometheus()
	assert.NotNil(t, Global.Prometheus.EvalCounter)
	assert.NotNil(t, Global.Prometheus.RequestCounter)
	assert.Nil(t, Global.Prometheus.RequestHistogram)
}

func TestSetupPrometheusWithLatencies(t *testing.T) {
	prometheus.DefaultRegisterer = prometheus.NewRegistry()
	Config.PrometheusEnabled = true
	Config.PrometheusIncludeLatencyHistogram = true
	defer func() {
		Config.PrometheusEnabled = false
		Config.PrometheusIncludeLatencyHistogram = false
	}()

	setupPrometheus()
	assert.NotNil(t, Global.Prometheus.EvalCounter)
	assert.NotNil(t, Global.Prometheus.RequestCounter)
	assert.NotNil(t, Global.Prometheus.RequestHistogram)
}
