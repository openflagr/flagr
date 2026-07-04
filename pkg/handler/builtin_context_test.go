package handler

import (
	"net/http"
	"testing"
	"time"

	"github.com/openflagr/flagr/pkg/config"
	"github.com/stretchr/testify/assert"
)

func TestInjectBuiltInContext_Disabled(t *testing.T) {
	config.Config.InjectedContextEnabled = false
	defer func() { config.Config.InjectedContextEnabled = false }()

	result := InjectBuiltInContext(nil, nil)
	assert.Nil(t, result, "should return nil when disabled (no conversion)")
}

func TestInjectBuiltInContext_CoreKeys(t *testing.T) {
	config.Config.InjectedContextEnabled = true
	defer func() { config.Config.InjectedContextEnabled = false }()

	t.Run("nil entityContext creates new map", func(t *testing.T) {
		ctx := InjectBuiltInContext(nil, nil).(map[string]any)
		assert.NotNil(t, ctx)
		assert.Contains(t, ctx, BuiltInKeyTs)
		assert.Contains(t, ctx, BuiltInKeyTsHour)
		assert.Contains(t, ctx, BuiltInKeyTsWeekday)
		assert.Contains(t, ctx, BuiltInKeyTsMonth)
	})

	t.Run("preserves client context", func(t *testing.T) {
		clientCtx := map[string]any{"country": "US", "tier": "premium"}
		ctx := InjectBuiltInContext(clientCtx, nil).(map[string]any)
		assert.Equal(t, "US", ctx["country"])
		assert.Equal(t, "premium", ctx["tier"])
		assert.Contains(t, ctx, BuiltInKeyTs)
	})

	t.Run("server overrides client ts", func(t *testing.T) {
		clientCtx := map[string]any{"@ts": 0}
		ctx := InjectBuiltInContext(clientCtx, nil).(map[string]any)
		now := time.Now().UTC().Unix()
		assert.NotEqual(t, int64(0), ctx[BuiltInKeyTs])
		assert.InDelta(t, now, ctx[BuiltInKeyTs], 2, "ts should be close to current time")
	})

	t.Run("ts is Unix epoch seconds", func(t *testing.T) {
		ctx := InjectBuiltInContext(nil, nil).(map[string]any)
		ts, ok := ctx[BuiltInKeyTs].(int64)
		assert.True(t, ok, "ts should be int64")
		now := time.Now().UTC().Unix()
		assert.InDelta(t, now, ts, 2)
	})

	t.Run("ts_hour is 0-23", func(t *testing.T) {
		ctx := InjectBuiltInContext(nil, nil).(map[string]any)
		hour, ok := ctx[BuiltInKeyTsHour].(int)
		assert.True(t, ok, "ts_hour should be int")
		assert.GreaterOrEqual(t, hour, 0)
		assert.LessOrEqual(t, hour, 23)
		assert.Equal(t, time.Now().UTC().Hour(), hour)
	})

	t.Run("ts_weekday is 0-6 (Sunday=0)", func(t *testing.T) {
		ctx := InjectBuiltInContext(nil, nil).(map[string]any)
		day, ok := ctx[BuiltInKeyTsWeekday].(int)
		assert.True(t, ok, "ts_weekday should be int")
		assert.GreaterOrEqual(t, day, 0)
		assert.LessOrEqual(t, day, 6)
		assert.Equal(t, int(time.Now().UTC().Weekday()), day)
	})

	t.Run("ts_month is 1-12", func(t *testing.T) {
		ctx := InjectBuiltInContext(nil, nil).(map[string]any)
		month, ok := ctx[BuiltInKeyTsMonth].(int)
		assert.True(t, ok, "ts_month should be int")
		assert.GreaterOrEqual(t, month, 1)
		assert.LessOrEqual(t, month, 12)
		assert.Equal(t, int(time.Now().UTC().Month()), month)
	})

	t.Run("non-map entityContext returns new map", func(t *testing.T) {
		ctx := InjectBuiltInContext("not a map", nil).(map[string]any)
		assert.NotNil(t, ctx)
		assert.Contains(t, ctx, BuiltInKeyTs)
	})
}

