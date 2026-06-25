package handler

import (
	"fmt"
	"time"

	"github.com/openflagr/flagr/pkg/config"
	"github.com/openflagr/flagr/pkg/entity"
	"github.com/openflagr/flagr/swagger_gen/models"
)

func validateAndBuildExposure(row *models.Exposure) (models.EvalResult, error) {
	if row.EntityID == nil || *row.EntityID == "" {
		return models.EvalResult{}, fmt.Errorf("entityID is required")
	}

	flag, err := resolveFlagFromExposure(row)
	if err != nil {
		return models.EvalResult{}, err
	}

	variantID, variantKey, err := resolveVariantOnFlag(flag, row.VariantID, row.VariantKey)
	if err != nil {
		return models.EvalResult{}, err
	}

	snapshotID := int64(flag.SnapshotID)
	if row.FlagSnapshotID > 0 {
		sid := uint(row.FlagSnapshotID)
		if !flagSnapshotExistsInDB(flag.ID, sid) {
			return models.EvalResult{}, fmt.Errorf("flagSnapshotID %d not found for flag", sid)
		}
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

	evalCtx := models.EvalContext{
		EntityID:      *row.EntityID,
		EntityType:    entityType,
		EntityContext: mergeExposureEntityContext(row.EntityContext, row.Metadata),
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

// flagSnapshotExistsInDB checks the snapshots table (skipped in eval-only mode where DB snapshots are unavailable).
func flagSnapshotExistsInDB(flagID, snapshotID uint) bool {
	if config.Config.EvalOnlyMode {
		return false
	}
	var fs entity.FlagSnapshot
	err := getDB().Where("id = ? AND flag_id = ?", snapshotID, flagID).First(&fs).Error
	return err == nil
}