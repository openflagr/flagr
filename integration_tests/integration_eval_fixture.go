//go:build integration

package flagr_integration

import (
	"fmt"
	"testing"
	"time"
)

// evalConstraintFixture describes one segment constraint for eval-ready flag setup.
type evalConstraintFixture struct {
	Property string
	Operator string
	Value    string
}

// evalReadyFlag is a flag enabled for evaluation with one constrained segment and distribution.
type evalReadyFlag struct {
	FlagID int64
}

// createEvalReadyFlag creates a new flag, enables it, adds variant/segment/constraint/distribution,
// and waits for the eval cache. Caller must delete the flag via deleteEvalReadyFlag.
func createEvalReadyFlag(
	t *testing.T,
	key string,
	description string,
	constraint evalConstraintFixture,
	variantKey string,
	segmentDescription string,
) evalReadyFlag {
	t.Helper()

	var created flagResponse
	postJSON(t, "/api/v1/flags", map[string]any{
		"key":         key,
		"description": description,
	}, &created)
	if created.ID == 0 {
		t.Fatal("expected non-zero flag id")
	}

	putJSON(t, fmt.Sprintf("/api/v1/flags/%d/enabled", created.ID), map[string]any{
		"enabled": true,
	}, nil)

	var variant variantResponse
	postJSON(t, fmt.Sprintf("/api/v1/flags/%d/variants", created.ID), map[string]any{
		"key": variantKey,
	}, &variant)
	if variant.ID == 0 {
		t.Fatal("expected non-zero variant id")
	}

	var segment segmentResponse
	postJSON(t, fmt.Sprintf("/api/v1/flags/%d/segments", created.ID), map[string]any{
		"description":    segmentDescription,
		"rolloutPercent": builtinCtxRolloutPercent,
	}, &segment)
	if segment.ID == 0 {
		t.Fatal("expected non-zero segment id")
	}

	var c constraintResponse
	postJSON(t, fmt.Sprintf("/api/v1/flags/%d/segments/%d/constraints", created.ID, segment.ID), map[string]any{
		"property": constraint.Property,
		"operator": constraint.Operator,
		"value":    constraint.Value,
	}, &c)
	if c.ID == 0 {
		t.Fatal("expected non-zero constraint id")
	}

	putJSON(t, fmt.Sprintf("/api/v1/flags/%d/segments/%d/distributions", created.ID, segment.ID), map[string]any{
		"distributions": []map[string]any{
			{
				"variantID":  variant.ID,
				"variantKey": variantKey,
				"percent":    builtinCtxRolloutPercent,
			},
		},
	}, nil)

	time.Sleep(builtinCtxEvalCacheWait)

	return evalReadyFlag{FlagID: created.ID}
}

func deleteEvalReadyFlag(t *testing.T, flagID int64) {
	t.Helper()
	deleteResource(t, fmt.Sprintf("/api/v1/flags/%d", flagID))
}

// postEval posts a single-flag evaluation and decodes the result.
func postEval(t *testing.T, flagID int64, entityID string, entityContext map[string]any) evalResponse {
	t.Helper()
	if entityContext == nil {
		entityContext = map[string]any{}
	}
	var result evalResponse
	postJSON(t, "/api/v1/evaluation", map[string]any{
		"flagID":        flagID,
		"entityID":      entityID,
		"entityType":    "user",
		"entityContext": entityContext,
	}, &result)
	return result
}
