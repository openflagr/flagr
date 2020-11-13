package config

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
)

func TestSetupConfig(t *testing.T) {
	nonsense := "asdf;lkj"
	orig := os.Getenv("HOST")
	err := os.Setenv("HOST", nonsense)
	assert.Nil(t, err)

	setupConfig()
	assert.Equal(t, nonsense, Config.Host)
	err = os.Setenv("HOST", orig)
	assert.Nil(t, err)
}

func TestParseServerEntityContextConfig(t *testing.T) {
	serverEntityContext := map[string]string{
		"foo": "bar",
	}
	serverEntityContextBytes, err := json.Marshal(serverEntityContext)
	assert.Nil(t, err)
	serverEntityContextJson := string(serverEntityContextBytes)

	orig := os.Getenv("FLAGR_EVAL_SERVER_ENTITY_CONTEXT")
	err = os.Setenv("FLAGR_EVAL_SERVER_ENTITY_CONTEXT", serverEntityContextJson)
	assert.Nil(t, err)

	setupConfig()
	assert.Equal(t, serverEntityContext, Config.EvalServerEntityContext)
	assert.Equal(t, "bar", Config.EvalServerEntityContext["foo"])
	err = os.Setenv("FLAGR_EVAL_SERVER_ENTITY_CONTEXT", orig)
	assert.Nil(t, err)
}

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
