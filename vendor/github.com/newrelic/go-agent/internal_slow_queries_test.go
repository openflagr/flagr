package newrelic

import (
	"strings"
	"testing"
	"time"

	"github.com/newrelic/go-agent/internal"
	"github.com/newrelic/go-agent/internal/crossagent"
)

func TestSlowQueryBasic(t *testing.T) {
	cfgfn := func(cfg *Config) {
		cfg.DatastoreTracer.SlowQuery.Threshold = 0
	}
	app := testApp(nil, cfgfn, t)
	txn := app.StartTransaction("hello", nil, helloRequest)
	s1 := DatastoreSegment{
		StartTime:          StartSegmentNow(txn),
		Product:            DatastoreMySQL,
		Collection:         "users",
		Operation:          "INSERT",
		ParameterizedQuery: "INSERT INTO users (name, age) VALUES ($1, $2)",
	}
	s1.End()
	txn.End()

	app.ExpectSlowQueries(t, []internal.WantSlowQuery{{
		Count:        1,
		MetricName:   "Datastore/statement/MySQL/users/INSERT",
		Query:        "INSERT INTO users (name, age) VALUES ($1, $2)",
		TxnName:      "WebTransaction/Go/hello",
		TxnURL:       "/hello",
		DatabaseName: "",
		Host:         "",
		PortPathOrID: "",
	}})
}

func TestSlowQueryLocallyDisabled(t *testing.T) {
	cfgfn := func(cfg *Config) {
		cfg.DatastoreTracer.SlowQuery.Threshold = 0
		cfg.DatastoreTracer.SlowQuery.Enabled = false
	}
	app := testApp(nil, cfgfn, t)
	txn := app.StartTransaction("hello", nil, helloRequest)
	s1 := DatastoreSegment{
		StartTime:          StartSegmentNow(txn),
		Product:            DatastoreMySQL,
		Collection:         "users",
		Operation:          "INSERT",
		ParameterizedQuery: "INSERT INTO users (name, age) VALUES ($1, $2)",
	}
	s1.End()
	txn.End()

	app.ExpectSlowQueries(t, []internal.WantSlowQuery{})
}

func TestSlowQueryRemotelyDisabled(t *testing.T) {
	cfgfn := func(cfg *Config) {
		cfg.DatastoreTracer.SlowQuery.Threshold = 0
	}
	replyfn := func(reply *internal.ConnectReply) {
		reply.CollectTraces = false
	}
	app := testApp(replyfn, cfgfn, t)
	txn := app.StartTransaction("hello", nil, helloRequest)
	s1 := DatastoreSegment{
		StartTime:          StartSegmentNow(txn),
		Product:            DatastoreMySQL,
		Collection:         "users",
		Operation:          "INSERT",
		ParameterizedQuery: "INSERT INTO users (name, age) VALUES ($1, $2)",
	}
	s1.End()
	txn.End()

	app.ExpectSlowQueries(t, []internal.WantSlowQuery{})
}

func TestSlowQueryBelowThreshold(t *testing.T) {
	cfgfn := func(cfg *Config) {
		cfg.DatastoreTracer.SlowQuery.Threshold = 1 * time.Hour
	}
	app := testApp(nil, cfgfn, t)
	txn := app.StartTransaction("hello", nil, helloRequest)
	s1 := DatastoreSegment{
		StartTime:          StartSegmentNow(txn),
		Product:            DatastoreMySQL,
		Collection:         "users",
		Operation:          "INSERT",
		ParameterizedQuery: "INSERT INTO users (name, age) VALUES ($1, $2)",
	}
	s1.End()
	txn.End()

	app.ExpectSlowQueries(t, []internal.WantSlowQuery{})
}

func TestSlowQueryDatabaseProvided(t *testing.T) {
	cfgfn := func(cfg *Config) {
		cfg.DatastoreTracer.SlowQuery.Threshold = 0
	}
	app := testApp(nil, cfgfn, t)
	txn := app.StartTransaction("hello", nil, helloRequest)
	s1 := DatastoreSegment{
		StartTime:          StartSegmentNow(txn),
		Product:            DatastoreMySQL,
		Collection:         "users",
		Operation:          "INSERT",
		ParameterizedQuery: "INSERT INTO users (name, age) VALUES ($1, $2)",
		DatabaseName:       "my_database",
	}
	s1.End()
	txn.End()

	app.ExpectSlowQueries(t, []internal.WantSlowQuery{{
		Count:        1,
		MetricName:   "Datastore/statement/MySQL/users/INSERT",
		Query:        "INSERT INTO users (name, age) VALUES ($1, $2)",
		TxnName:      "WebTransaction/Go/hello",
		TxnURL:       "/hello",
		DatabaseName: "my_database",
		Host:         "",
		PortPathOrID: "",
	}})
}

