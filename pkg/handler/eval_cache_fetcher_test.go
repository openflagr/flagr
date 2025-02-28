package handler

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/openflagr/flagr/pkg/config"
	"github.com/stretchr/testify/assert"
)

func TestJSONFileFetcher(t *testing.T) {
	t.Run("happy code path", func(t *testing.T) {
		jff := &jsonFileFetcher{filePath: "./testdata/sample_eval_cache.json"}
		fs, err := jff.fetch()
		assert.NoError(t, err)
		assert.NotZero(t, len(fs))
	})

	t.Run("non-exists file path", func(t *testing.T) {
		jff := &jsonFileFetcher{filePath: "./testdata/non-exists.json"}
		fs, err := jff.fetch()
		assert.Error(t, err)
		assert.Zero(t, fs)
	})
}

func TestJSONHTTPFetcher(t *testing.T) {
	t.Run("happy code path", func(t *testing.T) {
		h := func(w http.ResponseWriter, r *http.Request) {
			b, _ := os.ReadFile("./testdata/sample_eval_cache.json")
			w.Write(b)
		}

		server := httptest.NewServer(http.HandlerFunc(h))
		defer server.Close()

		jhf := &jsonHTTPFetcher{url: server.URL}
		fs, err := jhf.fetch()
		assert.NoError(t, err)
		assert.NotZero(t, len(fs))
	})

	t.Run("non-exists file path", func(t *testing.T) {
		jhf := &jsonHTTPFetcher{url: "http://invalid-url"}
		fs, err := jhf.fetch()
		assert.Error(t, err)
		assert.Zero(t, len(fs))
	})
}

func setDBDriverConfig(driver string, evalOnlyMode bool) (reset func()) {
	old := config.Config
	config.Config.DBDriver = driver
	config.Config.EvalOnlyMode = evalOnlyMode

	return func() {
		config.Config = old
	}
}

func TestNewFetcher(t *testing.T) {
	t.Run("regular db", func(t *testing.T) {
		reset := setDBDriverConfig("sqlite3", false)
		defer reset()

		fetcher, err := newFetcher()
		assert.NoError(t, err)
		assert.NotNil(t, fetcher)
	})

	t.Run("json file", func(t *testing.T) {
		reset := setDBDriverConfig("json_file", true)
		defer reset()

		fetcher, err := newFetcher()
		assert.NoError(t, err)
		assert.NotNil(t, fetcher)
	})

	t.Run("json http", func(t *testing.T) {
		reset := setDBDriverConfig("json_http", true)
		defer reset()

		fetcher, err := newFetcher()
		assert.NoError(t, err)
		assert.NotNil(t, fetcher)
	})

	t.Run("invalid driver", func(t *testing.T) {
		reset := setDBDriverConfig("invalid_driver", true)
		defer reset()

		fetcher, err := newFetcher()
		assert.Error(t, err)
		assert.Nil(t, fetcher)
	})
}
