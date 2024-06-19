package handler

import (
	"encoding/base64"

	"encoding/json"

	"github.com/brandur/simplebox"
	"github.com/openflagr/flagr/pkg/util"
	"github.com/openflagr/flagr/swagger_gen/models"
)

type dataRecordEncryptor interface {
	Encrypt([]byte) (string, error)
}

type simpleboxEncryptor struct{ key [simplebox.KeySize]byte }

func (se *simpleboxEncryptor) Encrypt(b []byte) (string, error) {
	s := base64.StdEncoding.EncodeToString(
		simplebox.NewFromSecretKey(&se.key).Encrypt(b),
	)
	return s, nil
}

func newSimpleboxEncryptor(k string) dataRecordEncryptor {
	key := [simplebox.KeySize]byte{}
	copy(key[:], k)
	return &simpleboxEncryptor{key: key}
}

const (
	frameOutputModePayloadRawJSON = "payload_raw_json"
)

// DataRecordFrameOptions represents the options we can set to create a DataRecordFrame
type DataRecordFrameOptions struct {
	Encrypted       bool
	Encryptor       dataRecordEncryptor
	FrameOutputMode string
}

type rawPayload struct {
	Payload json.RawMessage `json:"payload"`
}

type stringPayload struct {
	Payload   string `json:"payload"`
	Encrypted bool   `json:"encrypted"`
}

// DataRecordFrame represents the structure we can json.Marshal into data recorders
type DataRecordFrame struct {
	evalResult models.EvalResult
	options    DataRecordFrameOptions
}

// MarshalJSON defines the behavior of MarshalJSON for DataRecordFrame
func (drf *DataRecordFrame) MarshalJSON() ([]byte, error) {
	payload, err := drf.evalResult.MarshalBinary()
	if err != nil {
		return nil, err
	}

	if drf.options.FrameOutputMode == frameOutputModePayloadRawJSON {
		return json.Marshal(&rawPayload{
			Payload: payload,
		})
	}

	if drf.options.Encrypted && drf.options.Encryptor != nil {
		encryptedPayload, err := drf.options.Encryptor.Encrypt(payload)
		if err != nil {
			return nil, err
		}
		return json.Marshal(&stringPayload{
			Payload:   encryptedPayload,
			Encrypted: true,
		})
	}

	return json.Marshal(&stringPayload{
		Payload:   string(payload),
		Encrypted: false,
	})
}

// GetPartitionKey gets the partition key from entityID
func (drf *DataRecordFrame) GetPartitionKey() string {
	if drf.evalResult.EvalContext == nil {
		return ""
	}
	return util.SafeString(drf.evalResult.EvalContext.EntityID)
}

// Output sets the paylaod using its input and returns the json marshal bytes
func (drf *DataRecordFrame) Output() ([]byte, error) {
	return json.Marshal(drf)
}
