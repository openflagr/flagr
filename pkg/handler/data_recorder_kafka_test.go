package handler

import (
	"testing"

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
