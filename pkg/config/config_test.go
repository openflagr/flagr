package config

import (
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
)

func TestSetupSentry(t *testing.T) {
	Config.SentryEnabled = true
	assert.NotPanics(t, func() {
		setupSentry()
	})
	Config.SentryEnabled = false
}

func TestSetupNewRelic(t *testing.T) {
	Config.NewRelicEnabled = true
	assert.Panics(t, func() {
		setupNewrelic()
	})
	Config.NewRelicEnabled = false
}

func TestSetupStatsd(t *testing.T) {
	Config.StatsdEnabled = true
	assert.NotPanics(t, func() {
		setupStatsd()
	})
	Config.StatsdEnabled = false
}

func TestSetupPrometheus(t *testing.T) {
	prometheus.DefaultRegisterer = prometheus.NewRegistry()
	Config.PrometheusEnabled = false
	setupPrometheus()
	assert.Nil(t, Global.Prometheus.EvalCounter)
	Config.PrometheusEnabled = true
	setupPrometheus()
	assert.NotNil(t, Global.Prometheus.EvalCounter)
	assert.NotNil(t, Global.Prometheus.RequestCounter)
	assert.Nil(t, Global.Prometheus.RequestHistogram)
	Config.PrometheusEnabled = false
}

func TestSetupPrometheusWithLatencies(t *testing.T) {
	prometheus.DefaultRegisterer = prometheus.NewRegistry()
	Config.PrometheusEnabled = true
	Config.PrometheusIncludeLatencyHistogram = true
	setupPrometheus()
	assert.NotNil(t, Global.Prometheus.EvalCounter)
	assert.NotNil(t, Global.Prometheus.RequestCounter)
	assert.NotNil(t, Global.Prometheus.RequestHistogram)
	Config.PrometheusEnabled = false
}
