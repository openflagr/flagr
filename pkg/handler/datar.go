package handler

import (
	"sync"

	"github.com/openflagr/flagr/pkg/config"
	"github.com/openflagr/flagr/pkg/datar"

)

var (
	singletonEngine   *datar.Engine
	singletonEngineMu sync.Mutex
)

// GetDatar returns the singleton datar.Engine.
// Creates the instance on first call, starting its flush loop.
// Returns nil if Datar is not enabled — safe to call methods on nil.
func GetDatar() *datar.Engine {
	singletonEngineMu.Lock()
	defer singletonEngineMu.Unlock()
	if singletonEngine != nil {
		return singletonEngine
	}
	if !config.Config.DatarEnabled {
		return nil
	}
	singletonEngine = datar.New(
		getDB(),
		config.Config.DatarEnabled,
		config.Config.DatarFlushInterval,
	)
	return singletonEngine
}

// ResetDatar clears the singleton for test isolation.
func ResetDatar() {
	singletonEngineMu.Lock()
	defer singletonEngineMu.Unlock()
	if singletonEngine != nil {
		singletonEngine.Shutdown()
		singletonEngine = nil
	}
}

