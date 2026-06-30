package handler

import (
	"github.com/openflagr/flagr/pkg/entity"
	"github.com/openflagr/flagr/pkg/notification"
	"gorm.io/gorm"
)

func commitFlagMutation(
	flagID uint,
	subject string,
	operation notification.Operation,
	componentType notification.ComponentType,
	mutate func(tx *gorm.DB) error,
	notifyComponent func() (componentID uint, componentKey string),
) error {
	tx := getDB().Begin()
	if err := mutate(tx); err != nil {
		tx.Rollback()
		return err
	}
	snapshotFlagID := flagID
	if snapshotFlagID == 0 {
		componentID, _ := notifyComponent()
		snapshotFlagID = componentID
	}
	meta, err := entity.WriteFlagSnapshotTx(tx, snapshotFlagID, subject)
	if err != nil {
		tx.Rollback()
		return err
	}
	if err := tx.Commit().Error; err != nil {
		return err
	}
	componentID, componentKey := notifyComponent()
	entity.NotifyFlagSnapshot(snapshotFlagID, subject, operation, componentType, componentID, componentKey, meta)
	return nil
}