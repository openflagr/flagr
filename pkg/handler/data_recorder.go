package handler

import (
	"fmt"
	"sync"

	"github.com/openflagr/flagr/pkg/config"
	"github.com/openflagr/flagr/swagger_gen/models"
)

var (
	singletonDataRecorder     DataRecorder
	singletonDataRecorderOnce sync.Once
)

// DataRecorder can record and produce the evaluation result
type DataRecorder interface {
	AsyncRecord(models.EvalResult)
	NewDataRecordFrame(models.EvalResult) DataRecordFrame
}




// fanOutRecorder broadcasts AsyncRecord to multiple DataRecorder implementations.
type fanOutRecorder []DataRecorder

func (f fanOutRecorder) AsyncRecord(r models.EvalResult) {
	for _, rec := range f {
		rec.AsyncRecord(r)
	}
}

func (f fanOutRecorder) NewDataRecordFrame(_ models.EvalResult) DataRecordFrame {
	return DataRecordFrame{}
}

// GetDataRecorder gets the data recorder
func GetDataRecorder() DataRecorder {
	singletonDataRecorderOnce.Do(func() {
		if !config.Config.RecorderEnabled {
			singletonDataRecorder = fanOutRecorder(nil)
			return
		}

		var recs []DataRecorder
		for _, rt := range config.Config.RecorderType {
			switch rt {
			case "kafka":
				recs = append(recs, NewKafkaRecorder())
			case "kinesis":
				recs = append(recs, NewKinesisRecorder())
			case "pubsub":
				recs = append(recs, NewPubsubRecorder())
			case "datar":
				recs = append(recs, NewDatarRecorder())
			default:
				panic(fmt.Sprintf("recorderType %q not supported", rt))
			}
		}
		singletonDataRecorder = fanOutRecorder(recs)
	})

	return singletonDataRecorder
}