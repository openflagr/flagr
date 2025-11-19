package handler

import (
	"context"

	producer "github.com/a8m/kinesis-producer"
	"github.com/a8m/kinesis-producer/loggers/kplogrus"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/kinesis"
	flagrConfig "github.com/openflagr/flagr/pkg/config"
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
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		logrus.WithField("kinesis_error", err).Fatal("error creating aws session")
	}

	client := kinesis.NewFromConfig(cfg)

	p := newKinesisProducer(&producer.Config{
		StreamName:          flagrConfig.Config.RecorderKinesisStreamName,
		Client:              client,
		BacklogCount:        flagrConfig.Config.RecorderKinesisBacklogCount,
		MaxConnections:      flagrConfig.Config.RecorderKinesisMaxConnections,
		FlushInterval:       flagrConfig.Config.RecorderKinesisFlushInterval,
		BatchSize:           flagrConfig.Config.RecorderKinesisBatchSize,
		BatchCount:          flagrConfig.Config.RecorderKinesisBatchCount,
		AggregateBatchCount: flagrConfig.Config.RecorderKinesisAggregateBatchCount,
		AggregateBatchSize:  flagrConfig.Config.RecorderKinesisAggregateBatchSize,
		Verbose:             flagrConfig.Config.RecorderKinesisVerbose,
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
			FrameOutputMode: flagrConfig.Config.RecorderFrameOutputMode,
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
