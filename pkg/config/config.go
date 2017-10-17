package config

import (
	"time"

	"github.com/jinzhu/configor"
)

// Config is the whole configuration of the app
var Config = struct {
	DB struct {
		DBDriver        string `env:"FLAGR_DB_DBDRIVER" default:"mysql"`
		DBConnectionStr string `env:"FLAGR_DB_DBCONNECTIONSTR" default:"root:@tcp(127.0.0.1:18100)/flagr?parseTime=true"`
	}
	CORS struct {
		Enabled bool `env:"FLAGR_CORS_ENABLED" default:"true"`
	}
	Sentry struct {
		Enabled bool   `env:"FLAGR_SENTRY_ENABLED" default:"false"`
		DSN     string `env:"FLAGR_SENTRY_DSN" default:""`
	}
	EvalCache struct {
		RefreshTimeout  Duration `env:"FLAGR_EVALCACHE_REFRESHTIMEOUT" default:"59s"`
		RefreshInterval Duration `env:"FLAGR_EVALCACHE_REFRESHINTERVAL" default:"10s"`
	}
	PProf struct {
		Enabled bool `env:"FLAGR_PPROF_ENABLED" default:"true"`
	}
	Recorder struct {
		Enabled bool   `env:"FLAGR_RECORDER_ENABLED" default:"true"`
		Type    string `env:"FLAGR_RECORDER_TYPE" default:"kafka"`
		Kafka   struct {
			Brokers   string `env:"FLAGR_RECORDER_KAFKA_BROKERS" default:":9092"`
			CertFile  string `env:"FLAGR_RECORDER_KAFKA_CERTFILE" default:""`
			KeyFile   string `env:"FLAGR_RECORDER_KAFKA_KEYFILE" default:""`
			CAFile    string `env:"FLAGR_RECORDER_KAFKA_CAFILE" default:""`
			VerifySSL bool   `env:"FLAGR_RECORDER_KAFKA_VERIFYSSL" default:"false"`
			Verbose   bool   `env:"FLAGR_RECORDER_KAFKA_VERBOSE" default:"true"`

			Topic          string   `env:"FLAGR_RECORDER_KAFKA_TOPIC" default:"flagr-records"`
			RetryMax       int      `env:"FLAGR_RECORDER_KAFKA_RETRYMAX" default:"5"`
			FlushFrequency Duration `env:"FLAGR_RECORDER_KAFKA_FLUSHFREQUENCY" default:"500ms"`
		}
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
