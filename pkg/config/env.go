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
	// LogrusFormat sets the logrus logging formatter
	// Possible values: text, json
	LogrusFormat string `env:"FLAGR_LOGRUS_FORMAT" envDefault:"text"`
	// PProfEnabled - to enable the standard pprof of golang's http server
	PProfEnabled bool `env:"FLAGR_PPROF_ENABLED" envDefault:"true"`

	// MiddlewareVerboseLoggerEnabled - to enable the negroni-logrus logger for all the endpoints useful for debugging
	MiddlewareVerboseLoggerEnabled bool `env:"FLAGR_MIDDLEWARE_VERBOSE_LOGGER_ENABLED" envDefault:"true"`
	// MiddlewareGzipEnabled - to enable gzip middleware
	MiddlewareGzipEnabled bool `env:"FLAGR_MIDDLEWARE_GZIP_ENABLED" envDefault:"true"`

	// RateLimiterPerFlagPerSecondConsoleLogging - to rate limit the logging rate
	// per flag per second
	RateLimiterPerFlagPerSecondConsoleLogging int `env:"FLAGR_RATELIMITER_PERFLAG_PERSECOND_CONSOLE_LOGGING" envDefault:"100"`

	// EvalEnableDebug - controls if we want to return evaluation debugging information back to the api requests
	// Note that this is a global switch:
	//     if it's disabled, no evaluation debug info will be returned.
	//     if it's enabled, it respects evaluation request's enableDebug field
	EvalDebugEnabled bool `env:"FLAGR_EVAL_DEBUG_ENABLED" envDefault:"true"`
	// EvalLoggingEnabled - to enable the logging for eval results
	EvalLoggingEnabled bool `env:"FLAGR_EVAL_LOGGING_ENABLED" envDefault:"true"`
	// EvalCacheRefreshTimeout - timeout of getting the flags data from DB into the in-memory evaluation cache
	EvalCacheRefreshTimeout time.Duration `env:"FLAGR_EVALCACHE_REFRESHTIMEOUT" envDefault:"59s"`
	// EvalCacheRefreshInterval - time interval of getting the flags data from DB into the in-memory evaluation cache
	EvalCacheRefreshInterval time.Duration `env:"FLAGR_EVALCACHE_REFRESHINTERVAL" envDefault:"3s"`
	// EvalOnlyMode - will only expose the evaluation related endpoints.
	// This field will be derived from DBDriver
	EvalOnlyMode bool `env:"FLAGR_EVAL_ONLY_MODE" envDefault:"false"`

	/**
	DBDriver and DBConnectionStr define how we can write and read flags data.
	For databases, flagr supports sqlite3, mysql and postgres.
	For read-only evaluation, flagr supports file and http.

	Examples:

	FLAGR_DB_DBDRIVER     FLAGR_DB_DBCONNECTIONSTR
	=================     ===============================================================
	"sqlite3"             "/tmp/file.db"
	"sqlite3"             ":memory:"
	"mysql"               "root:@tcp(127.0.0.1:18100)/flagr?parseTime=true"
	"postgres"            "postgres://user:password@host:5432/flagr?sslmode=disable"

	"json_file"           "/tmp/flags.json"                    # (it automatically sets EvalOnlyMode=true)
	"json_http"           "https://example.com/flags.json"     # (it automatically sets EvalOnlyMode=true)

	*/
	DBDriver        string `env:"FLAGR_DB_DBDRIVER" envDefault:"sqlite3"`
	DBConnectionStr string `env:"FLAGR_DB_DBCONNECTIONSTR" envDefault:"flagr.sqlite"`
	// DBConnectionDebug controls whether to show the database connection debugging logs
	// warning: it may log the credentials to the stdout
	DBConnectionDebug bool `env:"FLAGR_DB_DBCONNECTION_DEBUG" envDefault:"true"`
	// DBConnectionRetryAttempts controls how we are going to retry on db connection when start the flagr server
	DBConnectionRetryAttempts uint          `env:"FLAGR_DB_DBCONNECTION_RETRY_ATTEMPTS" envDefault:"9"`
	DBConnectionRetryDelay    time.Duration `env:"FLAGR_DB_DBCONNECTION_RETRY_DELAY" envDefault:"100ms"`

	// CORSEnabled - enable CORS
	CORSEnabled bool `env:"FLAGR_CORS_ENABLED" envDefault:"true"`

	// SentryEnabled - enable Sentry and Sentry DSN
	SentryEnabled bool   `env:"FLAGR_SENTRY_ENABLED" envDefault:"false"`
	SentryDSN     string `env:"FLAGR_SENTRY_DSN" envDefault:""`

	// NewRelicEnabled - enable the NewRelic monitoring for all the endpoints and DB operations
	NewRelicEnabled                   bool   `env:"FLAGR_NEWRELIC_ENABLED" envDefault:"false"`
	NewRelicDistributedTracingEnabled bool   `env:"FLAGR_NEWRELIC_DISTRIBUTED_TRACING_ENABLED" envDefault:"false"`
	NewRelicAppName                   string `env:"FLAGR_NEWRELIC_NAME" envDefault:"flagr"`
	NewRelicKey                       string `env:"FLAGR_NEWRELIC_KEY" envDefault:""`

	// StatsdEnabled - enable statsd metrics for all the endpoints and DB operations
	StatsdEnabled        bool   `env:"FLAGR_STATSD_ENABLED" envDefault:"false"`
	StatsdHost           string `env:"FLAGR_STATSD_HOST" envDefault:"127.0.0.1"`
	StatsdPort           string `env:"FLAGR_STATSD_PORT" envDefault:"8125"`
	StatsdPrefix         string `env:"FLAGR_STATSD_PREFIX" envDefault:"flagr."`
	StatsdAPMEnabled     bool   `env:"FLAGR_STATSD_APM_ENABLED" envDefault:"false"`
	StatsdAPMPort        string `env:"FLAGR_STATSD_APM_PORT" envDefault:"8126"`
	StatsdAPMServiceName string `env:"FLAGR_STATSD_APM_SERVICE_NAME" envDefault:"flagr"`

	// PrometheusEnabled - enable prometheus metrics export
	PrometheusEnabled bool `env:"FLAGR_PROMETHEUS_ENABLED" envDefault:"false"`
	// PrometheusPath - set the path on which prometheus metrics are available to scrape
	PrometheusPath string `env:"FLAGR_PROMETHEUS_PATH" envDefault:"/metrics"`
	// PrometheusIncludeLatencyHistogram - set whether Prometheus should also export a histogram of request latencies (this increases cardinality significantly)
	PrometheusIncludeLatencyHistogram bool `env:"FLAGR_PROMETHEUS_INCLUDE_LATENCY_HISTOGRAM" envDefault:"false"`

	// RecorderEnabled - enable data records logging
	RecorderEnabled bool `env:"FLAGR_RECORDER_ENABLED" envDefault:"false"`
	// RecorderType - the pipeline to log data records, e.g. Kafka
	RecorderType string `env:"FLAGR_RECORDER_TYPE" envDefault:"kafka"`

	/**
	RecorderFrameOutputMode - indicates which data record frame output mode should we use.
	Possible values: payload_string, payload_raw_json

	* payload_string mode:
		it respects the encryption settings, and it will stringify the payload to unify
		the type of the output for both plaintext and encrypted payload.

		{"payload":"{\"evalContext\":{\"entityID\":\"123\"},\"flagID\":1,\"flagKey\":null,\"flagSnapshotID\":1,\"segmentID\":1,\"timestamp\":null,\"variantAttachment\":null,\"variantID\":1,\"variantKey\":\"control\"}","encrypted": false}

	* payload_raw_json mode:
		it ignores the encryption settings.

		{"payload":{"evalContext":{"entityID":"123"},"flagID":1,"flagKey":null,"flagSnapshotID":1,"segmentID":1,"timestamp":null,"variantAttachment":null,"variantID":1,"variantKey":"control"}}
	*/
	RecorderFrameOutputMode string `env:"FLAGR_RECORDER_FRAME_OUTPUT_MODE" envDefault:"payload_string"`

	// Kafka related configurations for data records logging (Flagr Metrics)
	RecorderKafkaVersion        string        `env:"FLAGR_RECORDER_KAFKA_VERSION" envDefault:"0.8.2.0"`
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

	// Pubsub related configurations for data records logging (Flagr Metrics)
	RecorderPubsubProjectID            string        `env:"FLAGR_RECORDER_PUBSUB_PROJECT_ID" envDefault:""`
	RecorderPubsubTopicName            string        `env:"FLAGR_RECORDER_PUBSUB_TOPIC_NAME" envDefault:"flagr-records"`
	RecorderPubsubKeyFile              string        `env:"FLAGR_RECORDER_PUBSUB_KEYFILE" envDefault:""`
	RecorderPubsubVerbose              bool          `env:"FLAGR_RECORDER_PUBSUB_VERBOSE" envDefault:"false"`
	RecorderPubsubVerboseCancelTimeout time.Duration `env:"FLAGR_RECORDER_PUBSUB_VERBOSE_CANCEL_TIMEOUT" envDefault:"5s"`

	/**
	JWTAuthEnabled enables the JWT Auth

	Via Cookies:
		The pattern of using JWT auth token using cookies is that it redirects to the URL to set cross subdomain cookie
		For example, redirect to auth.example.com/signin, which sets Cookie access_token=jwt_token for domain
		".example.com". One can also whitelist some routes so that they don't get blocked by JWT auth

	Via Headers:
		If you wish to use JWT Auth via headers you can simply set the header `Authorization Bearer [access_token]`

	Supported signing methods:
		* HS256/HS512, in this case `FLAGR_JWT_AUTH_SECRET` contains the passphrase
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
	// JWTAuthUserClaim can be used as the indicator of a user for created_by or updated_by.
	// E.g. sub, email, user, name, and etc in a JWT token.
	JWTAuthUserClaim string `env:"FLAGR_JWT_AUTH_USER_CLAIM" envDefault:"sub"`

	// "HS256" and "RS256" supported
	JWTAuthSigningMethod string `env:"FLAGR_JWT_AUTH_SIGNING_METHOD" envDefault:"HS256"`

	// Identify users through headers
	HeaderAuthEnabled   bool   `env:"FLAGR_HEADER_AUTH_ENABLED" envDefault:"false"`
	HeaderAuthUserField string `env:"FLAGR_HEADER_AUTH_USER_FIELD" envDefault:"X-Email"`

	// Authenticate with basic auth
	BasicAuthEnabled              bool     `env:"FLAGR_BASIC_AUTH_ENABLED" envDefault:"false"`
	BasicAuthUsername             string   `env:"FLAGR_BASIC_AUTH_USERNAME" envDefault:""`
	BasicAuthPassword             string   `env:"FLAGR_BASIC_AUTH_PASSWORD" envDefault:""`
	BasicAuthPrefixWhitelistPaths []string `env:"FLAGR_BASIC_AUTH_WHITELIST_PATHS" envDefault:"/api/v1/flags,/api/v1/evaluation" envSeparator:","`
	BasicAuthExactWhitelistPaths  []string `env:"FLAGR_BASIC_AUTH_EXACT_WHITELIST_PATHS" envDefault:"" envSeparator:","`

	// WebPrefix - base path for web and API
	// e.g. FLAGR_WEB_PREFIX=/foo
	// UI path  => localhost:18000/foo"
	// API path => localhost:18000/foo/api/v1"
	WebPrefix string `env:"FLAGR_WEB_PREFIX" envDefault:""`
}{}