func TestSlowQueryHostProvided(t *testing.T) {
	cfgfn := func(cfg *Config) {
		cfg.DatastoreTracer.SlowQuery.Threshold = 0
	}
	app := testApp(nil, cfgfn, t)
	txn := app.StartTransaction("hello", nil, helloRequest)
	s1 := DatastoreSegment{
		StartTime:          StartSegmentNow(txn),
		Product:            DatastoreMySQL,
		Collection:         "users",
		Operation:          "INSERT",
		ParameterizedQuery: "INSERT INTO users (name, age) VALUES ($1, $2)",
		Host:               "db-server-1",
	}
	s1.End()
	txn.End()

	app.ExpectSlowQueries(t, []internal.WantSlowQuery{{
		Count:        1,
		MetricName:   "Datastore/statement/MySQL/users/INSERT",
		Query:        "INSERT INTO users (name, age) VALUES ($1, $2)",
		TxnName:      "WebTransaction/Go/hello",
		TxnURL:       "/hello",
		DatabaseName: "",
		Host:         "db-server-1",
		PortPathOrID: "unknown",
	}})
	scope := "WebTransaction/Go/hello"
	app.ExpectMetrics(t, append([]internal.WantMetric{
		{Name: "Datastore/all", Scope: "", Forced: true, Data: nil},
		{Name: "Datastore/allWeb", Scope: "", Forced: true, Data: nil},
		{Name: "Datastore/MySQL/all", Scope: "", Forced: true, Data: nil},
		{Name: "Datastore/MySQL/allWeb", Scope: "", Forced: true, Data: nil},
		{Name: "Datastore/operation/MySQL/INSERT", Scope: "", Forced: false, Data: nil},
		{Name: "Datastore/statement/MySQL/users/INSERT", Scope: "", Forced: false, Data: nil},
		{Name: "Datastore/statement/MySQL/users/INSERT", Scope: scope, Forced: false, Data: nil},
		{Name: "Datastore/instance/MySQL/db-server-1/unknown", Scope: "", Forced: false, Data: nil},
	}, webMetrics...))
}

func TestSlowQueryPortProvided(t *testing.T) {
	cfgfn := func(cfg *Config) {
		cfg.DatastoreTracer.SlowQuery.Threshold = 0
	}
	app := testApp(nil, cfgfn, t)
	txn := app.StartTransaction("hello", nil, helloRequest)
	s1 := DatastoreSegment{
		StartTime:          StartSegmentNow(txn),
		Product:            DatastoreMySQL,
		Collection:         "users",
		Operation:          "INSERT",
		ParameterizedQuery: "INSERT INTO users (name, age) VALUES ($1, $2)",
		PortPathOrID:       "98021",
	}
	s1.End()
	txn.End()

	app.ExpectSlowQueries(t, []internal.WantSlowQuery{{
		Count:        1,
		MetricName:   "Datastore/statement/MySQL/users/INSERT",
		Query:        "INSERT INTO users (name, age) VALUES ($1, $2)",
		TxnName:      "WebTransaction/Go/hello",
		TxnURL:       "/hello",
		DatabaseName: "",
		Host:         "unknown",
		PortPathOrID: "98021",
	}})
	scope := "WebTransaction/Go/hello"
	app.ExpectMetrics(t, append([]internal.WantMetric{
		{Name: "Datastore/all", Scope: "", Forced: true, Data: nil},
		{Name: "Datastore/allWeb", Scope: "", Forced: true, Data: nil},
		{Name: "Datastore/MySQL/all", Scope: "", Forced: true, Data: nil},
		{Name: "Datastore/MySQL/allWeb", Scope: "", Forced: true, Data: nil},
		{Name: "Datastore/operation/MySQL/INSERT", Scope: "", Forced: false, Data: nil},
		{Name: "Datastore/statement/MySQL/users/INSERT", Scope: "", Forced: false, Data: nil},
		{Name: "Datastore/statement/MySQL/users/INSERT", Scope: scope, Forced: false, Data: nil},
		{Name: "Datastore/instance/MySQL/unknown/98021", Scope: "", Forced: false, Data: nil},
	}, webMetrics...))
}

