package handler

import (
	"context"

	"cloud.google.com/go/pubsub"
	"github.com/openflagr/flagr/pkg/config"
	"github.com/openflagr/flagr/swagger_gen/models"
	"github.com/sirupsen/logrus"
	"google.golang.org/api/option"
)

type pubsubRecorder struct {
	producer *pubsub.Client
	topic    *pubsub.Topic
	options  DataRecordFrameOptions
}

var (
	pubsubClient = func() (*pubsub.Client, error) {
		return pubsub.NewClient(
			context.Background(),
			config.Config.RecorderPubsubProjectID,
			option.WithCredentialsFile(config.Config.RecorderPubsubKeyFile),
		)
	}
)

// NewPubsubRecorder creates a new Pubsub recorder
var NewPubsubRecorder = func() DataRecorder {
	client, err := pubsubClient()
	if err != nil {
		logrus.WithField("pubsub_error", err).Fatal("error getting pubsub client")
	}

	return &pubsubRecorder{
		producer: client,
		topic:    client.Topic(config.Config.RecorderPubsubTopicName),
		options: DataRecordFrameOptions{
			Encrypted:       false, // not implemented yet
			FrameOutputMode: config.Config.RecorderFrameOutputMode,
		},
	}
}

func (p *pubsubRecorder) NewDataRecordFrame(r models.EvalResult) DataRecordFrame {
	return DataRecordFrame{
		evalResult: r,
		options:    p.options,
	}
}

func (p *pubsubRecorder) AsyncRecord(r models.EvalResult) {
	frame := p.NewDataRecordFrame(r)
	output, err := frame.Output()
	if err != nil {
		logrus.WithField("err", err).Error("failed to generate data record frame for pubsub recorder")
		return
	}
	ctx := context.Background()
	res := p.topic.Publish(ctx, &pubsub.Message{Data: output})
	if config.Config.RecorderPubsubVerbose {
		go func() {
			ctx, cancel := context.WithTimeout(ctx, config.Config.RecorderPubsubVerboseCancelTimeout)
			defer cancel()
			id, err := res.Get(ctx)
			if err != nil {
				logrus.WithFields(logrus.Fields{"pubsub_error": err, "id": id}).Error("error pushing to pubsub")
			}
		}()
	}
}
