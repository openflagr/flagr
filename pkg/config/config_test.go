package config

import (
	"testing"

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
