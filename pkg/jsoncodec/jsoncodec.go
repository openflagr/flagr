// Package jsoncodec provides HTTP JSON encode/decode helpers shared by the API
// server and benchmarks. UseNumber on decode preserves json.Number for entityContext.
package jsoncodec

import (
	"encoding/json"
	"io"

	"github.com/bytedance/sonic"
)

// Codec selects the JSON implementation for API bodies.
type Codec string

const (
	CodecStd   Codec = "std"
	CodecSonic Codec = "sonic"
)

// DecodeJSON reads one JSON value into v.
func DecodeJSON(codec Codec, r io.Reader, v any) error {
	switch codec {
	case CodecSonic:
		return sonic.Config{UseNumber: true}.Froze().NewDecoder(r).Decode(v)
	default:
		dec := json.NewDecoder(r)
		dec.UseNumber()
		return dec.Decode(v)
	}
}

// EncodeJSON writes one JSON value.
func EncodeJSON(codec Codec, w io.Writer, v any) error {
	switch codec {
	case CodecSonic:
		return sonic.Config{EscapeHTML: false}.Froze().NewEncoder(w).Encode(v)
	default:
		enc := json.NewEncoder(w)
		enc.SetEscapeHTML(false)
		return enc.Encode(v)
	}
}

// ParseCodec maps config string to Codec (unknown → std).
func ParseCodec(s string) Codec {
	switch s {
	case string(CodecSonic):
		return CodecSonic
	default:
		return CodecStd
	}
}