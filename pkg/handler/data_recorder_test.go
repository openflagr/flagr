package handler

import (
	"testing"

	"github.com/prashantv/gostub"
	"github.com/stretchr/testify/assert"
)

func TestGetDataRecorder(t *testing.T) {
	defer gostub.StubFunc(&NewKafkaRecorder, nil).Reset()
	assert.NotPanics(t, func() {
		GetDataRecorder()
	})
}
