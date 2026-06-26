package jsoncodec

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/openflagr/flagr/swagger_gen/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRoundTripEvalContext_stdAndSonic(t *testing.T) {
	body := []byte(`{"entityID":"e1","entityContext":{"k":1,"nested":{"x":2}},"flagID":1}`)
	for _, codec := range []Codec{CodecStd, CodecSonic} {
		t.Run(string(codec), func(t *testing.T) {
			var ec models.EvalContext
			require.NoError(t, DecodeJSON(codec, bytes.NewReader(body), &ec))
			assert.Equal(t, "e1", ec.EntityID)
			var buf bytes.Buffer
			require.NoError(t, EncodeJSON(codec, &buf, &ec))
			assert.Contains(t, buf.String(), "entityID")
		})
	}
}

func TestParseCodec(t *testing.T) {
	assert.Equal(t, CodecStd, ParseCodec(""))
	assert.Equal(t, CodecSonic, ParseCodec("sonic"))
	assert.Equal(t, CodecStd, ParseCodec("unknown"))
}

func TestStdUseNumber(t *testing.T) {
	var m map[string]any
	require.NoError(t, DecodeJSON(CodecStd, bytes.NewReader([]byte(`{"n":42}`)), &m))
	_, ok := m["n"].(json.Number)
	assert.True(t, ok)
}