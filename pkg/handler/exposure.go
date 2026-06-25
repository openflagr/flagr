package handler

import (
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
			ErrorMessage("exposure batch size %d exceeds maximum of %d", len(exposures), max))
	}

	var logged int64
	var rowErrors []*models.ExposureRowError

	for i, row := range exposures {
		if row == nil {
			rowErrors = append(rowErrors, exposureRowErr(int64(i), "exposure row is null"))
			logExposureIngestStatsd("rejected", 0, "")
			continue
		}
		synthetic, err := validateAndBuildExposure(row)
		if err != nil {
			rowErrors = append(rowErrors, exposureRowErr(int64(i), err.Error()))
			logExposureIngestStatsd("rejected", 0, "")
			continue
		}

		logExposureIngestStatsd("accepted", synthetic.FlagID, synthetic.FlagKey)

		if !config.Config.RecorderEnabled {
			continue
		}
		flag := GetEvalCache().GetByFlagKeyOrID(synthetic.FlagID)
		if flag == nil || !flag.DataRecordsEnabled {
			continue
		}

		GetDataRecorder().AsyncRecord(synthetic)
		logExposureIngestStatsd("recorded", synthetic.FlagID, synthetic.FlagKey)
		logged++
	}

	msg := "Exposures logged successfully"
	resp := exposureapi.NewPostExposuresOK()
	resp.SetPayload(&models.ExposuresResponse{
		LoggedCount: logged,
		Message:     msg,
		Errors:      rowErrors,
	})
	return resp
}

func exposureRowErr(index int64, message string) *models.ExposureRowError {
	return &models.ExposureRowError{Index: index, Message: message}
}

func validateAndBuildExposure(row *models.Exposure) (models.EvalResult, error) {
	if row.EntityID == nil || *row.EntityID == "" {
		return models.EvalResult{}, fmt.Errorf("entityID is required")
	}

	hasID := row.FlagID > 0
	hasKey := row.FlagKey != ""
	if !hasID && !hasKey {
		return models.EvalResult{}, fmt.Errorf("flagID or flagKey is required")
	}

	ec := GetEvalCache()
	var flag *entity.Flag
	if hasID {
		flag = ec.GetByFlagKeyOrID(row.FlagID)
	}
	if hasKey {
		byKey := ec.GetByFlagKeyOrID(row.FlagKey)
		if byKey == nil {
			if flag == nil {
				return models.EvalResult{}, fmt.Errorf("flag not found")
			}
		} else if flag == nil {
			flag = byKey
		} else if flag.ID != byKey.ID {
			return models.EvalResult{}, fmt.Errorf("flagID and flagKey refer to different flags")
		}
	}
	if flag == nil {
		return models.EvalResult{}, fmt.Errorf("flag not found")
	}

	var variantID int64
	variantKey := row.VariantKey
	if row.VariantID > 0 {
		vid := uint(row.VariantID)
		if !variantOnFlag(flag, vid, "") {
			return models.EvalResult{}, fmt.Errorf("variantID %d not found on flag", vid)
		}
		variantID = row.VariantID
		for _, v := range flag.Variants {
			if v.ID == vid {
				variantKey = v.Key
				break
			}
		}
	}
	if row.VariantKey != "" {
		if !variantOnFlag(flag, 0, row.VariantKey) {
			return models.EvalResult{}, fmt.Errorf("variantKey %q not found on flag", row.VariantKey)
		}
		variantKey = row.VariantKey
		for _, v := range flag.Variants {
			if v.Key == variantKey {
				variantID = int64(v.ID)
				break
			}
		}
	}
	if row.VariantID > 0 && row.VariantKey != "" {
		if !variantOnFlag(flag, uint(row.VariantID), row.VariantKey) {
			return models.EvalResult{}, fmt.Errorf("variantID and variantKey do not match")
		}
	}

	snapshotID := int64(flag.SnapshotID)
	if row.FlagSnapshotID > 0 {
		sid := uint(row.FlagSnapshotID)
		if !flagSnapshotExists(flag.ID, sid) {
			return models.EvalResult{}, fmt.Errorf("flagSnapshotID %d not found for flag", sid)
		}
		snapshotID = row.FlagSnapshotID
	}

	entityCtx := mergeExposureContext(row.EntityContext, row.Metadata)
	entityType := row.EntityType
	if flag.EntityType != "" {
		entityType = flag.EntityType
	}

	ts := time.Now().UTC().Format(time.RFC3339)
	if !time.Time(row.Timestamp).IsZero() {
		ts = time.Time(row.Timestamp).UTC().Format(time.RFC3339)
	}

	evalCtx := models.EvalContext{
		EntityID:      *row.EntityID,
		EntityType:    entityType,
		EntityContext: entityCtx,
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
	}, nil
}

func variantOnFlag(flag *entity.Flag, variantID uint, variantKey string) bool {
	for _, v := range flag.Variants {
		if variantID > 0 && v.ID == variantID {
			return true
		}
		if variantKey != "" && v.Key == variantKey {
			return true
		}
	}
	return false
}

func flagSnapshotExists(flagID, snapshotID uint) bool {
	if config.Config.EvalOnlyMode {
		return false
	}
	var fs entity.FlagSnapshot
	err := getDB().Where("id = ? AND flag_id = ?", snapshotID, flagID).First(&fs).Error
	return err == nil
}

func mergeExposureContext(entityContext, metadata any) map[string]interface{} {
	out := map[string]interface{}{}
	if m, ok := entityContext.(map[string]interface{}); ok {
		for k, v := range m {
			out[k] = v
		}
	}
	if m, ok := metadata.(map[string]interface{}); ok {
		for k, v := range m {
			out[k] = v
		}
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

var logExposureIngestStatsd = func(status string, flagID int64, flagKey string) {
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