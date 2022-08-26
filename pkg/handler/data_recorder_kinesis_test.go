package handler

import (
	"testing"

	producer "github.com/a8m/kinesis-producer"
	"github.com/openflagr/flagr/swagger_gen/models"
	"github.com/stretchr/testify/assert"
)

func TestNewKinesisRecorder(t *testing.T) {
	t.Run("no panics", func(t *testing.T) {
		assert.NotPanics(t, func() { NewKinesisRecorder() })
	})
}

func TestKinesisAsyncRecord(t *testing.T) {
	t.Run("invalid stream name", func(t *testing.T) {
		assert.Panics(t, func() {
			kr := &kinesisRecorder{
				producer: newKinesisProducer(&producer.Config{}),
			}

			kr.AsyncRecord(models.EvalResult{})
		})
	})

	t.Run("valid stream name", func(t *testing.T) {
		assert.NotPanics(t, func() {
			kr := &kinesisRecorder{
				producer: newKinesisProducer(&producer.Config{StreamName: "hallo"}),
			}

			kr.AsyncRecord(
				models.EvalResult{
					EvalContext: &models.EvalContext{
						EntityID: "d08042018",
					},
					FlagID:         1,
					FlagSnapshotID: 1,
					SegmentID:      1,
					VariantID:      1,
					VariantKey:     "control",
				},
			)
		})
	})
}
