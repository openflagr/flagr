package handler

import (
	"encoding/json"

	"github.com/checkr/flagr/pkg/config"
	"github.com/checkr/flagr/pkg/util"
	"github.com/checkr/flagr/swagger_gen/models"

	"github.com/a8m/kinesis-producer"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/kinesis"
	"github.com/sirupsen/logrus"
)

var (
	newKinesisProducer = producer.New
)

type kinesisRecorder struct {
	enabled  bool
	producer *producer.Producer
}

// NewKinesisRecorder creates a new Kinesis recorder
var NewKinesisRecorder = func() DataRecorder {
	client := kinesis.New(session.New(aws.NewConfig()))

	p := newKinesisProducer(&producer.Config{
		StreamName:          config.Config.RecorderKinesisStreamName,
		Client:              client,
		BacklogCount:        config.Config.RecorderKinesisBacklogCount,
		MaxConnections:      config.Config.RecorderKinesisMaxConnections,
		FlushInterval:       config.Config.RecorderKinesisFlushInterval,
		BatchSize:           config.Config.RecorderKinesisBatchSize,
		BatchCount:          config.Config.RecorderKinesisBatchCount,
		AggregateBatchCount: config.Config.RecorderKinesisAggregateBatchCount,
		AggregateBatchSize:  config.Config.RecorderKinesisAggregateBatchSize,
		Verbose:             config.Config.RecorderKinesisVerbose,
		Logger:              logrus.WithField("producer", "kinesis"),
	})

	p.Start()

	go func() {
		for err := range p.NotifyFailures() {
			logrus.WithField("kinesis_error", err).Error("error pushing to kinesis")
		}
	}()

	return &kinesisRecorder{
		producer: p,
		enabled:  config.Config.RecorderEnabled,
	}
}

func (k *kinesisRecorder) AsyncRecord(r *models.EvalResult) {
	if !k.enabled {
		return
	}

	kr := &kinesisEvalResult{
		EvalResult: r,
	}

	payload, err := kr.Payload()
	if err != nil {
		logrus.WithField("kinesis_error", err).Error("error marshaling")
	}

	messageFrame := kinesisMessageFrame{
		Payload:   string(payload),
		Encrypted: false, // ignoring encryption at this time - https://github.com/checkr/flagr/pull/151#discussion_r208313230
	}

	message, err := messageFrame.encode()
	if err != nil {
		logrus.WithField("kinesis_error", err).Error("error marshaling")
	}

	err = k.producer.Put(message, kr.Key())
	if err != nil {
		logrus.WithField("kinesis_error", err).Error("error pushing to kinesis")
	}
}

type kinesisEvalResult struct {
	*models.EvalResult
}

type kinesisMessageFrame struct {
	Payload   string `json:"payload"`
	Encrypted bool   `json:"encrypted"`
}

func (kmf *kinesisMessageFrame) encode() ([]byte, error) {
	return json.MarshalIndent(kmf, "", "  ")
}

// Payload marshals the EvalResult
func (r *kinesisEvalResult) Payload() ([]byte, error) {
	return r.EvalResult.MarshalBinary()
}

// Key generates the partition key
func (r *kinesisEvalResult) Key() string {
	if r.EvalResult == nil || r.EvalContext == nil {
		return ""
	}
	return util.SafeString(r.EvalContext.EntityID)
}
