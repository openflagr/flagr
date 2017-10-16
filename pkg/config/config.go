package config

import (
	"time"

	"github.com/jinzhu/configor"
)

// Config is the whole configuration of the app
var Config = struct {
	DB struct {
		DBDriver        string `env:"DBDriver" default:"mysql"`
		DBConnectionStr string `env:"DBConnectionStr" default:"root:@tcp(127.0.0.1:18100)/flagr?parseTime=true"`
	}
	CORS struct {
		Enabled bool `env:"CORSEnabled" default:"true"`
	}
	Sentry struct {
		Enabled bool   `env:"SentryEnabled" default:"false"`
		DSN     string `env:"SentryDSN" default:""`
	}
	EvalCache struct {
		RefreshTimeout  Duration `env:"EvalCacheRefreshTimeout" default:"59s"`
		RefreshInterval Duration `env:"EvalCacheRefreshInterval" default:"10s"`
	}
	PProf struct {
		Enabled bool `env:"PProfEnabled" default:"true"`
	}
}{}

func init() {
	configor.Load(&Config)
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
