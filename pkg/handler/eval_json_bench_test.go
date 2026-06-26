package handler

import (
	"bytes"
	"testing"

	"github.com/openflagr/flagr/pkg/jsoncodec"
	"github.com/openflagr/flagr/swagger_gen/models"
)

func benchEvalContextJSON(b *testing.B, codec jsoncodec.Codec, body []byte) {
	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		var ec models.EvalContext
		if err := jsoncodec.DecodeJSON(codec, bytes.NewReader(body), &ec); err != nil {
			b.Fatal(err)
		}
	}
}

func benchEvalResultJSON(b *testing.B, codec jsoncodec.Codec, r *models.EvalResult) {
	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		var buf bytes.Buffer
		if err := jsoncodec.EncodeJSON(codec, &buf, r); err != nil {
			b.Fatal(err)
		}
	}
}

var benchEvalContextBody = []byte(`{
  "entityID": "bench-entity",
  "flagID": 1,
  "entityContext": {
    "region": "us-west",
    "tier": "premium",
    "user": {"age": 30, "name": "alice"},
    "tags": ["a","b","c"]
  }
}`)

func largeEvalResult() *models.EvalResult {
	ec := map[string]any{"region": "us-west", "user": map[string]any{"age": float64(30)}}
	return &models.EvalResult{
		EvalContext: &models.EvalContext{
			EntityID:      "bench-entity",
			EntityContext: ec,
			FlagID:        1,
		},
		FlagID:     1,
		VariantKey: "control",
		Timestamp:  "2026-01-01T00:00:00Z",
	}
}

func BenchmarkJSONDecodeEvalContext_Std(b *testing.B) {
	benchEvalContextJSON(b, jsoncodec.CodecStd, benchEvalContextBody)
}

func BenchmarkJSONDecodeEvalContext_Sonic(b *testing.B) {
	benchEvalContextJSON(b, jsoncodec.CodecSonic, benchEvalContextBody)
}

func BenchmarkJSONEncodeEvalResult_Std(b *testing.B) {
	benchEvalResultJSON(b, jsoncodec.CodecStd, largeEvalResult())
}

func BenchmarkJSONEncodeEvalResult_Sonic(b *testing.B) {
	benchEvalResultJSON(b, jsoncodec.CodecSonic, largeEvalResult())
}

func BenchmarkDataRecordFrame_Output(b *testing.B) {
	er := *largeEvalResult()
	frame := DataRecordFrame{
		evalResult: er,
		options:    DataRecordFrameOptions{FrameOutputMode: frameOutputModePayloadRawJSON},
	}
	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		if _, err := frame.Output(); err != nil {
			b.Fatal(err)
		}
	}
}