package handler

import (
	"testing"

	"github.com/a8m/kinesis-producer"
	"github.com/aws/aws-sdk-go/service/kinesis"

	"github.com/checkr/flagr/pkg/util"
	"github.com/checkr/flagr/swagger_gen/models"

	"github.com/stretchr/testify/assert"
)

type mockKinesisProducer struct {
	inputChan  chan *kinesis.PutRecordsRequestEntry
	errorsChan chan *producer.FailureRecord
}

func (m *mockKinesisProducer) Start()                                         {}
func (m *mockKinesisProducer) NotifyFailures() <-chan *producer.FailureRecord { return m.errorsChan }
func (m *mockKinesisProducer) Put(data []byte, partitionKey string) error     { return nil }

func TestKinesisMessageFrame(t *testing.T) {
	t.Run("happy code path - encrypted", func(t *testing.T) {
		kmf := kinesisMessageFrame{
			Payload:   "123",
			Encrypted: true,
		}
		encoded, err := kmf.encode()
		assert.NoError(t, err)
		assert.NotEmpty(t, encoded)
	})

	t.Run("happy code path - not encrypted", func(t *testing.T) {
		kmf := kinesisMessageFrame{
			Payload:   "456",
			Encrypted: false,
		}
		encoded, err := kmf.encode()
		assert.NoError(t, err)
		assert.NotEmpty(t, encoded)
	})
}

func TestKinesisEvalResult(t *testing.T) {
	t.Run("happy code path", func(t *testing.T) {
		r := &kinesisEvalResult{
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

		k := r.Key()
		assert.Equal(t, k, "d08042018")

		p, err := r.Payload()
		assert.Equal(t, err, nil)
		assert.Equal(t, "{\"evalContext\":{\"entityID\":\"d08042018\",\"entityType\":null},\"flagID\":1,\"flagKey\":\"random_flag_key\",\"flagSnapshotID\":1,\"segmentID\":1,\"timestamp\":null,\"variantAttachment\":null,\"variantID\":1,\"variantKey\":\"control\"}", string(p))
	})

	t.Run("empty EvalResult", func(t *testing.T) {
		r := &kinesisEvalResult{}

		assert.Zero(t, r.Key())
	})

	t.Run("empty Context", func(t *testing.T) {
		r := &kinesisEvalResult{}

		assert.Zero(t, r.Key())
	})
}

func TestNewKinesisRecorder(t *testing.T) {
	t.Run("no panics", func(t *testing.T) {
		assert.NotPanics(t, func() { NewKinesisRecorder() })
	})
}

func TestKinesisAsyncRecord(t *testing.T) {
	t.Run("not enabled", func(t *testing.T) {
		k := &kinesisRecorder{
			enabled: false,
		}

		k.AsyncRecord(nil)
	})

	t.Run("enabled and stream name is invalid", func(t *testing.T) {
		assert.Panics(t, func() {
			kr := &kinesisRecorder{
				producer: newKinesisProducer(&producer.Config{}),
				enabled:  true,
			}

			kr.AsyncRecord(&models.EvalResult{})
		})
	})

	t.Run("enabled and valid", func(t *testing.T) {
		assert.NotPanics(t, func() {
			kr := &kinesisRecorder{
				producer: newKinesisProducer(&producer.Config{StreamName: "hallo"}),
				enabled:  true,
			}

			kr.AsyncRecord(
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
