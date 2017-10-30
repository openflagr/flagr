package config

import (
	"encoding/gob"
	"net/http"
	"os"
	"time"

	"github.com/caarlos0/env"
	"github.com/evalphobia/logrus_sentry"
	raven "github.com/getsentry/raven-go"
	"github.com/gohttp/pprof"
	negronilogrus "github.com/meatballhat/negroni-logrus"
	newrelic "github.com/newrelic/go-agent"
	"github.com/rs/cors"
	"github.com/sirupsen/logrus"
	"github.com/urfave/negroni"
	negroninewrelic "github.com/yadvendar/negroni-newrelic-go-agent"
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

// Global is the global dependency we can use, such as the new relic app instance
var Global = struct {
	NewRelicApp newrelic.Application
}{}

func init() {
	env.Parse(&Config)

	setupSentry()
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

func setupSentry() {
	if Config.SentryEnabled {
		raven.SetDSN(Config.SentryDSN)
		hook, err := logrus_sentry.NewSentryHook(Config.SentryDSN, []logrus.Level{
			logrus.PanicLevel,
			logrus.FatalLevel,
			logrus.ErrorLevel,
		})
		if err != nil {
			logrus.WithField("err", err).Error("failed to hook logurs to sentry")
			return
		}
		logrus.StandardLogger().Hooks.Add(hook)
	}
}

// SetupGlobalMiddleware setup the global middleware
func SetupGlobalMiddleware(handler http.Handler) http.Handler {
	pwd, _ := os.Getwd()
	n := negroni.New()

	if Config.CORSEnabled {
		c := cors.New(cors.Options{
			AllowedOrigins: []string{"*"},
			AllowedHeaders: []string{"Content-Type", "Accepts"},
			AllowedMethods: []string{"GET", "POST", "PUT", "DELETE"},
		})
		n.Use(c)
	}

	if Config.NewRelicEnabled {
		nCfg := newrelic.NewConfig(Config.NewRelicAppName, Config.NewRelicKey)
		nCfg.Enabled = true
		newRelicMiddleware, err := negroninewrelic.New(nCfg)
		Global.NewRelicApp = *newRelicMiddleware.Application
		if err != nil {
			logrus.Fatalf("unable to initialize newrelic. %s", err)
		}
		n.Use(newRelicMiddleware)
	}

	n.Use(negronilogrus.NewMiddlewareFromLogger(logrus.StandardLogger(), "flagr"))
	n.Use(negroni.NewRecovery())
	n.Use(negroni.NewStatic(http.Dir(pwd + "/browser/flagr-ui/dist/")))

	if Config.PProfEnabled {
		n.UseHandler(pprof.New()(handler))
	} else {
		n.UseHandler(handler)
	}

	return n
}
