// Package nrlogxi forwards go-agent log messages to mgutz/logxi.  If you would
// like to use mgutz/logxi for go-agent log messages, wrap your logxi Logger
// using nrlogxi.New to create a newrelic.Logger.
//
//	l := log.New("newrelic")
//	l.SetLevel(log.LevelInfo)
//	cfg.Logger = nrlogxi.New(l)
//
package nrlogxi

import (
	"github.com/mgutz/logxi/v1"
	newrelic "github.com/newrelic/go-agent"
	"github.com/newrelic/go-agent/internal"
)

func init() { internal.TrackUsage("integration", "logging", "logxi", "v1") }

type shim struct {
	e log.Logger
}

func (l *shim) Error(msg string, context map[string]interface{}) {
	l.e.Error(msg, convert(context)...)
}
func (l *shim) Warn(msg string, context map[string]interface{}) {
	l.e.Warn(msg, convert(context)...)
}
func (l *shim) Info(msg string, context map[string]interface{}) {
	l.e.Info(msg, convert(context)...)
}
func (l *shim) Debug(msg string, context map[string]interface{}) {
	l.e.Debug(msg, convert(context)...)
}
func (l *shim) DebugEnabled() bool {
	return l.e.IsDebug()
}

func convert(c map[string]interface{}) []interface{} {
	output := make([]interface{}, 0, 2*len(c))
	for k, v := range c {
		output = append(output, k, v)
	}
	return output
}

// New returns a newrelic.Logger which forwards agent log messages to the
// provided logxi Logger.
func New(l log.Logger) newrelic.Logger {
	return &shim{
		e: l,
	}
}
