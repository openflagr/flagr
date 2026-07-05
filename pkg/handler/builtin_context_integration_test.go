package handler

import (
	"net/http"
	"strconv"
	"testing"
	"time"

	"github.com/openflagr/flagr/pkg/config"
	"github.com/openflagr/flagr/pkg/entity"
	"github.com/openflagr/flagr/swagger_gen/models"
	"github.com/prashantv/gostub"
	"github.com/stretchr/testify/assert"
)

func TestEvalWithBuiltinContext(t *testing.T) {
	defer gostub.StubFunc(&logEvalResult).Reset()

	t.Run("ts GTE constraint matches", func(t *testing.T) {
		config.Config.InjectedContextEnabled = true
		defer func() { config.Config.InjectedContextEnabled = false }()

		s := entity.GenFixtureSegment()
		s.Constraints = []entity.Constraint{
			{
				Property: "@ts",
				Operator: models.ConstraintOperatorGTE,
				Value:    "0", // always matches (epoch 0)
			},
		}
		s.PrepareEvaluation()

		ctx := InjectBuiltInContext(map[string]any{}, nil)
		vID, _, evalNextSegment := evalSegment(models.EvalContext{
			EnableDebug:   true,
			EntityContext: ctx,
			EntityID:      "entity1",
			EntityType:    "entityType1",
			FlagID:        100,
		}, s)

		assert.NotNil(t, vID, "should match because ts >= 0 is always true")
		assert.False(t, evalNextSegment)
	})

	t.Run("ts GTE future timestamp does not match", func(t *testing.T) {
		config.Config.InjectedContextEnabled = true
		defer func() { config.Config.InjectedContextEnabled = false }()

		// Far future timestamp that will never match
		futureTS := time.Date(2099, 1, 1, 0, 0, 0, 0, time.UTC).Unix()

		s := entity.GenFixtureSegment()
		s.Constraints = []entity.Constraint{
			{
				Property: "@ts",
				Operator: models.ConstraintOperatorGTE,
				Value:    strconv.FormatInt(futureTS, 10),
			},
		}
		s.PrepareEvaluation()

		ctx := InjectBuiltInContext(map[string]any{}, nil)
		vID, _, evalNextSegment := evalSegment(models.EvalContext{
			EnableDebug:   true,
			EntityContext: ctx,
			EntityID:      "entity1",
			EntityType:    "entityType1",
			FlagID:        100,
		}, s)

		assert.Nil(t, vID, "should not match because ts < future timestamp")
		assert.True(t, evalNextSegment)
	})

	t.Run("ts_hour constraint matches business hours", func(t *testing.T) {
		config.Config.InjectedContextEnabled = true
		defer func() { config.Config.InjectedContextEnabled = false }()

		now := time.Now().UTC()
		s := entity.GenFixtureSegment()
		s.Constraints = []entity.Constraint{
			{
				Property: "@ts_hour",
				Operator: models.ConstraintOperatorGTE,
				Value:    "0", // always matches (hour >= 0)
			},
		}
		s.PrepareEvaluation()

		ctx := InjectBuiltInContext(map[string]any{}, nil)
		vID, _, evalNextSegment := evalSegment(models.EvalContext{
			EnableDebug:   true,
			EntityContext: ctx,
			EntityID:      "entity1",
			EntityType:    "entityType1",
			FlagID:        100,
		}, s)

		assert.NotNil(t, vID, "should match because ts_hour >= 0 is always true")
		assert.False(t, evalNextSegment)
		_ = now // used for documentation
	})

	t.Run("ts_weekday constraint matches weekday", func(t *testing.T) {
		config.Config.InjectedContextEnabled = true
		defer func() { config.Config.InjectedContextEnabled = false }()

		s := entity.GenFixtureSegment()
		s.Constraints = []entity.Constraint{
			{
				Property: "@ts_weekday",
				Operator: models.ConstraintOperatorGTE,
				Value:    "0", // always matches (day >= 0)
			},
		}
		s.PrepareEvaluation()

		ctx := InjectBuiltInContext(map[string]any{}, nil)
		vID, _, evalNextSegment := evalSegment(models.EvalContext{
			EnableDebug:   true,
			EntityContext: ctx,
			EntityID:      "entity1",
			EntityType:    "entityType1",
			FlagID:        100,
		}, s)

		assert.NotNil(t, vID, "should match because ts_weekday >= 0 is always true")
		assert.False(t, evalNextSegment)
	})

	t.Run("http_x_environment constraint matches", func(t *testing.T) {
		config.Config.InjectedContextEnabled = true
		config.Config.InjectedContextHTTPHeaders = []string{"X-Environment"}
		defer func() {
			config.Config.InjectedContextEnabled = false
			config.Config.InjectedContextHTTPHeaders = nil
		}()

		r := &http.Request{
			Header: http.Header{
				"X-Environment": []string{"production"},
			},
			Host: "flagr.example.com",
		}

		ctx := InjectBuiltInContext(map[string]any{}, r)

		s := entity.GenFixtureSegment()
		s.Constraints = []entity.Constraint{
			{
				Property: "@http_x_environment",
				Operator: models.ConstraintOperatorEQ,
				Value:    `"production"`,
			},
		}
		s.PrepareEvaluation()

		vID, _, evalNextSegment := evalSegment(models.EvalContext{
			EnableDebug:   true,
			EntityContext: ctx,
			EntityID:      "entity1",
			EntityType:    "entityType1",
			FlagID:        100,
		}, s)

		assert.NotNil(t, vID, "should match because http_x_environment == production")
		assert.False(t, evalNextSegment)
	})

	t.Run("http_x_environment constraint does not match", func(t *testing.T) {
		config.Config.InjectedContextEnabled = true
		config.Config.InjectedContextHTTPHeaders = []string{"X-Environment"}
		defer func() {
			config.Config.InjectedContextEnabled = false
			config.Config.InjectedContextHTTPHeaders = nil
		}()

		r := &http.Request{
			Header: http.Header{
				"X-Environment": []string{"staging"},
			},
			Host: "flagr.example.com",
		}

		ctx := InjectBuiltInContext(map[string]any{}, r)

		s := entity.GenFixtureSegment()
		s.Constraints = []entity.Constraint{
			{
				Property: "@http_x_environment",
				Operator: models.ConstraintOperatorEQ,
				Value:    `"production"`,
			},
		}
		s.PrepareEvaluation()

		vID, _, evalNextSegment := evalSegment(models.EvalContext{
			EnableDebug:   true,
			EntityContext: ctx,
			EntityID:      "entity1",
			EntityType:    "entityType1",
			FlagID:        100,
		}, s)

		assert.Nil(t, vID, "should not match because http_x_environment == staging, not production")
		assert.True(t, evalNextSegment)
	})

	t.Run("client context preserved alongside built-in keys", func(t *testing.T) {
		config.Config.InjectedContextEnabled = true
		config.Config.InjectedContextHTTPHeaders = []string{"X-Environment"}
		defer func() {
			config.Config.InjectedContextEnabled = false
			config.Config.InjectedContextHTTPHeaders = nil
		}()

		r := &http.Request{
			Header: http.Header{
				"X-Environment": []string{"production"},
			},
			Host: "flagr.example.com",
		}

		ctx := InjectBuiltInContext(map[string]any{"dl_state": "CA"}, r)

		s := entity.GenFixtureSegment()
		// Default fixture has constraint: dl_state EQ "CA"
		s.PrepareEvaluation()

		vID, _, evalNextSegment := evalSegment(models.EvalContext{
			EnableDebug:   true,
			EntityContext: ctx,
			EntityID:      "entity1",
			EntityType:    "entityType1",
			FlagID:        100,
		}, s)

		assert.NotNil(t, vID, "should match because client context dl_state=CA is preserved")
		assert.False(t, evalNextSegment)
	})

	t.Run("disabled injection does not affect eval", func(t *testing.T) {
		config.Config.InjectedContextEnabled = false

		s := entity.GenFixtureSegment()
		s.Constraints = []entity.Constraint{
			{
				Property: "@ts",
				Operator: models.ConstraintOperatorGTE,
				Value:    "0",
			},
		}
		s.PrepareEvaluation()

		vID, _, evalNextSegment := evalSegment(models.EvalContext{
			EnableDebug:   true,
			EntityContext: map[string]any{},
			EntityID:      "entity1",
			EntityType:    "entityType1",
			FlagID:        100,
		}, s)

		assert.Nil(t, vID, "should not match because injection is disabled, ts not in context")
		assert.True(t, evalNextSegment)
	})

	t.Run("business hours constraint", func(t *testing.T) {
		config.Config.InjectedContextEnabled = true
		defer func() { config.Config.InjectedContextEnabled = false }()

		s := entity.GenFixtureSegment()
		s.Constraints = []entity.Constraint{
			{
				Property: "@ts_hour",
				Operator: models.ConstraintOperatorGTE,
				Value:    "9",
			},
			{
				Property: "@ts_hour",
				Operator: models.ConstraintOperatorLT,
				Value:    "17",
			},
		}
		s.PrepareEvaluation()

		vID, _, evalNextSegment := evalSegment(models.EvalContext{
			EnableDebug:   true,
			EntityContext: map[string]any{},
			EntityID:      "entity1",
			EntityType:    "entityType1",
			FlagID:        100,
		}, s)

		hour := time.Now().UTC().Hour()
		if hour >= 9 && hour < 17 {
			assert.NotNil(t, vID, "should match during business hours (9-17 UTC)")
			assert.False(t, evalNextSegment)
		} else {
			assert.Nil(t, vID, "should not match outside business hours")
			assert.True(t, evalNextSegment)
		}
	})
}
