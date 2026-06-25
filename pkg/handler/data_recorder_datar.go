package handler

import (
	"github.com/openflagr/flagr/pkg/datar"
	"github.com/openflagr/flagr/swagger_gen/models"
)

// NewDatarRecorder creates a DataRecorder that feeds evaluation results into the Datar engine.
func NewDatarRecorder() DataRecorder {
	return &datarRecorder{engine: GetDatar()}
}

// datarRecorder wraps a datar.Engine as a DataRecorder.
// It feeds evaluation results into the in-memory aggregate buffer.
// Unlike Kafka/Kinesis/Pubsub recorders, it does not produce serialized
// frames — NewDataRecordFrame returns an empty frame.
type datarRecorder struct {
	engine *datar.Engine
}

func (d *datarRecorder) AsyncRecord(r models.EvalResult) {
	if r.RecordSource == models.EvalResultRecordSourceExposure {
		return
	}
	d.engine.Record(r.FlagID, r.VariantID, r.SegmentID)
}

func (d *datarRecorder) NewDataRecordFrame(_ models.EvalResult) DataRecordFrame {
	return DataRecordFrame{}
}
