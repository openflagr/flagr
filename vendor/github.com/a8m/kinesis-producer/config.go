package producer

import (
	"time"

	k "github.com/aws/aws-sdk-go/service/kinesis"
	"github.com/sirupsen/logrus"
)

// Constants and default configuration take from:
// github.com/awslabs/amazon-kinesis-producer/.../KinesisProducerConfiguration.java
const (
	maxRecordSize        = 1 << 20 // 1MiB
	maxRequestSize       = 5 << 20 // 5MiB
	maxRecordsPerRequest = 500
	maxAggregationSize   = 51200 // 50KB
	// The KinesisProducerConfiguration set the default to 4294967295L;
	// it's kinda odd, because the maxAggregationSize is limit to 51200L;
	maxAggregationCount   = 4294967295
	defaultMaxConnections = 24
	defaultFlushInterval  = time.Second
)

// Putter is the interface that wraps the KinesisAPI.PutRecords method.
type Putter interface {
	PutRecords(*k.PutRecordsInput) (*k.PutRecordsOutput, error)
}

// Config is the Producer configuration.
type Config struct {
	// StreamName is the Kinesis stream.
	StreamName string

	// FlushInterval is a regular interval for flushing the buffer. Defaults to 5s.
	FlushInterval time.Duration

	// BatchCount determine the maximum number of items to pack in batch.
	// Must not exceed length. Defaults to 500.
	BatchCount int

	// BatchSize determine the maximum number of bytes to send with a PutRecords request.
	// Must not exceed 5MiB; Default to 5MiB.
	BatchSize int

	// AggregateBatchCount determine the maximum number of items to pack into an aggregated record.
	AggregateBatchCount int

	// AggregationBatchSize determine the maximum number of bytes to pack into an aggregated record.
	AggregateBatchSize int

	// BacklogCount determines the channel capacity before Put() will begin blocking. Default to `BatchCount`.
	BacklogCount int

	// Number of requests to sent concurrently. Default to 24.
	MaxConnections int

	// Logger is the logger used. Default to logrus.Log.
	Logger logrus.FieldLogger

	// Enabling verbose logging. Default to false.
	Verbose bool

	// Client is the Putter interface implementation.
	Client Putter
}

// defaults for configuration
func (c *Config) defaults() {
	if c.Logger == nil {
		c.Logger = logrus.New()
	}
	if c.BatchCount == 0 {
		c.BatchCount = maxRecordsPerRequest
	}
	falseOrPanic(c.BatchCount > maxRecordsPerRequest, "kinesis: BatchCount exceeds 500")
	if c.BatchSize == 0 {
		c.BatchSize = maxRequestSize
	}
	falseOrPanic(c.BatchSize > maxRequestSize, "kinesis: BatchSize exceeds 5MiB")
	if c.BacklogCount == 0 {
		c.BacklogCount = maxRecordsPerRequest
	}
	if c.AggregateBatchCount == 0 {
		c.AggregateBatchCount = maxAggregationCount
	}
	falseOrPanic(c.AggregateBatchCount > maxAggregationCount, "kinesis: AggregateBatchCount exceeds 4294967295")
	if c.AggregateBatchSize == 0 {
		c.AggregateBatchSize = maxAggregationSize
	}
	falseOrPanic(c.AggregateBatchSize > maxAggregationSize, "kinesis: AggregateBatchSize exceeds 50KB")
	if c.MaxConnections == 0 {
		c.MaxConnections = defaultMaxConnections
	}
	falseOrPanic(c.MaxConnections < 1 || c.MaxConnections > 256, "kinesis: MaxConnections must be between 1 and 256")
	if c.FlushInterval == 0 {
		c.FlushInterval = time.Second * 5
	}
	falseOrPanic(len(c.StreamName) == 0, "kinesis: StreamName length must be at least 1")
}

func falseOrPanic(p bool, msg string) {
	if p {
		panic(msg)
	}
}
