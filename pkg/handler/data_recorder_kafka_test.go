package handler

import (
	"strings"
	"testing"

	"github.com/checkr/flagr/pkg/config"
	"github.com/checkr/flagr/pkg/util"
	"github.com/checkr/flagr/swagger_gen/models"

	"github.com/Shopify/sarama"
	"github.com/stretchr/testify/assert"
)

type Closer interface {
	Close() error
}

func close(t *testing.T, c Closer) {
	assert := assert.New(t)
	err := c.Close()
	assert.NoError(err)
}

func TestKafkaRecorder(t *testing.T) {
	t.Run("happy code path - record", func(t *testing.T) {
		kr := NewKafkaRecorder()
		kr.AsyncRecord(&models.EvalResult{
			FlagID: util.Int64Ptr(int64(100)),
		})
		defer close(t, kr)
		brokerList := strings.Split(config.Config.RecorderKafkaBrokers, ",")
		cfg := sarama.NewConfig()
		consumer, err := sarama.NewConsumer(brokerList, cfg)
		assert.NoError(t, err)
		defer close(t, consumer)
		partitionConsumer, err := consumer.ConsumePartition(config.Config.RecorderKafkaTopic, 0, 0)
		assert.NoError(t, err)
		defer close(t, partitionConsumer)

		select {
		case msg := <-partitionConsumer.Messages():
			assert.Equal(t, "what", msg.Topic)
		}
	})
}

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