func TestSlowQueryHostPortProvided(t *testing.T) {
	cfgfn := func(cfg *Config) {
		cfg.DatastoreTracer.SlowQuery.Threshold = 0
	}
	app := testApp(nil, cfgfn, t)
	txn := app.StartTransaction("hello", nil, helloRequest)
	s1 := DatastoreSegment{
		StartTime:          StartSegmentNow(txn),
		Product:            DatastoreMySQL,
		Collection:         "users",
		Operation:          "INSERT",
		ParameterizedQuery: "INSERT INTO users (name, age) VALUES ($1, $2)",
		Host:               "db-server-1",
		PortPathOrID:       "98021",
	}
	s1.End()
	txn.End()

	app.ExpectSlowQueries(t, []internal.WantSlowQuery{{
		Count:        1,
		MetricName:   "Datastore/statement/MySQL/users/INSERT",
		Query:        "INSERT INTO users (name, age) VALUES ($1, $2)",
		TxnName:      "WebTransaction/Go/hello",
		TxnURL:       "/hello",
		DatabaseName: "",
		Host:         "db-server-1",
		PortPathOrID: "98021",
	}})
	scope := "WebTransaction/Go/hello"
	app.ExpectMetrics(t, append([]internal.WantMetric{
		{Name: "Datastore/all", Scope: "", Forced: true, Data: nil},
		{Name: "Datastore/allWeb", Scope: "", Forced: true, Data: nil},
		{Name: "Datastore/MySQL/all", Scope: "", Forced: true, Data: nil},
		{Name: "Datastore/MySQL/allWeb", Scope: "", Forced: true, Data: nil},
		{Name: "Datastore/operation/MySQL/INSERT", Scope: "", Forced: false, Data: nil},
		{Name: "Datastore/statement/MySQL/users/INSERT", Scope: "", Forced: false, Data: nil},
		{Name: "Datastore/statement/MySQL/users/INSERT", Scope: scope, Forced: false, Data: nil},
		{Name: "Datastore/instance/MySQL/db-server-1/98021", Scope: "", Forced: false, Data: nil},
	}, webMetrics...))
}

func TestSlowQueryAggregation(t *testing.T) {
	cfgfn := func(cfg *Config) {
		cfg.DatastoreTracer.SlowQuery.Threshold = 0
	}
	app := testApp(nil, cfgfn, t)
	txn := app.StartTransaction("hello", nil, helloRequest)
	DatastoreSegment{
		StartTime:          StartSegmentNow(txn),
		Product:            DatastoreMySQL,
		Collection:         "users",
		Operation:          "INSERT",
		ParameterizedQuery: "INSERT INTO users (name, age) VALUES ($1, $2)",
	}.End()
	DatastoreSegment{
		StartTime:          StartSegmentNow(txn),
		Product:            DatastoreMySQL,
		Collection:         "users",
		Operation:          "INSERT",
		ParameterizedQuery: "INSERT INTO users (name, age) VALUES ($1, $2)",
	}.End()
	DatastoreSegment{
		StartTime:          StartSegmentNow(txn),
		Product:            DatastorePostgres,
		Collection:         "products",
		Operation:          "INSERT",
		ParameterizedQuery: "INSERT INTO products (name, price) VALUES ($1, $2)",
	}.End()
	txn.End()

	app.ExpectSlowQueries(t, []internal.WantSlowQuery{{
		Count:        2,
		MetricName:   "Datastore/statement/MySQL/users/INSERT",
		Query:        "INSERT INTO users (name, age) VALUES ($1, $2)",
		TxnName:      "WebTransaction/Go/hello",
		TxnURL:       "/hello",
		DatabaseName: "",
		Host:         "",
		PortPathOrID: "",
	}, {
		Count:        1,
		MetricName:   "Datastore/statement/Postgres/products/INSERT",
		Query:        "INSERT INTO products (name, price) VALUES ($1, $2)",
		TxnName:      "WebTransaction/Go/hello",
		TxnURL:       "/hello",
		DatabaseName: "",
		Host:         "",
		PortPathOrID: "",
	},
	})
}

