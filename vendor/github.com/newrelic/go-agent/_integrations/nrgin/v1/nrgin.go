// Package nrgin introduces middleware to support the Gin framework.
//
//	router := gin.Default()
//	router.Use(nrgin.Middleware(app))
//
package nrgin

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/newrelic/go-agent"
	"github.com/newrelic/go-agent/internal"
)

func init() { internal.TrackUsage("integration", "framework", "gin", "v1") }

// headerResponseWriter exists to give the transaction access to response
// headers.
type headerResponseWriter struct{ w gin.ResponseWriter }

func (w *headerResponseWriter) Header() http.Header       { return w.w.Header() }
func (w *headerResponseWriter) Write([]byte) (int, error) { return 0, nil }
func (w *headerResponseWriter) WriteHeader(int)           {}

var _ http.ResponseWriter = &headerResponseWriter{}

type replacementResponseWriter struct {
	gin.ResponseWriter
	txn     newrelic.Transaction
	code    int
	written bool
}

var _ gin.ResponseWriter = &replacementResponseWriter{}

func (w *replacementResponseWriter) flushHeader() {
	if !w.written {
		w.txn.WriteHeader(w.code)
		w.written = true
	}
}

func (w *replacementResponseWriter) WriteHeader(code int) {
	w.code = code
	w.ResponseWriter.WriteHeader(code)
}

func (w *replacementResponseWriter) Write(data []byte) (int, error) {
	w.flushHeader()
	return w.ResponseWriter.Write(data)
}

func (w *replacementResponseWriter) WriteString(s string) (int, error) {
	w.flushHeader()
	return w.ResponseWriter.WriteString(s)
}

func (w *replacementResponseWriter) WriteHeaderNow() {
	w.flushHeader()
	w.ResponseWriter.WriteHeaderNow()
}

var (
	ctxKey = "newRelicTransaction"
)

// Transaction returns the transaction stored inside the context, or nil if not
// found.
func Transaction(c *gin.Context) newrelic.Transaction {
	if v, exists := c.Get(ctxKey); exists {
		if txn, ok := v.(newrelic.Transaction); ok {
			return txn
		}
	}
	return nil
}

// Middleware creates Gin middleware that instruments requests.
//
//	router := gin.Default()
//	router.Use(nrgin.Middleware(app))
//
func Middleware(app newrelic.Application) gin.HandlerFunc {
	return func(c *gin.Context) {
		name := c.HandlerName()
		w := &headerResponseWriter{w: c.Writer}
		txn := app.StartTransaction(name, w, c.Request)
		defer txn.End()

		c.Writer = &replacementResponseWriter{
			ResponseWriter: c.Writer,
			txn:            txn,
			code:           http.StatusOK,
		}
		c.Set(ctxKey, txn)
		c.Next()
	}
}
