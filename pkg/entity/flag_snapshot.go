package entity

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/openflagr/flagr/pkg/config"
	"github.com/openflagr/flagr/pkg/notification"
	"github.com/openflagr/flagr/pkg/util"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// FlagSnapshot is the snapshot of a flag
// Any change of the flag will create a new snapshot
type FlagSnapshot struct {
	gorm.Model
	FlagID    uint `gorm:"index:idx_flagsnapshot_flagid"`
	UpdatedBy string
	Flag      []byte `gorm:"type:text"`
}

// FlagSnapshotCommitMeta is returned after a snapshot row is written inside a caller transaction.
type FlagSnapshotCommitMeta struct {
	FlagKey   string
	PreValue  string
	PostValue string
	Diff      string
}

// WriteFlagSnapshotTx records a flag snapshot using tx. The caller must Commit or Rollback.
func WriteFlagSnapshotTx(
	tx *gorm.DB,
	flagID uint,
	updatedBy string,
) (FlagSnapshotCommitMeta, error) {
	var meta FlagSnapshotCommitMeta
	f := &Flag{}
	// Use Unscoped to include soft-deleted flags. This is necessary for:
	// 1. Delete operations: we need to snapshot the flag after it's been soft-deleted
	// 2. Restore operations: we need to update the flag that was previously soft-deleted
	if err := tx.Unscoped().First(f, flagID).Error; err != nil {
		logrus.WithFields(logrus.Fields{
			"err":    err,
			"flagID": flagID,
		}).Error("failed to find the flag when WriteFlagSnapshotTx")
		return meta, err
	}
	if err := PreloadSegmentsVariantsTags(tx.Unscoped()).First(f, flagID).Error; err != nil {
		return meta, err
	}

	b, err := json.Marshal(f)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"err":    err,
			"flagID": flagID,
		}).Error("failed to marshal the flag into JSON when WriteFlagSnapshotTx")
		return meta, err
	}

	preFS := &FlagSnapshot{}
	if err := tx.Unscoped().Where("flag_id = ?", flagID).Order("id desc").First(preFS).Error; err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		logrus.WithError(err).WithField("flagID", flagID).Warn("failed to find previous flag snapshot")
	}

	fs := FlagSnapshot{FlagID: f.ID, UpdatedBy: updatedBy, Flag: b}
	if err := tx.Create(&fs).Error; err != nil {
		logrus.WithFields(logrus.Fields{
			"err":    err,
			"flagID": f.Model.ID,
		}).Error("failed to save FlagSnapshot")
		return meta, err
	}

	f.UpdatedBy = updatedBy
	f.SnapshotID = fs.ID

	if err := tx.Unscoped().Save(f).Error; err != nil {
		logrus.WithFields(logrus.Fields{
			"err":            err,
			"flagID":         f.Model.ID,
			"flagSnapshotID": fs.Model.ID,
		}).Error("failed to save Flag's UpdatedBy and SnapshotID")
		return meta, err
	}

	meta.FlagKey = f.Key
	if config.Config.NotificationDetailedDiffEnabled {
		meta.PreValue = string(preFS.Flag)
		meta.PostValue = string(fs.Flag)
		meta.Diff = notification.CalculateDiff(meta.PreValue, meta.PostValue)
	}
	return meta, nil
}

// NotifyFlagSnapshot sends webhook/metrics after the outer transaction committed.
func NotifyFlagSnapshot(
	flagID uint,
	updatedBy string,
	operation notification.Operation,
	componentType notification.ComponentType,
	componentID uint,
	componentKey string,
	meta FlagSnapshotCommitMeta,
) {
	logFlagSnapshotUpdate(flagID, updatedBy)
	notification.SendNotification(notification.Notification{
		Operation:     operation,
		FlagID:        flagID,
		FlagKey:       meta.FlagKey,
		ComponentType: componentType,
		ComponentID:   componentID,
		ComponentKey:  componentKey,
		PreValue:      meta.PreValue,
		PostValue:     meta.PostValue,
		Diff:          meta.Diff,
		User:          updatedBy,
	})
}

// SaveFlagSnapshot saves the Flag Snapshot and sends a notification in its own transaction.
func SaveFlagSnapshot(db *gorm.DB, flagID uint, updatedBy string, operation notification.Operation, componentType notification.ComponentType, componentID uint, componentKey string) {
	tx := db.Begin()
	meta, err := WriteFlagSnapshotTx(tx, flagID, updatedBy)
	if err != nil {
		tx.Rollback()
		return
	}
	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		logrus.WithError(err).WithField("flagID", flagID).Error("failed to commit flag snapshot")
		return
	}
	NotifyFlagSnapshot(flagID, updatedBy, operation, componentType, componentID, componentKey, meta)
}

var logFlagSnapshotUpdate = func(flagID uint, updatedBy string) {
	if config.Global.StatsdClient == nil {
		return
	}

	config.Global.StatsdClient.Incr(
		"flag.snapshot.updated",
		[]string{
			fmt.Sprintf("FlagID:%d", flagID),
			fmt.Sprintf("UpdatedBy:%s", util.SafeStringWithDefault(updatedBy, "null")),
		},
		float64(1),
	)
}