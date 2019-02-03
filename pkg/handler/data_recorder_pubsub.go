package handler

import (
	"context"
	"encoding/json"

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
		option.WithCredentialsFile(config.Config.RecorderPubsubKeyFile),
	)
	if err != nil {
		logrus.WithField("pubsub_error", err).Fatal("error getting pubsub client")
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

	messageFrame := pubsubMessageFrame{
		Payload:   string(payload),
		Encrypted: false,
	}

	message, err := messageFrame.encode()
	if err != nil {
		logrus.WithField("pubsub_error", err).Error("error marshaling")
	}

	ctx := context.Background()
	res := p.topic.Publish(ctx, &pubsub.Message{Data: message})
	if config.Config.RecorderPubsubVerbose {
		go func() {
			ctx, cancel := context.WithTimeout(ctx, config.Config.RecorderPubsubVerboseCancel)
			defer cancel()
			id, err := res.Get(ctx)
			if err != nil {
				logrus.WithFields(logrus.Fields{"pubsub_error": err, "id": id}).Error("error pushing to pubsub")
			}
		}()
	}
}

type pubsubEvalResult struct {
	*models.EvalResult
}

type pubsubMessageFrame struct {
	Payload   string `json:"payload"`
	Encrypted bool   `json:"encrypted"`
}

func (pmf *pubsubMessageFrame) encode() ([]byte, error) {
	return json.MarshalIndent(pmf, "", "  ")
}

// Payload marshals the EvalResult
func (r *pubsubEvalResult) Payload() ([]byte, error) {
	return r.EvalResult.MarshalBinary()
}
