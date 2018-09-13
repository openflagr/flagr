package config

import "time"

// Config is the whole configuration of the app
var Config = struct {
	// Host - Flagr server host
	Host string `env:"HOST" envDefault:"localhost"`
	// Port - Flagr server port
	Port int `env:"PORT" envDefault:"18000"`

	// LogrusLevel sets the logrus logging level
	LogrusLevel string `env:"FLAGR_LOGRUS_LEVEL" envDefault:"info"`
	// PProfEnabled - to enable the standard pprof of golang's http server
	PProfEnabled bool `env:"FLAGR_PPROF_ENABLED" envDefault:"true"`

	// MiddlewareVerboseLoggerEnabled - to enable the negroni-logrus logger for all the endpoints
	// useful for debugging
	MiddlewareVerboseLoggerEnabled bool `env:"FLAGR_MIDDLEWARE_VERBOSE_LOGGER_ENABLED" envDefault:"true"`

	// EvalLoggingEnabled - to enable the logging for eval results
	EvalLoggingEnabled bool `env:"FLAGR_EVAL_LOGGING_ENABLED" envDefault:"true"`

	// RateLimiterPerFlagPerSecondConsoleLogging - to rate limit the logging rate
	// per flag per second
	RateLimiterPerFlagPerSecondConsoleLogging int `env:"FLAGR_RATELIMITER_PERFLAG_PERSECOND_CONSOLE_LOGGING" envDefault:"100"`

	// EvalCacheRefreshTimeout - timeout of getting the flags data from DB into the in-memory evaluation cache
	EvalCacheRefreshTimeout time.Duration `env:"FLAGR_EVALCACHE_REFRESHTIMEOUT" envDefault:"59s"`
	// EvalCacheRefreshInterval - time interval of getting the flags data from DB into the in-memory evaluation cache
	EvalCacheRefreshInterval time.Duration `env:"FLAGR_EVALCACHE_REFRESHINTERVAL" envDefault:"3s"`

	// DBDriver - Flagr supports sqlite3, mysql, postgres
	DBDriver string `env:"FLAGR_DB_DBDRIVER" envDefault:"sqlite3"`
	// DBConnectionStr - examples
	// sqlite3:  "/tmp/file.db"
	// sqlite3:  ":memory:"
	// mysql:    "root:@tcp(127.0.0.1:18100)/flagr?parseTime=true"
	// postgres: "host=myhost user=root dbname=flagr password=mypassword"
	DBConnectionStr string `env:"FLAGR_DB_DBCONNECTIONSTR" envDefault:"flagr.sqlite"`
	// DBConnectionDebug controls whether to show the database connection debugging logs
	// warning: it may log the credentials to the stdout
	DBConnectionDebug bool `env:"FLAGR_DB_DBCONNECTION_DEBUG" envDefault:"true"`

	// CORSEnabled - enable CORS
	CORSEnabled bool `env:"FLAGR_CORS_ENABLED" envDefault:"true"`

	// SentryEnabled - enable Sentry and Sentry DSN
	SentryEnabled bool   `env:"FLAGR_SENTRY_ENABLED" envDefault:"false"`
	SentryDSN     string `env:"FLAGR_SENTRY_DSN" envDefault:""`

	// NewRelicEnabled - enable the NewRelic monitoring for all the endpoints and DB operations
	NewRelicEnabled bool   `env:"FLAGR_NEWRELIC_ENABLED" envDefault:"false"`
	NewRelicAppName string `env:"FLAGR_NEWRELIC_NAME" envDefault:"flagr"`
	NewRelicKey     string `env:"FLAGR_NEWRELIC_KEY" envDefault:""`

	// StatsdEnabled - enable statsd metrics for all the endpoints and DB operations
	StatsdEnabled bool   `env:"FLAGR_STATSD_ENABLED" envDefault:"false"`
	StatsdHost    string `env:"FLAGR_STATSD_HOST" envDefault:"127.0.0.1"`
	StatsdPort    string `env:"FLAGR_STATSD_PORT" envDefault:"8125"`
	StatsdPrefix  string `env:"FLAGR_STATSD_PREFIX" envDefault:"flagr."`

	// RecorderEnabled - enable data records logging
	RecorderEnabled bool `env:"FLAGR_RECORDER_ENABLED" envDefault:"false"`
	// RecorderType - the pipeline to log data records, e.g. Kafka
	RecorderType string `env:"FLAGR_RECORDER_TYPE" envDefault:"kafka"`

	// Kafka related configurations for data records logging (Flagr Metrics)
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

	// Kinesis related configurations for data records logging (Flagr Metrics)
	RecorderKinesisStreamName          string        `env:"FLAGR_RECORDER_KINESIS_STREAM_NAME" envDefault:"flagr-records"`
	RecorderKinesisBacklogCount        int           `env:"FLAGR_RECORDER_KINESIS_BACKLOG_COUNT" envDefault:"500"`
	RecorderKinesisMaxConnections      int           `env:"FLAGR_RECORDER_KINESIS_MAX_CONNECTIONS" envDefault:"24"`
	RecorderKinesisFlushInterval       time.Duration `env:"FLAGR_RECORDER_KINESIS_FLUSH_INTERVAL" envDefault:"5s"`
	RecorderKinesisBatchCount          int           `env:"FLAGR_RECORDER_KINESIS_BATCH_COUNT" envDefault:"500"`
	RecorderKinesisBatchSize           int           `env:"FLAGR_RECORDER_KINESIS_BATCH_SIZE" envDefault:"0"`
	RecorderKinesisAggregateBatchCount int           `env:"FLAGR_RECORDER_KINESIS_AGGREGATE_BATCH_COUNT" envDefault:"4294967295"`
	RecorderKinesisAggregateBatchSize  int           `env:"FLAGR_RECORDER_KINESIS_AGGREGATE_BATCH_SIZE" envDefault:"51200"`
	RecorderKinesisVerbose             bool          `env:"FLAGR_RECORDER_KINESIS_VERBOSE" envDefault:"false"`

	/**
	JWTAuthEnabled enables the JWT Auth

	Via Cookies:
		The pattern of using JWT auth token using cookies is that it redirects to the URL to set cross subdomain cookie
		For example, redirect to auth.example.com/signin, which sets Cookie access_token=jwt_token for domain
		".example.com". One can also whitelist some routes so that they don't get blocked by JWT auth

	Via Headers:
		If you wish to use JWT Auth via headers you can simply set the header `Authorization Bearer [access_token]`

	Supported signing methods:
		* HS256, in this case `FLAGR_JWT_AUTH_SECRET` contains the passphrase
		* RS256, in this case `FLAGR_JWT_AUTH_SECRET` contains the key in PEM Format

	Note:
		If the access_token is present in both the header and cookie only the latest will be used
	*/
	JWTAuthEnabled              bool     `env:"FLAGR_JWT_AUTH_ENABLED" envDefault:"false"`
	JWTAuthDebug                bool     `env:"FLAGR_JWT_AUTH_DEBUG" envDefault:"false"`
	JWTAuthPrefixWhitelistPaths []string `env:"FLAGR_JWT_AUTH_WHITELIST_PATHS" envDefault:"/api/v1/evaluation,/static" envSeparator:","`
	JWTAuthExactWhitelistPaths  []string `env:"FLAGR_JWT_AUTH_EXACT_WHITELIST_PATHS" envDefault:",/" envSeparator:","`
	JWTAuthCookieTokenName      string   `env:"FLAGR_JWT_AUTH_COOKIE_TOKEN_NAME" envDefault:"access_token"`
	JWTAuthSecret               string   `env:"FLAGR_JWT_AUTH_SECRET" envDefault:""`
	JWTAuthNoTokenStatusCode    int      `env:"FLAGR_JWT_AUTH_NO_TOKEN_STATUS_CODE" envDefault:"307"` // "307" or "401"
	JWTAuthNoTokenRedirectURL   string   `env:"FLAGR_JWT_AUTH_NO_TOKEN_REDIRECT_URL" envDefault:""`
	JWTAuthUserProperty         string   `env:"FLAGR_JWT_AUTH_USER_PROPERTY" envDefault:"flagr_user"`

	// "HS256" and "RS256" supported
	JWTAuthSigningMethod string `env:"FLAGR_JWT_AUTH_SIGNING_METHOD" envDefault:"HS256"`
}{}
