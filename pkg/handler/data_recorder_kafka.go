package handler

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strings"
	"time"

	"github.com/checkr/flagr/pkg/config"
	"github.com/checkr/flagr/pkg/util"
	"github.com/checkr/flagr/swagger_gen/models"

	"github.com/Shopify/sarama"
	"github.com/brandur/simplebox"
	"github.com/sirupsen/logrus"
)

var (
	saramaNewAsyncProducer = sarama.NewAsyncProducer
)

// NewKafkaRecorder creates a new Kafka recorder
var NewKafkaRecorder = func() DataRecorder {
	cfg := sarama.NewConfig()
	tlscfg := createTLSConfiguration(
		config.Config.RecorderKafkaCertFile,
		config.Config.RecorderKafkaKeyFile,
		config.Config.RecorderKafkaCAFile,
		config.Config.RecorderKafkaVerifySSL,
	)
	if tlscfg != nil {
		cfg.Net.TLS.Enable = true
		cfg.Net.TLS.Config = tlscfg
	}
	cfg.Producer.RequiredAcks = sarama.WaitForLocal
	cfg.Producer.Retry.Max = config.Config.RecorderKafkaRetryMax
	cfg.Producer.Flush.Frequency = config.Config.RecorderKafkaFlushFrequency

	brokerList := strings.Split(config.Config.RecorderKafkaBrokers, ",")
	producer, err := saramaNewAsyncProducer(brokerList, cfg)
	if err != nil {
		logrus.WithField("kafka_error", err).Fatal("Failed to start Sarama producer:")
	}

	// We will just log to STDOUT if we're not able to produce messages.
	if producer != nil {
		go func() {
			for err := range producer.Errors() {
				logrus.WithField("kafka_error", err).Error("failed to write access log entry")
			}
		}()
	}

	return &kafkaRecorder{
		producer: producer,
		topic:    config.Config.RecorderKafkaTopic,
		enabled:  config.Config.RecorderEnabled,
	}
}

func createTLSConfiguration(certFile string, keyFile string, caFile string, verifySSL bool) (t *tls.Config) {
	if certFile != "" && keyFile != "" && caFile != "" {
		cert, err := tls.LoadX509KeyPair(certFile, keyFile)
		if err != nil {
			logrus.WithField("TLSConfigurationError", err).Panic(err)
		}

		caCert, err := ioutil.ReadFile(caFile)
		if err != nil {
			logrus.WithField("TLSConfigurationError", err).Panic(err)
		}

		caCertPool := x509.NewCertPool()
		caCertPool.AppendCertsFromPEM(caCert)

		t = &tls.Config{
			Certificates:       []tls.Certificate{cert},
			RootCAs:            caCertPool,
			InsecureSkipVerify: !verifySSL,
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

func (k *kafkaRecorder) AsyncRecord(r *models.EvalResult) {
	if !k.enabled {
		return
	}
	kr := &kafkaEvalResult{
		EvalResult: r,
		encrypted:  config.Config.RecorderKafkaEncrypted,
	}
	k.producer.Input() <- &sarama.ProducerMessage{
		Topic:     k.topic,
		Key:       sarama.StringEncoder(kr.Key()),
		Value:     kr,
		Timestamp: time.Now().UTC(),
	}

	logKafkaAsyncRecordToDatadog(r)
}

var logKafkaAsyncRecordToDatadog = func(r *models.EvalResult) {
	if config.Global.StatsdClient == nil {
		return
	}
	config.Global.StatsdClient.Incr(
		"data_recorder.kafka",
		[]string{
			fmt.Sprintf("FlagID:%d", util.SafeUint(r.FlagID)),
		},
		float64(1),
	)
}

type kafkaEvalResult struct {
	*models.EvalResult

	encrypted bool
	encoded   []byte
	err       error
}

func (r *kafkaEvalResult) ensureEncoded() {
	if r.encoded == nil && r.err == nil {
		payload, err := r.EvalResult.MarshalBinary()
		if err != nil {
			r.err = err
			return
		}
		kmf := &kafkaMessageFrame{
			Payload:   string(payload),
			Encrypted: r.encrypted,
		}
		r.encoded, r.err = kmf.encode(config.Config.RecorderKafkaEncryptionKey)
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
	if r.EvalResult == nil || r.EvalContext == nil {
		return ""
	}
	return util.SafeString(r.EvalContext.EntityID)
}

type kafkaMessageFrame struct {
	Payload   string `json:"payload"`
	Encrypted bool   `json:"encrypted"`
}

func (kmf *kafkaMessageFrame) encode(k string) ([]byte, error) {
	if !kmf.Encrypted {
		return json.MarshalIndent(kmf, "", "  ")
	}

	key := [simplebox.KeySize]byte{}
	copy(key[:], k)
	kmf.Payload = base64.StdEncoding.EncodeToString(simplebox.NewFromSecretKey(&key).Encrypt([]byte(kmf.Payload)))
	return json.MarshalIndent(kmf, "", "  ")
}
