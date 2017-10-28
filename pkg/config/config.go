package config

import (
	"encoding/gob"
	"os"
	"time"

	"github.com/caarlos0/env"
	raven "github.com/getsentry/raven-go"
	"github.com/sirupsen/logrus"
	"github.com/zhouzhuojie/conditions"
)

// Config is the whole configuration of the app
var Config = struct {
	Host                        string        `env:"HOST" envDefault:"127.0.0.1"`
	PProfEnabled                bool          `env:"FLAGR_PPROF_ENABLED" envDefault:"true"`
	DBDriver                    string        `env:"FLAGR_DB_DBDRIVER" envDefault:"mysql"`
	DBConnectionStr             string        `env:"FLAGR_DB_DBCONNECTIONSTR" envDefault:"root:@tcp(127.0.0.1:18100)/flagr?parseTime=true"`
	CORSEnabled                 bool          `env:"FLAGR_CORS_ENABLED" envDefault:"true"`
	SentryEnabled               bool          `env:"FLAGR_SENTRY_ENABLED" envDefault:"false"`
	SentryDSN                   string        `env:"FLAGR_SENTRY_DSN" envDefault:""`
	NewRelicEnabled             bool          `env:"FLAGR_NEWRELIC_ENABLED" envDefault:"false"`
	NewRelicAppName             string        `env:"FLAGR_NEWRELIC_NAME" envDefault:"flagr"`
	NewRelicKey                 string        `env:"FLAGR_NEWRELIC_KEY" envDefault:""`
	EvalCacheRefreshTimeout     time.Duration `env:"FLAGR_EVALCACHE_REFRESHTIMEOUT" envDefault:"59s"`
	EvalCacheRefreshInterval    time.Duration `env:"FLAGR_EVALCACHE_REFRESHINTERVAL" envDefault:"3s"`
	RecorderEnabled             bool          `env:"FLAGR_RECORDER_ENABLED" envDefault:"false"`
	RecorderType                string        `env:"FLAGR_RECORDER_TYPE" envDefault:"kafka"`
	RecorderKafkaBrokers        string        `env:"FLAGR_RECORDER_KAFKA_BROKERS" envDefault:":9092"`
	RecorderKafkaCertFile       string        `env:"FLAGR_RECORDER_KAFKA_CERTFILE" envDefault:""`
	RecorderKafkaKeyFile        string        `env:"FLAGR_RECORDER_KAFKA_KEYFILE" envDefault:""`
	RecorderKafkaCAFile         string        `env:"FLAGR_RECORDER_KAFKA_CAFILE" envDefault:""`
	RecorderKafkaVerifySSL      bool          `env:"FLAGR_RECORDER_KAFKA_VERIFYSSL" envDefault:"false"`
	RecorderKafkaVerbose        bool          `env:"FLAGR_RECORDER_KAFKA_VERBOSE" envDefault:"true"`
	RecorderKafkaTopic          string        `env:"FLAGR_RECORDER_KAFKA_TOPIC" envDefault:"flagr-records"`
	RecorderKafkaRetryMax       int           `env:"FLAGR_RECORDER_KAFKA_RETRYMAX" envDefault:"5"`
	RecorderKafkaFlushFrequency time.Duration `env:"FLAGR_RECORDER_KAFKA_FLUSHFREQUENCY" envDefault:"500ms"`
	RecorderKafkaEncrypted      bool          `env:"FLAGR_RECORDER_KAFKA_ENCRYPTED" envDefault:"false"`
	RecorderKafkaEncryptionKey  string        `env:"FLAGR_RECORDER_KAFKA_ENCRYPTION_KEY" envDefault:""`
}{}

func init() {
	env.Parse(&Config)

	setupRaven()
	setupGob()
	setupLogrus()
}

func setupLogrus() {
	logrus.SetOutput(os.Stdout)
	logrus.SetLevel(logrus.InfoLevel)
}

func setupGob() {
	gob.Register(conditions.BinaryExpr{})
	gob.Register(conditions.VarRef{})
}

func setupRaven() {
	if Config.SentryEnabled {
		raven.SetDSN(Config.SentryDSN)
	}
}
