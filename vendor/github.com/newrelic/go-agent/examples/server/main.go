package main

import (
	"errors"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"os"
	"time"

	newrelic "github.com/newrelic/go-agent"
)

var (
	app newrelic.Application
)

func index(w http.ResponseWriter, r *http.Request) {
	io.WriteString(w, "hello world")
}

func versionHandler(w http.ResponseWriter, r *http.Request) {
	io.WriteString(w, "New Relic Go Agent Version: "+newrelic.Version)
}

func noticeError(w http.ResponseWriter, r *http.Request) {
	io.WriteString(w, "noticing an error")

	if txn, ok := w.(newrelic.Transaction); ok {
		txn.NoticeError(errors.New("my error message"))
	}
}

func noticeErrorWithAttributes(w http.ResponseWriter, r *http.Request) {
	io.WriteString(w, "noticing an error")

	if txn, ok := w.(newrelic.Transaction); ok {
		txn.NoticeError(newrelic.Error{
			Message: "uh oh. something went very wrong",
			Class:   "errors are aggregated by class",
			Attributes: map[string]interface{}{
				"important_number": 97232,
				"relevant_string":  "zap",
			},
		})
	}
}

func customEvent(w http.ResponseWriter, r *http.Request) {
	io.WriteString(w, "recording a custom event")

	app.RecordCustomEvent("my_event_type", map[string]interface{}{
		"myString": "hello",
		"myFloat":  0.603,
		"myInt":    123,
		"myBool":   true,
	})
}

func setName(w http.ResponseWriter, r *http.Request) {
	io.WriteString(w, "changing the transaction's name")

	if txn, ok := w.(newrelic.Transaction); ok {
		txn.SetName("other-name")
	}
}

func addAttribute(w http.ResponseWriter, r *http.Request) {
	io.WriteString(w, "adding attributes")

	if txn, ok := w.(newrelic.Transaction); ok {
		txn.AddAttribute("myString", "hello")
		txn.AddAttribute("myInt", 123)
	}
}

func background(w http.ResponseWriter, r *http.Request) {
	// Transactions started without an http.Request are classified as
	// background transactions.
	txn := app.StartTransaction("background", nil, nil)
	defer txn.End()

	io.WriteString(w, "background transaction")
	time.Sleep(150 * time.Millisecond)
}

func ignore(w http.ResponseWriter, r *http.Request) {
	if coinFlip := (0 == rand.Intn(2)); coinFlip {
		if txn, ok := w.(newrelic.Transaction); ok {
			txn.Ignore()
		}
		io.WriteString(w, "ignoring the transaction")
	} else {
		io.WriteString(w, "not ignoring the transaction")
	}
}

func segments(w http.ResponseWriter, r *http.Request) {
	txn, _ := w.(newrelic.Transaction)

	func() {
		defer newrelic.StartSegment(txn, "f1").End()

		func() {
			defer newrelic.StartSegment(txn, "f2").End()

			io.WriteString(w, "segments!")
			time.Sleep(10 * time.Millisecond)
		}()
		time.Sleep(15 * time.Millisecond)
	}()
	time.Sleep(20 * time.Millisecond)
}

func mysql(w http.ResponseWriter, r *http.Request) {
	txn, _ := w.(newrelic.Transaction)
	s := newrelic.DatastoreSegment{
		StartTime:          newrelic.StartSegmentNow(txn),
		Product:            newrelic.DatastoreMySQL,
		Collection:         "users",
		Operation:          "INSERT",
		ParameterizedQuery: "INSERT INTO users (name, age) VALUES ($1, $2)",
		QueryParameters: map[string]interface{}{
			"name": "Dracula",
			"age":  439,
		},
		Host:         "mysql-server-1",
		PortPathOrID: "3306",
		DatabaseName: "my_database",
	}
	defer s.End()

	time.Sleep(20 * time.Millisecond)
	io.WriteString(w, `performing fake query "INSERT * from users"`)
}

func external(w http.ResponseWriter, r *http.Request) {
	url := "http://example.com/"
	txn, _ := w.(newrelic.Transaction)
	// This demonstrates an external segment where only the URL is known. If
	// an http.Request is accessible then `StartExternalSegment` is
	// recommended. See the implementation of `NewRoundTripper` for an
	// example.
	defer newrelic.ExternalSegment{
		StartTime: newrelic.StartSegmentNow(txn),
		URL:       url,
	}.End()

	resp, err := http.Get(url)
	if nil != err {
		io.WriteString(w, err.Error())
		return
	}
	defer resp.Body.Close()
	io.Copy(w, resp.Body)
}

func roundtripper(w http.ResponseWriter, r *http.Request) {
	client := &http.Client{}
	txn, _ := w.(newrelic.Transaction)
	client.Transport = newrelic.NewRoundTripper(txn, nil)
	resp, err := client.Get("http://example.com/")
	if nil != err {
		io.WriteString(w, err.Error())
		return
	}
	defer resp.Body.Close()
	io.Copy(w, resp.Body)
}

func customMetric(w http.ResponseWriter, r *http.Request) {
	for _, vals := range r.Header {
		for _, v := range vals {
			// This custom metric will have the name
			// "Custom/HeaderLength" in the New Relic UI.
			app.RecordCustomMetric("HeaderLength", float64(len(v)))
		}
	}
	io.WriteString(w, "custom metric recorded")
}

func mustGetEnv(key string) string {
	if val := os.Getenv(key); "" != val {
		return val
	}
	panic(fmt.Sprintf("environment variable %s unset", key))
}

func main() {
	cfg := newrelic.NewConfig("Example App", mustGetEnv("NEW_RELIC_LICENSE_KEY"))
	cfg.Logger = newrelic.NewDebugLogger(os.Stdout)

	var err error
	app, err = newrelic.NewApplication(cfg)
	if nil != err {
		fmt.Println(err)
		os.Exit(1)
	}

	http.HandleFunc(newrelic.WrapHandleFunc(app, "/", index))
	http.HandleFunc(newrelic.WrapHandleFunc(app, "/version", versionHandler))
	http.HandleFunc(newrelic.WrapHandleFunc(app, "/notice_error", noticeError))
	http.HandleFunc(newrelic.WrapHandleFunc(app, "/notice_error_with_attributes", noticeErrorWithAttributes))
	http.HandleFunc(newrelic.WrapHandleFunc(app, "/custom_event", customEvent))
	http.HandleFunc(newrelic.WrapHandleFunc(app, "/set_name", setName))
	http.HandleFunc(newrelic.WrapHandleFunc(app, "/add_attribute", addAttribute))
	http.HandleFunc(newrelic.WrapHandleFunc(app, "/ignore", ignore))
	http.HandleFunc(newrelic.WrapHandleFunc(app, "/segments", segments))
	http.HandleFunc(newrelic.WrapHandleFunc(app, "/mysql", mysql))
	http.HandleFunc(newrelic.WrapHandleFunc(app, "/external", external))
	http.HandleFunc(newrelic.WrapHandleFunc(app, "/roundtripper", roundtripper))
	http.HandleFunc(newrelic.WrapHandleFunc(app, "/custommetric", customMetric))
	http.HandleFunc("/background", background)

	http.ListenAndServe(":8000", nil)
}
