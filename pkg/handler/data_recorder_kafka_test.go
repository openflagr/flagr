package handler

import (
	"testing"

	"github.com/Shopify/sarama"
	"github.com/openflagr/flagr/swagger_gen/models"
	"github.com/prashantv/gostub"
	"github.com/stretchr/testify/assert"
)

type mockAsyncProducer struct {
	inputCh   chan *sarama.ProducerMessage
	successCh chan *sarama.ProducerMessage
	errorCh   chan *sarama.ProducerError
}

func (m *mockAsyncProducer) AsyncClose()                               {}
func (m *mockAsyncProducer) Close() error                              { return nil }
func (m *mockAsyncProducer) Input() chan<- *sarama.ProducerMessage     { return m.inputCh }
func (m *mockAsyncProducer) Successes() <-chan *sarama.ProducerMessage { return m.successCh }
func (m *mockAsyncProducer) Errors() <-chan *sarama.ProducerError      { return m.errorCh }
func (m *mockAsyncProducer) IsTransactional() bool                     { return false }
func (m *mockAsyncProducer) TxnStatus() sarama.ProducerTxnStatusFlag   { return 0 }
func (m *mockAsyncProducer) BeginTxn() error                           { return nil }
func (m *mockAsyncProducer) CommitTxn() error                          { return nil }
func (m *mockAsyncProducer) AbortTxn() error                           { return nil }
func (m *mockAsyncProducer) AddOffsetsToTxn(offsets map[string][]*sarama.PartitionOffsetMetadata, groupId string) error {
	return nil
}
func (m *mockAsyncProducer) AddMessageToTxn(msg *sarama.ConsumerMessage, groupId string, metadata *string) error {
	return nil
}

var _ sarama.AsyncProducer = (*mockAsyncProducer)(nil)

func TestNewKafkaRecorder(t *testing.T) {
	t.Run("no panics", func(t *testing.T) {
		defer gostub.StubFunc(
			&saramaNewAsyncProducer,
			&mockAsyncProducer{
				inputCh: make(chan *sarama.ProducerMessage),
			},
			nil,
		).Reset()

		assert.NotPanics(t, func() { NewKafkaRecorder() })
	})
}

func TestCreateTLSConfiguration(t *testing.T) {
	t.Run("happy code path", func(t *testing.T) {
		tlsConfig := createTLSConfiguration(
			"./testdata/certificates/alice.crt",
			"./testdata/certificates/alice.key",
			"./testdata/certificates/ca.crt",
			true,
			false,
		)
		assert.NotZero(t, tlsConfig)

		tlsConfig = createTLSConfiguration(
			"",
			"",
			"",
			true,
			false,
		)
		assert.Zero(t, tlsConfig)

		tlsConfig = createTLSConfiguration(
			"",
			"",
			"",
			true,
			true,
		)
		assert.NotZero(t, tlsConfig)
	})

	t.Run("cert or key file not found", func(t *testing.T) {
		assert.Panics(t, func() {
			createTLSConfiguration(
				"./testdata/certificates/not_found.crt",
				"./testdata/certificates/not_found.key",
				"./testdata/certificates/ca.crt",
				true,
				false,
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
				false,
			)
		})
	})
}

func TestAsyncRecord(t *testing.T) {
	t.Run("happy code path", func(t *testing.T) {
		p := &mockAsyncProducer{inputCh: make(chan *sarama.ProducerMessage)}
		kr := &kafkaRecorder{
			producer: p,
			topic:    "test-topic",
		}

		go kr.AsyncRecord(models.EvalResult{})
		r := <-p.inputCh
		assert.NotNil(t, r)
	})
}

func TestMustParseKafkaVersion(t *testing.T) {
	assert.NotPanics(t, func() {
		mustParseKafkaVersion("0.8.2.0")
		mustParseKafkaVersion("1.1.0") // for version >1.0, use 3 numbers
	})

	assert.Panics(t, func() {
		mustParseKafkaVersion("1.1.0.0") // for version >1.0, use 3 numbers
	})
}
