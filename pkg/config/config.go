package config

import (
	"os"
	"time"

	"github.com/jinzhu/configor"
)

// Config is the whole configuration of the app
var Config = struct {
	DB struct {
		DBDriver        string `env:"DBDriver"`
		DBConnectionStr string `env:"DBConnectionStr"`
	}
	CORS struct {
		Enabled bool `env:"CORSEnabled"`
	}
	Sentry struct {
		Enabled bool   `env:"SentryEnabled"`
		DSN     string `env:"SentryDSN"`
	}
	EvalCache struct {
		RefreshTimeout  Duration `env:"EvalCacheRefreshTimeout"`
		RefreshInterval Duration `env:"EvalCacheRefreshInterval"`
	}
	PProf struct {
		Enabled bool `env:"PProfEnabled"`
	}
}{}

func init() {
	pwd, _ := os.Getwd()
	configor.Load(&Config, pwd+"/pkg/config/config.toml")
}

// Duration is an alias type of time.Duration
type Duration struct {
	time.Duration
}

// UnmarshalText implements the encoding.TextUnmarshaler interface
func (d *Duration) UnmarshalText(text []byte) error {
	var err error
	d.Duration, err = time.ParseDuration(string(text))
	return err
}
