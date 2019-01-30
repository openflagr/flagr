package handler

import (
	"context"

	"cloud.google.com/go/pubsub"
	"github.com/checkr/flagr/pkg/config"
	"github.com/checkr/flagr/swagger_gen/models"
	"google.golang.org/api/option"

	"github.com/sirupsen/logrus"
)

type pubsubRecorder struct {
	enabled  bool
	producer *pubsub.Client
	topic    *pubsub.Topic
}

// NewPubsubRecorder creates a new Pubsub recorder
var NewPubsubRecorder = func() DataRecorder {
	client, err := pubsub.NewClient(
		context.Background(),
		config.Config.RecorderPubsubProjectID,
		option.WithServiceAccountFile(config.Config.RecorderPubsubKeyFile),
	)
	if err != nil {
		logrus.WithField("pubsub_error", err).Error("error getting pubsub client")
	}

	return &pubsubRecorder{
		producer: client,
		topic:    client.Topic(config.Config.RecorderPubsubTopicName),
		enabled:  config.Config.RecorderEnabled,
	}
}

func (p *pubsubRecorder) AsyncRecord(r *models.EvalResult) {
	if !p.enabled {
		return
	}

	pr := &pubsubEvalResult{
		EvalResult: r,
	}

	payload, err := pr.Payload()
	if err != nil {
		logrus.WithField("pubsub_error", err).Error("error marshaling")
	}

	ctx := context.Background()
	res := p.topic.Publish(ctx, &pubsub.Message{Data: payload})
	if config.Config.RecorderPubsubVerbose {
		id, err := res.Get(ctx)
		if err != nil {
			logrus.WithField("pubsub_error", err).Errorf("error pushing to pubsub: %v", id)
		}
	}
}

type pubsubEvalResult struct {
	*models.EvalResult
}

// Payload marshals the EvalResult
func (r *pubsubEvalResult) Payload() ([]byte, error) {
	return r.EvalResult.MarshalBinary()
}