func TestInjectBuiltInContext_HTTPHeaders(t *testing.T) {
	config.Config.InjectedContextEnabled = true
	config.Config.InjectedContextHTTPHeaders = []string{"X-Environment", "X-Tenant-ID"}
	config.Config.InjectedContextHTTPHeaderPrefixes = []string{"CF-"}
	defer func() {
		config.Config.InjectedContextEnabled = false
		config.Config.InjectedContextHTTPHeaders = nil
		config.Config.InjectedContextHTTPHeaderPrefixes = nil
	}()

	t.Run("exact header match", func(t *testing.T) {
		r := &http.Request{
			Header: http.Header{
				"X-Environment": []string{"production"},
				"X-Tenant-ID":   []string{"acme-corp"},
			},
			Host: "flagr.example.com",
		}
		ctx := InjectBuiltInContext(nil, r).(map[string]any)
		assert.Equal(t, "production", ctx["@http_x_environment"])
		assert.Equal(t, "acme-corp", ctx["@http_x_tenant_id"])
	})

	t.Run("prefix header match", func(t *testing.T) {
		r := &http.Request{
			Header: http.Header{
				"CF-IPCountry": []string{"US"},
				"CF-Ray":       []string{"abc-123"},
			},
			Host: "flagr.example.com",
		}
		ctx := InjectBuiltInContext(nil, r).(map[string]any)
		assert.Equal(t, "US", ctx["@http_cf_ipcountry"])
		assert.Equal(t, "abc-123", ctx["@http_cf_ray"])
	})

	t.Run("non-matching header not injected", func(t *testing.T) {
		r := &http.Request{
			Header: http.Header{
				"X-Unrelated": []string{"value"},
			},
			Host: "flagr.example.com",
		}
		ctx := InjectBuiltInContext(nil, r).(map[string]any)
		assert.NotContains(t, ctx, "@http_x_unrelated")
	})

	t.Run("Host header special case", func(t *testing.T) {
		config.Config.InjectedContextHTTPHeaders = append(config.Config.InjectedContextHTTPHeaders, "Host")
		defer func() {
			config.Config.InjectedContextHTTPHeaders = []string{"X-Environment", "X-Tenant-ID"}
		}()

		r := &http.Request{
			Header: http.Header{},
			Host:   "flagr.staging.internal",
		}
		ctx := InjectBuiltInContext(nil, r).(map[string]any)
		assert.Equal(t, "flagr.staging.internal", ctx["@http_host"])
	})

	t.Run("Host not in config not injected", func(t *testing.T) {
		r := &http.Request{
			Header: http.Header{},
			Host:   "flagr.staging.internal",
		}
		ctx := InjectBuiltInContext(nil, r).(map[string]any)
		assert.NotContains(t, ctx, "@http_host")
	})

	t.Run("case insensitive header matching", func(t *testing.T) {
		config.Config.InjectedContextHTTPHeaders = []string{"x-environment"}
		defer func() {
			config.Config.InjectedContextHTTPHeaders = []string{"X-Environment", "X-Tenant-ID"}
		}()

		r := &http.Request{
			Header: http.Header{
				"X-Environment": []string{"staging"},
			},
			Host: "flagr.example.com",
		}
		ctx := InjectBuiltInContext(nil, r).(map[string]any)
		assert.Equal(t, "staging", ctx["@http_x_environment"])
	})

	t.Run("empty header value skipped", func(t *testing.T) {
		config.Config.InjectedContextHTTPHeaders = []string{"X-Empty"}
		defer func() {
			config.Config.InjectedContextHTTPHeaders = []string{"X-Environment", "X-Tenant-ID"}
		}()

		r := &http.Request{
			Header: http.Header{
				"X-Empty": []string{""},
			},
			Host: "flagr.example.com",
		}
		ctx := InjectBuiltInContext(nil, r).(map[string]any)
		assert.NotContains(t, ctx, "@http_x_empty")
	})

	t.Run("multi-value header joined with comma", func(t *testing.T) {
		config.Config.InjectedContextHTTPHeaders = []string{"X-Multi"}
		defer func() {
			config.Config.InjectedContextHTTPHeaders = []string{"X-Environment", "X-Tenant-ID"}
		}()

		r := &http.Request{
			Header: http.Header{
				"X-Multi": []string{"value1", "value2"},
			},
			Host: "flagr.example.com",
		}
		ctx := InjectBuiltInContext(nil, r).(map[string]any)
		assert.Equal(t, "value1, value2", ctx["@http_x_multi"])
	})

	t.Run("nil request skips header injection", func(t *testing.T) {
		ctx := InjectBuiltInContext(nil, nil).(map[string]any)
		assert.Contains(t, ctx, BuiltInKeyTs)
		assert.NotContains(t, ctx, "@http_x_environment")
	})
}