func TestSlowQueryMissingQuery(t *testing.T) {
	cfgfn := func(cfg *Config) {
		cfg.DatastoreTracer.SlowQuery.Threshold = 0
	}
	app := testApp(nil, cfgfn, t)
	txn := app.StartTransaction("hello", nil, helloRequest)
	s1 := DatastoreSegment{
		StartTime:  StartSegmentNow(txn),
		Product:    DatastoreMySQL,
		Collection: "users",
		Operation:  "INSERT",
	}
	s1.End()
	txn.End()

	app.ExpectSlowQueries(t, []internal.WantSlowQuery{{
		Count:        1,
		MetricName:   "Datastore/statement/MySQL/users/INSERT",
		Query:        "'INSERT' on 'users' using 'MySQL'",
		TxnName:      "WebTransaction/Go/hello",
		TxnURL:       "/hello",
		DatabaseName: "",
		Host:         "",
		PortPathOrID: "",
	}})
}

func TestSlowQueryMissingEverything(t *testing.T) {
	cfgfn := func(cfg *Config) {
		cfg.DatastoreTracer.SlowQuery.Threshold = 0
	}
	app := testApp(nil, cfgfn, t)
	txn := app.StartTransaction("hello", nil, helloRequest)
	s1 := DatastoreSegment{
		StartTime: StartSegmentNow(txn),
	}
	s1.End()
	txn.End()

	app.ExpectSlowQueries(t, []internal.WantSlowQuery{{
		Count:        1,
		MetricName:   "Datastore/operation/Unknown/other",
		Query:        "'other' on 'unknown' using 'Unknown'",
		TxnName:      "WebTransaction/Go/hello",
		TxnURL:       "/hello",
		DatabaseName: "",
		Host:         "",
		PortPathOrID: "",
	}})
	scope := "WebTransaction/Go/hello"
	app.ExpectMetrics(t, append([]internal.WantMetric{
		{Name: "Datastore/all", Scope: "", Forced: true, Data: nil},
		{Name: "Datastore/allWeb", Scope: "", Forced: true, Data: nil},
		{Name: "Datastore/Unknown/all", Scope: "", Forced: true, Data: nil},
		{Name: "Datastore/Unknown/allWeb", Scope: "", Forced: true, Data: nil},
		{Name: "Datastore/operation/Unknown/other", Scope: "", Forced: false, Data: nil},
		{Name: "Datastore/operation/Unknown/other", Scope: scope, Forced: false, Data: nil},
	}, webMetrics...))
}

func TestSlowQueryWithQueryParameters(t *testing.T) {
	cfgfn := func(cfg *Config) {
		cfg.DatastoreTracer.SlowQuery.Threshold = 0
	}
	app := testApp(nil, cfgfn, t)
	txn := app.StartTransaction("hello", nil, helloRequest)
	params := map[string]interface{}{
		"str": "zap",
		"int": 123,
	}
	s1 := DatastoreSegment{
		StartTime:          StartSegmentNow(txn),
		Product:            DatastoreMySQL,
		Collection:         "users",
		Operation:          "INSERT",
		ParameterizedQuery: "INSERT INTO users (name, age) VALUES ($1, $2)",
		QueryParameters:    params,
	}
	s1.End()
	txn.End()

	app.ExpectSlowQueries(t, []internal.WantSlowQuery{{
		Count:        1,
		MetricName:   "Datastore/statement/MySQL/users/INSERT",
		Query:        "INSERT INTO users (name, age) VALUES ($1, $2)",
		TxnName:      "WebTransaction/Go/hello",
		TxnURL:       "/hello",
		DatabaseName: "",
		Host:         "",
		PortPathOrID: "",
		Params:       params,
	}})
}

