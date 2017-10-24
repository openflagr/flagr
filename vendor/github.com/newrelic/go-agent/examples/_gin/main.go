package main

import (
	"fmt"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/newrelic/go-agent"
	"github.com/newrelic/go-agent/_integrations/nrgin/v1"
)

func makeGinEndpoint(s string) func(*gin.Context) {
	return func(c *gin.Context) {
		c.Writer.WriteString(s)
	}
}

func v1login(c *gin.Context)  { c.Writer.WriteString("v1 login") }
func v1submit(c *gin.Context) { c.Writer.WriteString("v1 submit") }
func v1read(c *gin.Context)   { c.Writer.WriteString("v1 read") }

func endpoint404(c *gin.Context) {
	c.Writer.WriteHeader(404)
	c.Writer.WriteString("returning 404")
}

func endpointChangeCode(c *gin.Context) {
	// gin.ResponseWriter buffers the response code so that it can be
	// changed before the first write.
	c.Writer.WriteHeader(404)
	c.Writer.WriteHeader(200)
	c.Writer.WriteString("actually ok!")
}

func endpointResponseHeaders(c *gin.Context) {
	// Since gin.ResponseWriter buffers the response code, response headers
	// can be set afterwards.
	c.Writer.WriteHeader(200)
	c.Writer.Header().Set("Content-Type", "application/json")
	c.Writer.WriteString(`{"zip":"zap"}`)
}

func endpointNotFound(c *gin.Context) {
	c.Writer.WriteString("there's no endpoint for that!")
}

func endpointAccessTransaction(c *gin.Context) {
	if txn := nrgin.Transaction(c); nil != txn {
		txn.SetName("custom-name")
	}
	c.Writer.WriteString("changed the name of the transaction!")
}

func mustGetEnv(key string) string {
	if val := os.Getenv(key); "" != val {
		return val
	}
	panic(fmt.Sprintf("environment variable %s unset", key))
}

func main() {
	cfg := newrelic.NewConfig("Gin App", mustGetEnv("NEW_RELIC_LICENSE_KEY"))
	cfg.Logger = newrelic.NewDebugLogger(os.Stdout)
	app, err := newrelic.NewApplication(cfg)
	if nil != err {
		fmt.Println(err)
		os.Exit(1)
	}

	router := gin.Default()
	router.Use(nrgin.Middleware(app))

	router.GET("/404", endpoint404)
	router.GET("/change", endpointChangeCode)
	router.GET("/headers", endpointResponseHeaders)
	router.GET("/txn", endpointAccessTransaction)

	// Since the handler function name is used as the transaction name,
	// anonymous functions do not get usefully named.  We encourage
	// transforming anonymous functions into named functions.
	router.GET("/anon", func(c *gin.Context) {
		c.Writer.WriteString("anonymous function handler")
	})

	v1 := router.Group("/v1")
	v1.GET("/login", v1login)
	v1.GET("/submit", v1submit)
	v1.GET("/read", v1read)

	router.NoRoute(endpointNotFound)

	router.Run(":8000")
}
