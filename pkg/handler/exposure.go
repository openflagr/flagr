package handler

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/go-openapi/runtime/middleware"
	"github.com/openflagr/flagr/pkg/config"
	"github.com/openflagr/flagr/pkg/entity"
	"github.com/openflagr/flagr/swagger_gen/models"
	exposureapi "github.com/openflagr/flagr/swagger_gen/restapi/operations/exposure"
)

// Exposure handles POST /exposures.
type Exposure interface {
	PostExposures(exposureapi.PostExposuresParams) middleware.Responder
}

// NewExposure creates an Exposure handler.
func NewExposure() Exposure {
	return &exposureHandler{}
}

type exposureHandler struct{}

func (h *exposureHandler) PostExposures(params exposureapi.PostExposuresParams) middleware.Responder {
	if params.Body == nil || len(params.Body.Exposures) == 0 {
		return exposureapi.NewPostExposuresDefault(400).WithPayload(
			ErrorMessage("exposures array is required and must not be empty"))
	}

	exposures := params.Body.Exposures
	if max := config.Config.ExposureBatchSize; max > 0 && len(exposures) > max {
		return exposureapi.NewPostExposuresDefault(400).WithPayload(
			ErrorMessage("exposure batch size %d exceeds maximum allowed size of %d", len(exposures), max))
	}

	var logged int64
	var rowErrors []*models.ExposureRowError

	for i, row := range exposures {
		if row == nil {
			logExposureStatsd("rejected", 0, "")
			rowErrors = append(rowErrors, &models.ExposureRowError{Index: int64(i), Message: "exposure row is null"})
			continue
		}

		dataRecord, flag, err := buildExposureDataRecord(row)
		if err != nil {
			logExposureStatsd("rejected", row.FlagID, row.FlagKey)
			rowErrors = append(rowErrors, &models.ExposureRowError{Index: int64(i), Message: err.Error()})
			continue
		}

		logExposureStatsd("accepted", dataRecord.FlagID, dataRecord.FlagKey)

		if dataRecordEnabled(flag) {
			GetDataRecorder().AsyncRecord(dataRecord)
			logExposureStatsd("recorded", dataRecord.FlagID, dataRecord.FlagKey)
			logged++
		}
	}

	resp := exposureapi.NewPostExposuresOK()
	resp.SetPayload(&models.ExposuresResponse{
		LoggedCount: logged,
		Message:     "Exposures logged successfully",
		Errors:      rowErrors,
	})
	return resp
}

// buildExposureDataRecord validates an exposure row and returns the EvalResult
// passed to DataRecorder.AsyncRecord (recordSource: exposure) and the resolved flag.
func buildExposureDataRecord(row *models.Exposure) (models.EvalResult, *entity.Flag, error) {
	if row.EntityID == nil || *row.EntityID == "" {
		return models.EvalResult{}, nil, fmt.Errorf("entityID is required")
	}

	flag, err := resolveExposureFlag(GetEvalCache(), row)
	if err != nil {
		return models.EvalResult{}, nil, err
	}

	variantID, variantKey, err := resolveExposureVariant(flag, row.VariantID, row.VariantKey)
	if err != nil {
		return models.EvalResult{}, nil, err
	}

	snapshotID := int64(flag.SnapshotID)
	if row.FlagSnapshotID > 0 {
		snapshotID = row.FlagSnapshotID
	}

	entityType := row.EntityType
	if flag.EntityType != "" {
		entityType = flag.EntityType
	}

	ts := time.Now().UTC().Format(time.RFC3339)
	if !time.Time(row.Timestamp).IsZero() {
		ts = time.Time(row.Timestamp).UTC().Format(time.RFC3339)
	}

	entityCtx := map[string]any{}
	mergeJSONIntoMap(entityCtx, row.EntityContext)
	mergeJSONIntoMap(entityCtx, row.Metadata)
	var merged any
	if len(entityCtx) > 0 {
		merged = entityCtx
	}

	evalCtx := models.EvalContext{
		EntityID:      *row.EntityID,
		EntityType:    entityType,
		EntityContext: merged,
	}

	return models.EvalResult{
		FlagID:         int64(flag.ID),
		FlagKey:        flag.Key,
		FlagSnapshotID: snapshotID,
		SegmentID:      0,
		VariantID:      variantID,
		VariantKey:     variantKey,
		Timestamp:      ts,
		RecordSource:   models.EvalResultRecordSourceExposure,
		EvalContext:    &evalCtx,
	}, flag, nil
}