func TestSlowQueryHighSecurity(t *testing.T) {
	cfgfn := func(cfg *Config) {
		cfg.DatastoreTracer.SlowQuery.Threshold = 0
		cfg.HighSecurity = true
	}
	app := testApp(nil, cfgfn, t)
	txn := app.StartTransaction("hello", nil, helloRequest)
	params := map[string]interface{}{
		"str": "zap",
		"int": 123,
	}
	s1 := DatastoreSegment{
		StartTime:          StartSegmentNow(txn),
		Product:            DatastoreMySQL,
		Collection:         "users",
		Operation:          "INSERT",
		ParameterizedQuery: "INSERT INTO users (name, age) VALUES ($1, $2)",
		QueryParameters:    params,
	}
	s1.End()
	txn.End()

	app.ExpectSlowQueries(t, []internal.WantSlowQuery{{
		Count:        1,
		MetricName:   "Datastore/statement/MySQL/users/INSERT",
		Query:        "INSERT INTO users (name, age) VALUES ($1, $2)",
		TxnName:      "WebTransaction/Go/hello",
		TxnURL:       "/hello",
		DatabaseName: "",
		Host:         "",
		PortPathOrID: "",
		Params:       nil,
	}})
}

func TestSlowQueryInvalidParameters(t *testing.T) {
	cfgfn := func(cfg *Config) {
		cfg.DatastoreTracer.SlowQuery.Threshold = 0
	}
	app := testApp(nil, cfgfn, t)
	txn := app.StartTransaction("hello", nil, helloRequest)
	params := map[string]interface{}{
		"str":                               "zap",
		"int":                               123,
		"invalid_value":                     struct{}{},
		strings.Repeat("key-too-long", 100): 1,
		"long-key":                          strings.Repeat("A", 300),
	}
	s1 := DatastoreSegment{
		StartTime:          StartSegmentNow(txn),
		Product:            DatastoreMySQL,
		Collection:         "users",
		Operation:          "INSERT",
		ParameterizedQuery: "INSERT INTO users (name, age) VALUES ($1, $2)",
		QueryParameters:    params,
	}
	s1.End()
	txn.End()

	app.ExpectSlowQueries(t, []internal.WantSlowQuery{{
		Count:        1,
		MetricName:   "Datastore/statement/MySQL/users/INSERT",
		Query:        "INSERT INTO users (name, age) VALUES ($1, $2)",
		TxnName:      "WebTransaction/Go/hello",
		TxnURL:       "/hello",
		DatabaseName: "",
		Host:         "",
		PortPathOrID: "",
		Params: map[string]interface{}{
			"str":      "zap",
			"int":      123,
			"long-key": strings.Repeat("A", 255),
		},
	}})
}

func TestSlowQueryParametersDisabled(t *testing.T) {
	cfgfn := func(cfg *Config) {
		cfg.DatastoreTracer.SlowQuery.Threshold = 0
		cfg.DatastoreTracer.QueryParameters.Enabled = false
	}
	app := testApp(nil, cfgfn, t)
	txn := app.StartTransaction("hello", nil, helloRequest)
	params := map[string]interface{}{
		"str": "zap",
		"int": 123,
	}
	s1 := DatastoreSegment{
		StartTime:          StartSegmentNow(txn),
		Product:            DatastoreMySQL,
		Collection:         "users",
		Operation:          "INSERT",
		ParameterizedQuery: "INSERT INTO users (name, age) VALUES ($1, $2)",
		QueryParameters:    params,
	}
	s1.End()
	txn.End()

	app.ExpectSlowQueries(t, []internal.WantSlowQuery{{
		Count:        1,
		MetricName:   "Datastore/statement/MySQL/users/INSERT",
		Query:        "INSERT INTO users (name, age) VALUES ($1, $2)",
		TxnName:      "WebTransaction/Go/hello",
		TxnURL:       "/hello",
		DatabaseName: "",
		Host:         "",
		PortPathOrID: "",
		Params:       nil,
	}})
}

