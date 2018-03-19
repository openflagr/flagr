package handler

import (
	"testing"

	"github.com/checkr/flagr/pkg/util"
	"github.com/checkr/flagr/swagger_gen/models"
	"github.com/prashantv/gostub"
	"github.com/stretchr/testify/assert"
)

func TestKafkaMessageFrame(t *testing.T) {
	t.Run("happy code path - encrypted", func(t *testing.T) {
		kmf := kafkaMessageFrame{
			Payload:   "123",
			Encrypted: true,
		}
		encoded, err := kmf.encode("o7hAxo52oOl7cmyq/X0UkJ3VMmIo5aAv")
		assert.NoError(t, err)
		assert.NotEmpty(t, encoded)
	})

	t.Run("happy code path - not encrypted", func(t *testing.T) {
		kmf := kafkaMessageFrame{
			Payload:   "456",
			Encrypted: false,
		}
		encoded, err := kmf.encode("o7hAxo52oOl7cmyq/X0UkJ3VMmIo5aAv")
		assert.NoError(t, err)
		assert.NotEmpty(t, encoded)
	})
}

func TestNewKafkaRecorder(t *testing.T) {
	defer gostub.StubFunc(&saramaNewAsyncProducer, nil, nil).Reset()
	assert.NotPanics(t, func() { NewKafkaRecorder() })
}

func TestCreateTLSConfiguration(t *testing.T) {
	t.Run("happy code path", func(t *testing.T) {
		tlsConfig := createTLSConfiguration(
			"./testdata/certificates/alice.crt",
			"./testdata/certificates/alice.key",
			"./testdata/certificates/ca.crt",
			true,
		)
		assert.NotZero(t, tlsConfig)

		tlsConfig = createTLSConfiguration(
			"",
			"",
			"",
			true,
		)
		assert.Zero(t, tlsConfig)
	})

	t.Run("cert or key file not found", func(t *testing.T) {
		assert.Panics(t, func() {
			createTLSConfiguration(
				"./testdata/certificates/not_found.crt",
				"./testdata/certificates/not_found.key",
				"./testdata/certificates/ca.crt",
				true,
			)
		})
	})

	t.Run("ca file not found", func(t *testing.T) {
		assert.Panics(t, func() {
			createTLSConfiguration(
				"./testdata/certificates/alice.crt",
				"./testdata/certificates/alice.key",
				"./testdata/certificates/not_found.crt",
				true,
			)
		})
	})
}

func TestKafkaEvalResult(t *testing.T) {
	t.Run("happy code path", func(t *testing.T) {
		r := &kafkaEvalResult{
			EvalResult: &models.EvalResult{
				EvalContext: &models.EvalContext{
					EntityID: util.StringPtr("123"),
				},
				FlagID:         util.Int64Ptr(int64(1)),
				FlagSnapshotID: 1,
				SegmentID:      util.Int64Ptr(int64(1)),
				VariantID:      util.Int64Ptr(int64(1)),
				VariantKey:     util.StringPtr("control"),
			},
			encrypted: false,
			encoded:   nil,
			err:       nil,
		}

		b, err := r.Encode()
		assert.NoError(t, err)
		assert.NotZero(t, len(b))

		l := r.Length()
		assert.NotZero(t, l)

		k := r.Key()
		assert.Equal(t, k, "123")
	})

	t.Run("empty EvalResult", func(t *testing.T) {
		r := &kafkaEvalResult{}
		assert.Zero(t, r.Key())
	})
}
