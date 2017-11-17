# New Relic Go Agent Guide

* [Installation](#installation)
* [Config and Application](#config-and-application)
* [Logging](#logging)
  * [logrus](#logrus)
* [Transactions](#transactions)
* [Segments](#segments)
  * [Datastore Segments](#datastore-segments)
  * [External Segments](#external-segments)
* [Attributes](#attributes)
* [Custom Metrics](#custom-metrics)
* [Custom Events](#custom-events)
* [Request Queuing](#request-queuing)

## Installation

Installing the Go Agent is the same as installing any other Go library.  The
simplest way is to run:

```
go get github.com/newrelic/go-agent
```

Then import the `github.com/newrelic/go-agent` package in your application.

## Config and Application

* [config.go](config.go)
* [application.go](application.go)

In your `main` function or in an `init` block:

```go
config := newrelic.NewConfig("Your Application Name", "__YOUR_NEW_RELIC_LICENSE_KEY__")
app, err := newrelic.NewApplication(config)
```

Find your application in the New Relic UI.  Click on it to see the Go runtime
tab that shows information about goroutine counts, garbage collection, memory,
and CPU usage.

If you are working in a development environment or running unit tests, you may
not want the Go Agent to spawn goroutines or report to New Relic.  You're in
luck!  Set the config's `Enabled` field to false.  This makes the license key
optional.

```go
config := newrelic.NewConfig("Your Application Name", "")
config.Enabled = false
app, err := newrelic.NewApplication(config)
```

## Logging

* [log.go](log.go)

The agent's logging system is designed to be easily extensible.  By default, no
logging will occur.  To enable logging, assign the `Config.Logger` field to
something implementing the `Logger` interface.  A basic logging
implementation is included.

To log at debug level to standard out, set:

```go
config.Logger = newrelic.NewDebugLogger(os.Stdout)
```

To log at info level to a file, set:

```go
w, err := os.OpenFile("my_log_file", os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
if nil == err {
  config.Logger = newrelic.NewLogger(w)
}
```

### logrus

* [_integrations/nrlogrus/nrlogrus.go](_integrations/nrlogrus/nrlogrus.go)

If you are using `logrus` and would like to send the agent's log messages to its
standard logger, import the
`github.com/newrelic/go-agent/_integrations/nrlogrus` package, then set:

```go
config.Logger = nrlogrus.StandardLogger()
```

## Transactions

* [transaction.go](transaction.go)
* [Naming Transactions](#naming-transactions-and-metrics)
* [More info on Transactions](https://docs.newrelic.com/docs/apm/applications-menu/monitoring/transactions-page)

Transactions time requests and background tasks.  Each transaction should only
be used in a single goroutine.  Start a new transaction when you spawn a new
goroutine.

The simplest way to create transactions is to use
`Application.StartTransaction` and `Transaction.End`.

```go
txn := app.StartTransaction("transactionName", responseWriter, request)
defer txn.End()
```

The response writer and request parameters are optional.  Leave them `nil` to
instrument a background task.

```go
txn := app.StartTransaction("backgroundTask", nil, nil)
defer txn.End()
```

The transaction has helpful methods like `NoticeError` and `SetName`.
See more in [transaction.go](transaction.go).

If you are using the `http` standard library package, use `WrapHandle` and
`WrapHandleFunc`.  These wrappers automatically start and end transactions with
the request and response writer.  See [instrumentation.go](instrumentation.go).

```go
http.HandleFunc(newrelic.WrapHandleFunc(app, "/users", usersHandler))
```

To access the transaction in your handler, use type assertion on the response
writer passed to the handler.

```go
func myHandler(w http.ResponseWriter, r *http.Request) {
	if txn, ok := w.(newrelic.Transaction); ok {
		txn.NoticeError(errors.New("my error message"))
	}
}
```

## Segments

* [segments.go](segments.go)

Find out where the time in your transactions is being spent!  Each transaction
should only track segments in a single goroutine.

`Segment` is used to instrument functions, methods, and blocks of code. A
segment begins when its `StartTime` field is populated, and finishes when its
`End` method is called.

```go
segment := newrelic.Segment{}
segment.Name = "mySegmentName"
segment.StartTime = newrelic.StartSegmentNow(txn)
// ... code you want to time here ...
segment.End()
```

`StartSegment` is a convenient helper.  It creates a segment and starts it:

```go
segment := newrelic.StartSegment(txn, "mySegmentName")
// ... code you want to time here ...
segment.End()
```

Timing a function is easy using `StartSegment` and `defer`.  Just add the
following line to the beginning of that function:

```go
defer newrelic.StartSegment(txn, "mySegmentName").End()
```

Segments may be nested.  The segment being ended must be the most recently
started segment.

```go
s1 := newrelic.StartSegment(txn, "outerSegment")
s2 := newrelic.StartSegment(txn, "innerSegment")
// s2 must be ended before s1
s2.End()
s1.End()
```

A zero value segment may safely be ended.  Therefore, the following code
is safe even if the conditional fails:

```go
var s newrelic.Segment
if txn, ok := w.(newrelic.Transaction); ok {
	s.StartTime = newrelic.StartSegmentNow(txn),
}
// ... code you wish to time here ...
s.End()
```

### Datastore Segments

Datastore segments appear in the transaction "Breakdown table" and in the
"Databases" tab.

* [datastore.go](datastore.go)
* [More info on Databases tab](https://docs.newrelic.com/docs/apm/applications-menu/monitoring/databases-slow-queries-page)

Datastore segments are instrumented using `DatastoreSegment`.  Just like basic
segments, datastore segments begin when the `StartTime` field is populated and
finish when the `End` method is called.  Here is an example:

```go
s := newrelic.DatastoreSegment{
	// Product is the datastore type.  See the constants in datastore.go.
	Product: newrelic.DatastoreMySQL,
	// Collection is the table or group.
	Collection: "my_table",
	// Operation is the relevant action, e.g. "SELECT" or "GET".
	Operation: "SELECT",
}
s.StartTime = newrelic.StartSegmentNow(txn)
// ... make the datastore call
s.End()
```

This may be combined into a single line when instrumenting a datastore call
that spans an entire function call:

```go
defer newrelic.DatastoreSegment{
	StartTime:  newrelic.StartSegmentNow(txn),
	Product:    newrelic.DatastoreMySQL,
	Collection: "my_table",
	Operation:  "SELECT",
}.End()
```

### External Segments

External segments appear in the transaction "Breakdown table" and in the
"External services" tab.  

* [More info on External Services tab](https://docs.newrelic.com/docs/apm/applications-menu/monitoring/external-services-page)

External segments are instrumented using `ExternalSegment`.  Populate either the
`URL` or `Request` field to indicate the endpoint.  Here is an example:

```go
func external(txn newrelic.Transaction, url string) (*http.Response, error) {
	defer newrelic.ExternalSegment{
		StartTime: newrelic.StartSegmentNow(txn),
		URL:   url,
	}.End()

	return http.Get(url)
}
```

We recommend using the `Request` and `Response` fields since they provide more
information about the external call.  The `StartExternalSegment` helper is
useful when the request is available.  This function may be modified in the
future to add headers that will trace activity between applications that are
instrumented by New Relic.

```go
func external(txn newrelic.Transaction, req *http.Request) (*http.Response, error) {
	s := newrelic.StartExternalSegment(txn, req)
	response, err := http.DefaultClient.Do(req)
	s.Response = response
	s.End()
	return response, err
}
```

`NewRoundTripper` is another useful helper.  As with all segments, the round
tripper returned **must** only be used in the same goroutine as the transaction.

```go
client := &http.Client{}
client.Transport = newrelic.NewRoundTripper(txn, nil)
resp, err := client.Get("http://example.com/")
```

## Attributes

Attributes add context to errors and allow you to filter performance data
in Insights.

You may add them using the `Transaction.AddAttribute` method.

```go
txn.AddAttribute("key", "value")
txn.AddAttribute("product", "widget")
txn.AddAttribute("price", 19.99)
txn.AddAttribute("importantCustomer", true)
```

* [More info on Custom Attributes](https://docs.newrelic.com/docs/insights/new-relic-insights/decorating-events/insights-custom-attributes)

Some attributes are recorded automatically.  These are called agent attributes.
They are listed here:

* [attributes.go](attributes.go)

To disable one of these agents attributes, `AttributeResponseCode` for
example, modify the config like this:

```go
config.Attributes.Exclude = append(config.Attributes.Exclude, newrelic.AttributeResponseCode)
```

* [More info on Agent Attributes](https://docs.newrelic.com/docs/agents/manage-apm-agents/agent-metrics/agent-attributes)

## Custom Metrics

* [More info on Custom Metrics](https://docs.newrelic.com/docs/agents/go-agent/instrumentation/create-custom-metrics-go)

You may [create custom metrics](https://docs.newrelic.com/docs/agents/manage-apm-agents/agent-data/collect-custom-metrics)
via the `RecordCustomMetric` method.

```go
app.RecordCustomMetric(
	"CustomMetricName", // Name of your metric
	132,                // Value
)
```

**Note:** The Go Agent will automatically prepend the metric name you pass to
`RecordCustomMetric` (`"CustomMetricName"` above) with the string `Custom/`.
This means the above code would produce a metric named
`Custom/CustomMetricName`.  You'll also want to read over the
[Naming Transactions and Metrics](#naming-transactions-and-metrics) section below for
advice on coming up with appropriate metric names.

## Custom Events

You may track arbitrary events using custom Insights events.

```go
app.RecordCustomEvent("MyEventType", map[string]interface{}{
	"myString": "hello",
	"myFloat":  0.603,
	"myInt":    123,
	"myBool":   true,
})
```

## Request Queuing

If you are running a load balancer or reverse web proxy then you may configure
it to add a `X-Queue-Start` header with a Unix timestamp.  This will create a
band on the application overview chart showing queue time.

* [More info on Request Queuing](https://docs.newrelic.com/docs/apm/applications-menu/features/request-queuing-tracking-front-end-time)

## Error Reporting

You may track errors using the `Transaction.NoticeError` method.  The easiest
way to get started with `NoticeError` is to use errors based on
[Go's standard error interface](https://blog.golang.org/error-handling-and-go).

```go
txn.NoticeError(errors.New("my error message"))
```

`NoticeError` will work with *any* sort of object that implements Go's standard
error type interface -- not just `errorStrings` created via `errors.New`.  

If you're interested in sending more than an error *message* to New Relic, the
Go Agent also offers a `newrelic.Error` struct.

```go
txn.NoticeError(newrelic.Error{
	Message: "my error message",
	Class:   "IdentifierForError",
	Attributes: map[string]interface{}{
		"important_number": 97232,
		"relevant_string":  "zap",
	},
})
```

Using the `newrelic.Error` struct requires you to manually marshall your error
data into the `Message`, `Class`, and `Attributes` fields.  However, there's two
**advantages** to using the `newrelic.Error` struct.

First, by setting an error `Class`, New Relic will be able to aggregate errors
in the *Error Analytics* section of APM.  Second, the `Attributes` field allows
you to send through key/value pairs with additional error debugging information
(also exposed in the *Error Analytics* section of APM).

## Advanced Error Reporting

You're not limited to using Go's built-in error type or the provided
`newrelic.Error` struct.  The Go Agent provides three error interfaces

```go
type StackTracer interface {
	StackTrace() []uintptr
}

type ErrorClasser interface {
	ErrorClass() string
}

type ErrorAttributer interface {
	ErrorAttributes() map[string]interface{}
}
```

If you implement any of these on your own error structs, the `txn.NoticeError`
method will recognize these methods and use their return values to provide error
information.

For example, you could implement a custom error struct named `MyErrorWithClass`

```go
type MyErrorWithClass struct {

}
```

Then, you could implement both an `Error` method (per Go's standard `error`
interface) and an `ErrorClass` method (per the Go Agent `ErrorClasser`
interface) for this struct.

```go
func (e MyErrorWithClass) Error() string { return "A hard coded error message" }

// ErrorClass implements the ErrorClasser interface.
func (e MyErrorWithClass) ErrorClass() string { return "MyErrorClassForAggregation" }
```

Finally, you'd use your new error by creating a new instance of your struct and
passing it to the `NoticeError` method

```go
txn.NoticeError(MyErrorWithClass{})
```

While this is an oversimplified example, these interfaces give you a great deal
of control over what error information is available for your application.

## Naming Transactions and Metrics

You'll want to think carefully about how you name your transactions and custom
metrics.  If your program creates too many unique names, you may end up with a
[Metric Grouping Issue (or MGI)](https://docs.newrelic.com/docs/agents/manage-apm-agents/troubleshooting/metric-grouping-issues).

MGIs occur when the granularity of names is too fine, resulting in hundreds or
thousands of uniquely identified metrics and transactions.  One common cause of
MGIs is relying on the full URL name for metric naming in web transactions.  A
few major code paths may generate many different full URL paths to unique
documents, articles, page, etc. If the unique element of the URL path is
included in the metric name, each of these common paths will have its own unique
metric name.


## For More Help

There's a variety of places online to learn more about the Go Agent.

[The New Relic docs site](https://docs.newrelic.com/docs/agents/go-agent/get-started/introduction-new-relic-go)
contains a number of useful code samples and more context about how to use the Go Agent.

[New Relic's discussion forums](https://discuss.newrelic.com) have a dedicated
public forum [for the Go Agent](https://discuss.newrelic.com/c/support-products-agents/go-agent).

When in doubt, [the New Relic support site](https://support.newrelic.com/) is
the best place to get started troubleshooting an agent issue.
