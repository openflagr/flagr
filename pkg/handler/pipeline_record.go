package handler

import (
	"github.com/openflagr/flagr/pkg/config"
	"github.com/openflagr/flagr/pkg/entity"
	"github.com/openflagr/flagr/swagger_gen/models"
)

// shouldRecordPipelineEvent is true when FLAGR_RECORDER_ENABLED and the flag has dataRecordsEnabled.
func shouldRecordPipelineEvent(flag *entity.Flag) bool {
	if flag == nil || !config.Config.RecorderEnabled {
		return false
	}
	return flag.DataRecordsEnabled
}

// recordPipelineEvent writes to configured data recorders (Kafka, Kinesis, Pub/Sub, Datar fan-out).
// Stubbable in tests. Does not emit eval Prometheus/Datadog metrics.
var recordPipelineEvent = func(r models.EvalResult) {
	if !config.Config.RecorderEnabled {
		return
	}
	GetDataRecorder().AsyncRecord(r)
}

// isExposurePipelineEvent is true when r is a client-reported exposure (not an evaluation assignment).
func isExposurePipelineEvent(r models.EvalResult) bool {
	return r.RecordSource == models.EvalResultRecordSourceExposure
}