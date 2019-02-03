package handler

import (
	"context"
	"testing"

	"cloud.google.com/go/pubsub"

	"github.com/checkr/flagr/pkg/util"
	"github.com/checkr/flagr/swagger_gen/models"

	"github.com/stretchr/testify/assert"
)

func TestPubsubMessageFrame(t *testing.T) {
	t.Run("happy code path - encrypted", func(t *testing.T) {
		pmf := pubsubMessageFrame{
			Payload:   "123",
			Encrypted: true,
		}
		encoded, err := pmf.encode()
		assert.NoError(t, err)
		assert.NotEmpty(t, encoded)
	})

	t.Run("happy code path - not encrypted", func(t *testing.T) {
		pmf := pubsubMessageFrame{
			Payload:   "456",
			Encrypted: false,
		}
		encoded, err := pmf.encode()
		assert.NoError(t, err)
		assert.NotEmpty(t, encoded)
	})
}

func TestPubsubEvalResult(t *testing.T) {
	t.Run("happy code path", func(t *testing.T) {
		r := &pubsubEvalResult{
			EvalResult: &models.EvalResult{
				EvalContext: &models.EvalContext{
					EntityID: "d08042018",
				},
				FlagID:         util.Int64Ptr(int64(1)),
				FlagKey:        util.StringPtr("random_flag_key"),
				FlagSnapshotID: 1,
				SegmentID:      util.Int64Ptr(int64(1)),
				VariantID:      util.Int64Ptr(int64(1)),
				VariantKey:     util.StringPtr("control"),
			},
		}

		p, err := r.Payload()
		assert.NoError(t, err)
		assert.NotEmpty(t, p)
		assert.Regexp(t, "d08042018", string(p))
	})

	t.Run("empty EvalResult", func(t *testing.T) {
		r := &pubsubEvalResult{}
		assert.Zero(t, r.EvalResult)
	})
}

func TestNewPubsubRecorder(t *testing.T) {
	t.Run("no panics", func(t *testing.T) {
		assert.NotPanics(t, func() { NewPubsubRecorder() })
	})
}

func TestPubsubAsyncRecord(t *testing.T) {
	t.Run("not enabled", func(t *testing.T) {
		p := &pubsubRecorder{
			enabled: false,
		}

		p.AsyncRecord(nil)
	})

	t.Run("enabled and valid", func(t *testing.T) {
		client, err := pubsub.NewClient(context.Background(), "")
		assert.NoError(t, err)
		topic := client.Topic("test")
		assert.NotPanics(t, func() {
			pr := &pubsubRecorder{
				producer: client,
				topic:    topic,
				enabled:  true,
			}

			pr.AsyncRecord(
				&models.EvalResult{
					EvalContext: &models.EvalContext{
						EntityID: "d08042018",
					},
					FlagID:         util.Int64Ptr(int64(1)),
					FlagSnapshotID: 1,
					SegmentID:      util.Int64Ptr(int64(1)),
					VariantID:      util.Int64Ptr(int64(1)),
					VariantKey:     util.StringPtr("control"),
				},
			)
		})
	})
}
