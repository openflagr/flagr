package handler

import (
	producer "github.com/a8m/kinesis-producer"
	"github.com/a8m/kinesis-producer/loggers/kplogrus"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/kinesis"
	"github.com/openflagr/flagr/pkg/config"
	"github.com/openflagr/flagr/swagger_gen/models"
	"github.com/sirupsen/logrus"
)

var (
	newKinesisProducer = producer.New
)

type kinesisRecorder struct {
	producer *producer.Producer
	options  DataRecordFrameOptions
}

// NewKinesisRecorder creates a new Kinesis recorder
var NewKinesisRecorder = func() DataRecorder {
	se, err := session.NewSession(aws.NewConfig())
	if err != nil {
		logrus.WithField("kinesis_error", err).Fatal("error creating aws session")
	}

	client := kinesis.New(se)

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
		Logger:              &kplogrus.Logger{Logger: logrus.StandardLogger()},
	})

	p.Start()

	go func() {
		for err := range p.NotifyFailures() {
			logrus.WithField("kinesis_error", err).Error("error pushing to kinesis")
		}
	}()

	return &kinesisRecorder{
		producer: p,
		options: DataRecordFrameOptions{
			Encrypted:       false, // not implemented yet
			FrameOutputMode: config.Config.RecorderFrameOutputMode,
		},
	}
}

func (k *kinesisRecorder) NewDataRecordFrame(r models.EvalResult) DataRecordFrame {
	return DataRecordFrame{
		evalResult: r,
		options:    k.options,
	}
}

func (k *kinesisRecorder) AsyncRecord(r models.EvalResult) {
	frame := k.NewDataRecordFrame(r)
	output, err := frame.Output()
	if err != nil {
		logrus.WithField("err", err).Error("failed to generate data record frame for kinesis recorder")
		return
	}
	err = k.producer.Put(output, frame.GetPartitionKey())
	if err != nil {
		logrus.WithField("kinesis_error", err).Error("error pushing to kinesis")
	}
}