func TestSlowQueryInstanceDisabled(t *testing.T) {
	cfgfn := func(cfg *Config) {
		cfg.DatastoreTracer.SlowQuery.Threshold = 0
		cfg.DatastoreTracer.InstanceReporting.Enabled = false
	}
	app := testApp(nil, cfgfn, t)
	txn := app.StartTransaction("hello", nil, helloRequest)
	s1 := DatastoreSegment{
		StartTime:          StartSegmentNow(txn),
		Product:            DatastoreMySQL,
		Collection:         "users",
		Operation:          "INSERT",
		ParameterizedQuery: "INSERT INTO users (name, age) VALUES ($1, $2)",
		Host:               "db-server-1",
	}
	s1.End()
	txn.End()

	app.ExpectSlowQueries(t, []internal.WantSlowQuery{{
		Count:        1,
		MetricName:   "Datastore/statement/MySQL/users/INSERT",
		Query:        "INSERT INTO users (name, age) VALUES ($1, $2)",
		TxnName:      "WebTransaction/Go/hello",
		TxnURL:       "/hello",
		DatabaseName: "",
		Host:         "",
		PortPathOrID: "",
	}})
	scope := "WebTransaction/Go/hello"
	app.ExpectMetrics(t, append([]internal.WantMetric{
		{Name: "Datastore/all", Scope: "", Forced: true, Data: nil},
		{Name: "Datastore/allWeb", Scope: "", Forced: true, Data: nil},
		{Name: "Datastore/MySQL/all", Scope: "", Forced: true, Data: nil},
		{Name: "Datastore/MySQL/allWeb", Scope: "", Forced: true, Data: nil},
		{Name: "Datastore/operation/MySQL/INSERT", Scope: "", Forced: false, Data: nil},
		{Name: "Datastore/statement/MySQL/users/INSERT", Scope: "", Forced: false, Data: nil},
		{Name: "Datastore/statement/MySQL/users/INSERT", Scope: scope, Forced: false, Data: nil},
	}, webMetrics...))
}

func TestSlowQueryInstanceDisabledLocalhost(t *testing.T) {
	cfgfn := func(cfg *Config) {
		cfg.DatastoreTracer.SlowQuery.Threshold = 0
		cfg.DatastoreTracer.InstanceReporting.Enabled = false
	}
	app := testApp(nil, cfgfn, t)
	txn := app.StartTransaction("hello", nil, helloRequest)
	s1 := DatastoreSegment{
		StartTime:          StartSegmentNow(txn),
		Product:            DatastoreMySQL,
		Collection:         "users",
		Operation:          "INSERT",
		ParameterizedQuery: "INSERT INTO users (name, age) VALUES ($1, $2)",
		Host:               "localhost",
		PortPathOrID:       "3306",
	}
	s1.End()
	txn.End()

	app.ExpectSlowQueries(t, []internal.WantSlowQuery{{
		Count:        1,
		MetricName:   "Datastore/statement/MySQL/users/INSERT",
		Query:        "INSERT INTO users (name, age) VALUES ($1, $2)",
		TxnName:      "WebTransaction/Go/hello",
		TxnURL:       "/hello",
		DatabaseName: "",
		Host:         "",
		PortPathOrID: "",
	}})
	scope := "WebTransaction/Go/hello"
	app.ExpectMetrics(t, append([]internal.WantMetric{
		{Name: "Datastore/all", Scope: "", Forced: true, Data: nil},
		{Name: "Datastore/allWeb", Scope: "", Forced: true, Data: nil},
		{Name: "Datastore/MySQL/all", Scope: "", Forced: true, Data: nil},
		{Name: "Datastore/MySQL/allWeb", Scope: "", Forced: true, Data: nil},
		{Name: "Datastore/operation/MySQL/INSERT", Scope: "", Forced: false, Data: nil},
		{Name: "Datastore/statement/MySQL/users/INSERT", Scope: "", Forced: false, Data: nil},
		{Name: "Datastore/statement/MySQL/users/INSERT", Scope: scope, Forced: false, Data: nil},
	}, webMetrics...))
}

func TestSlowQueryDatabaseNameDisabled(t *testing.T) {
	cfgfn := func(cfg *Config) {
		cfg.DatastoreTracer.SlowQuery.Threshold = 0
		cfg.DatastoreTracer.DatabaseNameReporting.Enabled = false
	}
	app := testApp(nil, cfgfn, t)
	txn := app.StartTransaction("hello", nil, helloRequest)
	s1 := DatastoreSegment{
		StartTime:          StartSegmentNow(txn),
		Product:            DatastoreMySQL,
		Collection:         "users",
		Operation:          "INSERT",
		ParameterizedQuery: "INSERT INTO users (name, age) VALUES ($1, $2)",
		DatabaseName:       "db-server-1",
	}
	s1.End()
	txn.End()

	app.ExpectSlowQueries(t, []internal.WantSlowQuery{{
		Count:        1,
		MetricName:   "Datastore/statement/MySQL/users/INSERT",
		Query:        "INSERT INTO users (name, age) VALUES ($1, $2)",
		TxnName:      "WebTransaction/Go/hello",
		TxnURL:       "/hello",
		DatabaseName: "",
		Host:         "",
		PortPathOrID: "",
	}})
}