func resolveExposureFlag(ec *EvalCache, row *models.Exposure) (*entity.Flag, error) {
	hasID := row.FlagID > 0
	hasKey := row.FlagKey != ""
	if !hasID && !hasKey {
		return nil, fmt.Errorf("flagID or flagKey is required")
	}

	var flag *entity.Flag
	if hasID {
		flag = ec.GetByFlagKeyOrID(row.FlagID)
	}
	if hasKey {
		byKey := ec.GetByFlagKeyOrID(row.FlagKey)
		switch {
		case byKey == nil && flag == nil:
			return nil, fmt.Errorf("flag not found")
		case byKey == nil:
			// ID resolved flag; key absent in cache — keep flag
		case flag == nil:
			flag = byKey
		case flag.ID != byKey.ID:
			return nil, fmt.Errorf("flagID and flagKey refer to different flags")
		}
	}
	if flag == nil {
		return nil, fmt.Errorf("flag not found")
	}
	return flag, nil
}

func resolveExposureVariant(flag *entity.Flag, variantID int64, variantKey string) (id int64, key string, err error) {
	if variantID <= 0 && variantKey == "" {
		return 0, "", nil
	}

	var byID, byKey *entity.Variant
	for i := range flag.Variants {
		v := &flag.Variants[i]
		if variantID > 0 && v.ID == uint(variantID) {
			byID = v
		}
		if variantKey != "" && v.Key == variantKey {
			byKey = v
		}
	}

	if variantID > 0 && variantKey != "" {
		if byID == nil {
			return 0, "", fmt.Errorf("variantID %d not found on flag", variantID)
		}
		if byKey == nil {
			return 0, "", fmt.Errorf("variantKey %q not found on flag", variantKey)
		}
		if byID.ID != byKey.ID || byID.Key != byKey.Key {
			return 0, "", fmt.Errorf("variantID and variantKey do not match")
		}
		return int64(byID.ID), byID.Key, nil
	}

	if variantID > 0 {
		if byID == nil {
			return 0, "", fmt.Errorf("variantID %d not found on flag", variantID)
		}
		return int64(byID.ID), byID.Key, nil
	}

	if byKey == nil {
		return 0, "", fmt.Errorf("variantKey %q not found on flag", variantKey)
	}
	return int64(byKey.ID), byKey.Key, nil
}

// mergeJSONIntoMap copies top-level keys from arbitrary client JSON (swagger any)
// into dst. Non-objects, null, and empty objects are intentionally ignored.
func mergeJSONIntoMap(dst map[string]any, src any) {
	if src == nil {
		return
	}
	b, err := json.Marshal(src)
	if err != nil || len(b) == 0 || string(b) == "null" {
		return
	}
	var m map[string]any
	if err := json.Unmarshal(b, &m); err != nil || len(m) == 0 {
		return
	}
	for k, v := range m {
		dst[k] = v
	}
}

var logExposureStatsd = func(status string, flagID int64, flagKey string) {
	if config.Global.StatsdClient == nil {
		return
	}
	tags := []string{fmt.Sprintf("status:%s", status)}
	if flagID > 0 {
		tags = append(tags, fmt.Sprintf("FlagID:%d", flagID))
	}
	if flagKey != "" {
		tags = append(tags, fmt.Sprintf("FlagKey:%s", flagKey))
	}
	metric := "exposure.ingest"
	if status == "recorded" {
		metric = "exposure.recorded"
	}
	config.Global.StatsdClient.Incr(metric, tags, 1)
}