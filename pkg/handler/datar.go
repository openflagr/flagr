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
	if !hasDatar(config.Config.RecorderType) {

		return nil
	}
	singletonEngine = datar.New(
		getDB(),
		true,
		config.Config.RecorderDatarFlushInterval,
	)
	return singletonEngine
}

// hasDatar returns true if the slice contains "datar".
func hasDatar(types []string) bool {
	for _, t := range types {
		if t == "datar" {
			return true
		}
	}
	return false
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

