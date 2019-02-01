package sql

type registerConfig struct{ serviceName string }

// RegisterOption represents an option that can be passed to Register.
type RegisterOption func(*registerConfig)

func defaults(cfg *registerConfig) {
	// default cfg.serviceName set in Register based on driver name
}

// WithServiceName sets the given service name for the registered driver.
func WithServiceName(name string) RegisterOption {
	return func(cfg *registerConfig) {
		cfg.serviceName = name
	}
}
