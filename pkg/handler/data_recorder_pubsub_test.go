package handler

import (
	"context"
	"testing"

	"cloud.google.com/go/pubsub"
	"cloud.google.com/go/pubsub/pstest"
	"github.com/openflagr/flagr/swagger_gen/models"
	"github.com/prashantv/gostub"
	"github.com/stretchr/testify/assert"
	"google.golang.org/api/option"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func TestNewPubsubRecorder(t *testing.T) {
	t.Run("no panics", func(t *testing.T) {
		client := mockClient(t)
		defer client.Close()

		defer gostub.StubFunc(
			&pubsubClient,
			client,
			nil,
		).Reset()

		assert.NotPanics(t, func() { NewPubsubRecorder() })
	})
}

func TestPubsubAsyncRecord(t *testing.T) {
	t.Run("enabled and valid", func(t *testing.T) {
		client := mockClient(t)
		defer client.Close()
		topic := client.Topic("test")
		assert.NotPanics(t, func() {
			pr := &pubsubRecorder{
				producer: client,
				topic:    topic,
			}

			pr.AsyncRecord(
				models.EvalResult{
					EvalContext: &models.EvalContext{
						EntityID: "d08042018",
					},
					FlagID:         1,
					FlagSnapshotID: 1,
					SegmentID:      1,
					VariantID:      1,
					VariantKey:     "control",
				},
			)
		})
	})
}

func mockClient(t *testing.T) *pubsub.Client {
	ctx := context.Background()
	srv := pstest.NewServer()
	defer srv.Close()
	conn, err := grpc.Dial(srv.Addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatal("cannot connect to mocked server")
	}
	defer conn.Close()
	client, err := pubsub.NewClient(ctx, "project", option.WithGRPCConn(conn))

	if err != nil {
		t.Fatal("failed creating mock client", err)
	}

	return client
}
