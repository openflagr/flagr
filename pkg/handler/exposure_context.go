package handler

import (
	"encoding/json"
)

// mergeExposureEntityContext merges entityContext and metadata into one map for evalContext.entityContext.
// Uses JSON round-trip so swagger interface{} payloads normalize consistently with eval JSON decoding.
func mergeExposureEntityContext(entityContext, metadata any) map[string]interface{} {
	out := map[string]interface{}{}
	mergeJSONMapInto(out, entityContext)
	mergeJSONMapInto(out, metadata)
	if len(out) == 0 {
		return nil
	}
	return out
}

func mergeJSONMapInto(dst map[string]interface{}, src any) {
	if src == nil {
		return
	}
	b, err := json.Marshal(src)
	if err != nil || len(b) == 0 || string(b) == "null" {
		return
	}
	var m map[string]interface{}
	if err := json.Unmarshal(b, &m); err != nil || len(m) == 0 {
		return
	}
	for k, v := range m {
		dst[k] = v
	}
}