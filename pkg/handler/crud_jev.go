package handler

import (
	"fmt"
	"strings"

	"github.com/openflagr/flagr/pkg/entity"
	"github.com/openflagr/flagr/pkg/mapper/entity_restapi/r2e"
	"github.com/openflagr/flagr/swagger_gen/models"
	"gorm.io/gorm"
)

// applyJevQuestion maps and stores the optional Jev question from a constraint request.
func applyJevQuestion(cons *entity.Constraint, r *models.JevQuestion) error {
	q, err := r2e.MapJevQuestion(r)
	if err != nil {
		return err
	}
	return cons.SetJevQuestion(q)
}

// loadFlagJevConstraints returns the Jev constraints of a flag across all of its
// live segments, optionally excluding one constraint (the one being updated).
func loadFlagJevConstraints(tx *gorm.DB, flagID, excludeID uint) ([]entity.Constraint, error) {
	cs := []entity.Constraint{}
	q := tx.Model(&entity.Constraint{}).
		Joins("JOIN segments ON segments.id = constraints.segment_id").
		Where("segments.flag_id = ?", flagID).
		Where("segments.deleted_at IS NULL").
		Where("constraints.jev_json <> ''")
	if excludeID != 0 {
		q = q.Where("constraints.id <> ?", excludeID)
	}
	if err := q.Find(&cs).Error; err != nil {
		return nil, err
	}
	return cs, nil
}

// validateJevQuestionNameUnique rejects a Jev constraint whose `@jev.<name>` is
// already used by another constraint on the flag. A name may be defined only
// once per flag, so the single batched System One call is unambiguous.
func validateJevQuestionNameUnique(tx *gorm.DB, flagID uint, cons *entity.Constraint, excludeID uint) error {
	if !cons.IsJev() {
		return nil
	}
	others, err := loadFlagJevConstraints(tx, flagID, excludeID)
	if err != nil {
		return err
	}
	names := entity.DuplicateJevQuestionProperties(append(others, *cons))
	if len(names) == 0 {
		return nil
	}
	return fmt.Errorf(
		"jev question %s is already used by another constraint on this flag; each Jev question may be defined only once per flag",
		strings.Join(names, ", "),
	)
}
