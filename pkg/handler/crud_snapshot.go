package handler

import (
	"github.com/openflagr/flagr/pkg/entity"
	"github.com/openflagr/flagr/pkg/notification"
	"gorm.io/gorm"
)

// MutationNotify carries webhook component metadata after a successful mutation.
type MutationNotify struct {
	ComponentID  uint
	ComponentKey string
}

// commitFlagMutation runs mutate in one transaction, writes a flag snapshot on the same tx, commits, then notifies.
// snapshotFlagID is the flag whose history row is updated (use 0 when the new flag ID is assigned inside mutate).
func commitFlagMutation(
	snapshotFlagID uint,
	subject string,
	operation notification.Operation,
	componentType notification.ComponentType,
	mutate func(tx *gorm.DB) (uint, MutationNotify, error),
) error {
	tx := getDB().Begin()
	resolvedID, notify, err := mutate(tx)
	if err != nil {
		tx.Rollback()
		return err
	}
	flagIDForSnapshot := snapshotFlagID
	if flagIDForSnapshot == 0 {
		flagIDForSnapshot = resolvedID
	}
	meta, err := entity.WriteFlagSnapshotTx(tx, flagIDForSnapshot, subject)
	if err != nil {
		tx.Rollback()
		return err
	}
	if err := tx.Commit().Error; err != nil {
		return err
	}
	entity.NotifyFlagSnapshot(flagIDForSnapshot, subject, operation, componentType, notify.ComponentID, notify.ComponentKey, meta)
	return nil
}