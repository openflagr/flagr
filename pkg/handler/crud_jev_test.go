package handler

import (
	"testing"

	"github.com/openflagr/flagr/swagger_gen/models"
	"github.com/openflagr/flagr/swagger_gen/restapi/operations/constraint"
	"github.com/openflagr/flagr/swagger_gen/restapi/operations/flag"
	"github.com/openflagr/flagr/swagger_gen/restapi/operations/segment"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func jevCrudFixture(t *testing.T) *crud {
	t.Helper()
	_, cleanup := handlerTestDB(t)
	t.Cleanup(cleanup)

	c := &crud{}
	res := c.CreateFlag(flag.CreateFlagParams{Body: &models.CreateFlagRequest{Description: new("jev flag")}})
	require.IsType(t, &flag.CreateFlagOK{}, res)
	res = c.CreateSegment(segment.CreateSegmentParams{
		FlagID: int64(1),
		Body:   &models.CreateSegmentRequest{Description: new("segment1"), RolloutPercent: new(int64(100))},
	})
	require.IsType(t, &segment.CreateSegmentOK{}, res)
	return c
}

func TestCrudJevConstraintRoundTrip(t *testing.T) {
	c := jevCrudFixture(t)

	questionType := "choice"
	created := c.CreateConstraint(constraint.CreateConstraintParams{
		FlagID:    int64(1),
		SegmentID: int64(1),
		Body: &models.CreateConstraintRequest{
			Property: new("@jev.plan_tier"),
			Operator: new("EQ"),
			Value:    new(`"pro"`),
			Jev: &models.JevQuestion{
				Type:         &questionType,
				Instructions: "Which plan should this account see?",
				Criteria:     map[string]any{"free": "self-serve", "pro": "team usage"},
			},
		},
	})
	createdOK, ok := created.(*constraint.CreateConstraintOK)
	require.True(t, ok, "expected a 200 response, got %T", created)
	require.NotNil(t, createdOK.Payload.Jev)
	assert.Equal(t, "choice", *createdOK.Payload.Jev.Type)
	assert.Equal(t, "Which plan should this account see?", createdOK.Payload.Jev.Instructions)

	found := c.FindConstraints(constraint.FindConstraintsParams{FlagID: int64(1), SegmentID: int64(1)})
	foundOK, ok := found.(*constraint.FindConstraintsOK)
	require.True(t, ok)
	require.Len(t, foundOK.Payload, 1)
	require.NotNil(t, foundOK.Payload[0].Jev)
	assert.Equal(t, "@jev.plan_tier", *foundOK.Payload[0].Property)

	// Updating without a jev body clears the question (PUT replaces).
	updated := c.PutConstraint(constraint.PutConstraintParams{
		FlagID:       int64(1),
		SegmentID:    int64(1),
		ConstraintID: int64(1),
		Body: &models.CreateConstraintRequest{
			Property: new("dl_state"),
			Operator: new("EQ"),
			Value:    new(`"CA"`),
		},
	})
	updatedOK, ok := updated.(*constraint.PutConstraintOK)
	require.True(t, ok, "expected a 200 response, got %T", updated)
	assert.Nil(t, updatedOK.Payload.Jev)
}

func TestCrudJevConstraintRejectsInvalid(t *testing.T) {
	c := jevCrudFixture(t)

	questionType := "choice"
	res := c.CreateConstraint(constraint.CreateConstraintParams{
		FlagID:    int64(1),
		SegmentID: int64(1),
		Body: &models.CreateConstraintRequest{
			Property: new("@jev.plan_tier"),
			Operator: new("EQ"),
			Value:    new(`"pro"`),
			Jev: &models.JevQuestion{
				Type:         &questionType,
				Instructions: "Which plan?",
				// criteria intentionally missing
			},
		},
	})
	defaultResp, ok := res.(*constraint.CreateConstraintDefault)
	require.True(t, ok, "expected a 400 response, got %T", res)
	assert.Contains(t, *defaultResp.Payload.Message, "criteria")
}

func TestCrudJevConstraintRejectsNonJevProperty(t *testing.T) {
	c := jevCrudFixture(t)

	questionType := "noul"
	res := c.CreateConstraint(constraint.CreateConstraintParams{
		FlagID:    int64(1),
		SegmentID: int64(1),
		Body: &models.CreateConstraintRequest{
			Property: new("plan"),
			Operator: new("GTE"),
			Value:    new("0.8"),
			Jev: &models.JevQuestion{
				Type:         &questionType,
				Instructions: "Is this intent?",
			},
		},
	})
	assert.IsType(t, &constraint.CreateConstraintDefault{}, res)
}
