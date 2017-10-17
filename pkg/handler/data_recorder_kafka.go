package handler

import (
	"crypto/tls"
	"crypto/x509"
	"io/ioutil"
	"log"
	"strings"
	"time"

	"github.com/checkr/flagr/pkg/config"
	"github.com/checkr/flagr/pkg/util"
	"github.com/checkr/flagr/swagger_gen/models"

	"github.com/Shopify/sarama"
)

func createTLSConfiguration(certFile string, keyFile string, caFile string, verifySSL bool) (t *tls.Config) {
	if certFile != "" && keyFile != "" && caFile != "" {
		cert, err := tls.LoadX509KeyPair(certFile, keyFile)
		if err != nil {
			log.Fatal(err)
		}

		caCert, err := ioutil.ReadFile(caFile)
		if err != nil {
			log.Fatal(err)
		}

		caCertPool := x509.NewCertPool()
		caCertPool.AppendCertsFromPEM(caCert)

		t = &tls.Config{
			Certificates:       []tls.Certificate{cert},
			RootCAs:            caCertPool,
			InsecureSkipVerify: verifySSL,
		}
	}
	// will be nil by default if nothing is provided
	return t
}

type kafkaRecorder struct {
	producer sarama.AsyncProducer
	topic    string
	enabled  bool
}

// NewKafkaRecorder creates a new Kafka recorder
func NewKafkaRecorder() DataRecorder {
	cfg := sarama.NewConfig()
	tlscfg := createTLSConfiguration(
		config.Config.Recorder.Kafka.CertFile,
		config.Config.Recorder.Kafka.KeyFile,
		config.Config.Recorder.Kafka.CAFile,
		config.Config.Recorder.Kafka.VerifySSL,
	)
	if tlscfg != nil {
		cfg.Net.TLS.Enable = true
		cfg.Net.TLS.Config = tlscfg
	}
	cfg.Producer.RequiredAcks = sarama.WaitForLocal
	cfg.Producer.Retry.Max = config.Config.Recorder.Kafka.RetryMax
	cfg.Producer.Flush.Frequency = config.Config.Recorder.Kafka.FlushFrequency.Duration

	brokerList := strings.Split(config.Config.Recorder.Kafka.Brokers, ",")
	producer, err := sarama.NewAsyncProducer(brokerList, cfg)
	if err != nil {
		log.Fatalln("Failed to start Sarama producer:", err)
	}

	// We will just log to STDOUT if we're not able to produce messages.
	go func() {
		for err := range producer.Errors() {
			log.Println("Failed to write access log entry:", err)
		}
	}()

	return &kafkaRecorder{
		producer: producer,
		topic:    config.Config.Recorder.Kafka.Topic,
		enabled:  config.Config.Recorder.Enabled,
	}
}

func (k *kafkaRecorder) AsyncRecord(r *models.EvalResult) {
	if !k.enabled {
		return
	}
	kr := &kafkaEvalResult{EvalResult: r}
	k.producer.Input() <- &sarama.ProducerMessage{
		Topic:     k.topic,
		Key:       sarama.StringEncoder(kr.Key()),
		Value:     kr,
		Timestamp: time.Now().UTC(),
	}
}

type kafkaEvalResult struct {
	*models.EvalResult

	encoded []byte
	err     error
}

func (r *kafkaEvalResult) ensureEncoded() {
	if r.encoded == nil && r.err == nil {
		r.encoded, r.err = r.MarshalBinary()
	}
}

func (r *kafkaEvalResult) Encode() ([]byte, error) {
	r.ensureEncoded()
	return r.encoded, r.err
}

func (r *kafkaEvalResult) Length() int {
	r.ensureEncoded()
	return len(r.encoded)
}

// Key generates the partition key
func (r *kafkaEvalResult) Key() string {
	if r.EvalResult == nil {
		return ""
	}
	return util.SafeString(r.EvalContext.EntityID)
}
