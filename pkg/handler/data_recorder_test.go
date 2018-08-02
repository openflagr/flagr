package handler

import (
	"sync"
	"testing"

	"github.com/checkr/flagr/pkg/config"

	"github.com/prashantv/gostub"
	"github.com/stretchr/testify/assert"
)

func TestGetDataRecorder(t *testing.T) {
	singletonDataRecorderOnce = sync.Once{}
	defer gostub.StubFunc(&NewKafkaRecorder, nil).Reset()

	assert.NotPanics(t, func() {
		GetDataRecorder()
	})
}

func TestGetDataRecorderWhenKinesisIsSet(t *testing.T) {
	singletonDataRecorderOnce = sync.Once{}
	defer gostub.StubFunc(&NewKinesisRecorder, nil).Reset()
	config.Config.RecorderType = "kinesis"

	assert.NotPanics(t, func() {
		GetDataRecorder()
	})

	config.Config.RecorderType = "kafka"
}

func TestGetDataRecorderPanicsWhenRecorderIsInvalid(t *testing.T) {
	singletonDataRecorderOnce = sync.Once{}
	config.Config.RecorderType = "invalid"

	assert.Panics(t, func() {
		GetDataRecorder()
	})

	config.Config.RecorderType = "kafka"
}