func TestDatastoreAPICrossAgent(t *testing.T) {
	var testcases []struct {
		TestName string `json:"test_name"`
		Input    struct {
			Parameters struct {
				Product      string `json:"product"`
				Collection   string `json:"collection"`
				Operation    string `json:"operation"`
				Host         string `json:"host"`
				PortPathOrID string `json:"port_path_or_id"`
				DatabaseName string `json:"database_name"`
			} `json:"parameters"`
			IsWeb          bool   `json:"is_web"`
			SystemHostname string `json:"system_hostname"`
			Configuration  struct {
				InstanceEnabled bool `json:"datastore_tracer.instance_reporting.enabled"`
				DatabaseEnabled bool `json:"datastore_tracer.database_name_reporting.enabled"`
			}
		}
		Expectation struct {
			MetricsScoped   []string `json:"metrics_scoped"`
			MetricsUnscoped []string `json:"metrics_unscoped"`
			Trace           struct {
				MetricName   string `json:"metric_name"`
				Host         string `json:"host"`
				PortPathOrID string `json:"port_path_or_id"`
				DatabaseName string `json:"database_name"`
			} `json:"transaction_segment_and_slow_query_trace"`
		}
	}

	err := crossagent.ReadJSON("datastores/datastore_api.json", &testcases)
	if err != nil {
		t.Fatal(err)
	}

	for _, tc := range testcases {
		query := "my query"
		cfgfn := func(cfg *Config) {
			cfg.DatastoreTracer.SlowQuery.Threshold = 0
			cfg.DatastoreTracer.InstanceReporting.Enabled =
				tc.Input.Configuration.InstanceEnabled
			cfg.DatastoreTracer.DatabaseNameReporting.Enabled =
				tc.Input.Configuration.DatabaseEnabled
		}
		app := testApp(nil, cfgfn, t)
		var txn Transaction
		var txnURL string
		if tc.Input.IsWeb {
			txnURL = helloPath
			txn = app.StartTransaction("hello", nil, helloRequest)
		} else {
			txn = app.StartTransaction("hello", nil, nil)
		}
		DatastoreSegment{
			StartTime:          StartSegmentNow(txn),
			Product:            DatastoreProduct(tc.Input.Parameters.Product),
			Operation:          tc.Input.Parameters.Operation,
			Collection:         tc.Input.Parameters.Collection,
			PortPathOrID:       tc.Input.Parameters.PortPathOrID,
			Host:               tc.Input.Parameters.Host,
			DatabaseName:       tc.Input.Parameters.DatabaseName,
			ParameterizedQuery: query,
		}.End()
		txn.End()

		var metrics []internal.WantMetric
		var scope string
		if tc.Input.IsWeb {
			scope = "WebTransaction/Go/hello"
			metrics = append([]internal.WantMetric{}, webMetrics...)
		} else {
			scope = "OtherTransaction/Go/hello"
			metrics = append([]internal.WantMetric{}, backgroundMetrics...)
		}

		for _, m := range tc.Expectation.MetricsScoped {
			metrics = append(metrics, internal.WantMetric{
				Name: m, Scope: scope, Forced: nil, Data: nil,
			})
		}
		for _, m := range tc.Expectation.MetricsUnscoped {
			metrics = append(metrics, internal.WantMetric{
				Name: m, Scope: "", Forced: nil, Data: nil,
			})
		}

		expectTraceHost := tc.Expectation.Trace.Host
		if tc.Input.SystemHostname != "" {
			for i := range metrics {
				metrics[i].Name = strings.Replace(metrics[i].Name,
					tc.Input.SystemHostname,
					internal.ThisHost, -1)
			}
			expectTraceHost = strings.Replace(expectTraceHost,
				tc.Input.SystemHostname,
				internal.ThisHost, -1)
		}

		tt := internal.ExtendValidator(t, tc.TestName)
		app.ExpectMetrics(tt, metrics)
		app.ExpectSlowQueries(tt, []internal.WantSlowQuery{{
			Count:        1,
			MetricName:   tc.Expectation.Trace.MetricName,
			TxnName:      scope,
			DatabaseName: tc.Expectation.Trace.DatabaseName,
			Host:         expectTraceHost,
			PortPathOrID: tc.Expectation.Trace.PortPathOrID,
			TxnURL:       txnURL,
			Query:        query,
		}})
	}
}
