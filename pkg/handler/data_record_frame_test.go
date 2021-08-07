package handler

import (
	"testing"

	"github.com/openflagr/flagr/swagger_gen/models"
	"github.com/stretchr/testify/assert"
)

func TestFrameOutput(t *testing.T) {
	er := models.EvalResult{
		EvalContext: &models.EvalContext{
			EntityID: "123",
		},
		FlagID:         1,
		FlagSnapshotID: 1,
		SegmentID:      1,
		VariantID:      1,
		VariantKey:     "control",
	}

	t.Run("empty options", func(t *testing.T) {
		frame := DataRecordFrame{
			evalResult: er,
			options:    DataRecordFrameOptions{},
		}
		output, err := frame.Output()
		assert.NoError(t, err)
		assert.Contains(t, string(output), "123")
		assert.Contains(t, string(output), "payload")
	})

	t.Run("payload with encryption options", func(t *testing.T) {
		frame := DataRecordFrame{
			evalResult: er,
			options: DataRecordFrameOptions{
				Encrypted:       true,
				Encryptor:       newSimpleboxEncryptor("fake_key"),
				FrameOutputMode: "payload",
			},
		}
		output, err := frame.Output()
		assert.NoError(t, err)
		assert.Contains(t, string(output), "payload")
	})

	t.Run("payload_raw_json options", func(t *testing.T) {
		frame := DataRecordFrame{
			evalResult: er,
			options: DataRecordFrameOptions{
				Encrypted:       false,
				Encryptor:       nil,
				FrameOutputMode: frameOutputModePayloadRawJSON,
			},
		}
		output, err := frame.Output()
		assert.NoError(t, err)
		assert.Contains(t, string(output), "123")
		assert.Contains(t, string(output), "payload")
		assert.NotContains(t, string(output), `"payload":""`)
	})
}

func TestGetPartitionKey(t *testing.T) {

	t.Run("empty evalResult", func(t *testing.T) {
		er := models.EvalResult{}
		frame := DataRecordFrame{evalResult: er}
		assert.Equal(t, "", frame.GetPartitionKey())
	})

	t.Run("happy code path evalResult", func(t *testing.T) {
		er := models.EvalResult{
			EvalContext: &models.EvalContext{
				EntityID: "123",
			},
			FlagID:         1,
			FlagSnapshotID: 1,
			SegmentID:      1,
			VariantID:      1,
			VariantKey:     "control",
		}
		frame := DataRecordFrame{evalResult: er}
		assert.Equal(t, "123", frame.GetPartitionKey())
	})

}
