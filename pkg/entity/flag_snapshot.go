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

// snapshotNotificationPayload is populated by WriteFlagSnapshotTx inside the caller's transaction
// (flag key and optional detailed-diff fields). It is not exported: pass the opaque
// SnapshotNotification wrapper to NotifyAfterCommit only after the outer transaction commits.
type snapshotNotificationPayload struct {
	flagKey   string
	preValue  string
	postValue string
	diff      string
}

// SnapshotNotification is an opaque handle returned from WriteFlagSnapshotTx. Call
// NotifyAfterCommit after the outer transaction commits; do not inspect or construct it outside entity.
type SnapshotNotification struct {
	payload snapshotNotificationPayload
}

// WriteFlagSnapshotTx records a flag snapshot using tx. The caller must Commit or Rollback.
func WriteFlagSnapshotTx(
	tx *gorm.DB,
	flagID uint,
	updatedBy string,
) (SnapshotNotification, error) {
	var out SnapshotNotification
	f := &Flag{}
	// Use Unscoped to include soft-deleted flags. This is necessary for:
	// 1. Delete operations: we need to snapshot the flag after it's been soft-deleted
	// 2. Restore operations: we need to update the flag that was previously soft-deleted
	if err := tx.Unscoped().First(f, flagID).Error; err != nil {
		logrus.WithFields(logrus.Fields{
			"err":    err,
			"flagID": flagID,
		}).Error("failed to find the flag when WriteFlagSnapshotTx")
		return out, err
	}
	if err := PreloadSegmentsVariantsTags(tx.Unscoped()).First(f, flagID).Error; err != nil {
		return out, err
	}

	b, err := json.Marshal(f)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"err":    err,
			"flagID": flagID,
		}).Error("failed to marshal the flag into JSON when WriteFlagSnapshotTx")
		return out, err
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
		return out, err
	}

	f.UpdatedBy = updatedBy
	f.SnapshotID = fs.ID

	if err := tx.Unscoped().Save(f).Error; err != nil {
		logrus.WithFields(logrus.Fields{
			"err":            err,
			"flagID":         f.Model.ID,
			"flagSnapshotID": fs.Model.ID,
		}).Error("failed to save Flag's UpdatedBy and SnapshotID")
		return out, err
	}

	out.payload.flagKey = f.Key
	if config.Config.NotificationDetailedDiffEnabled {
		out.payload.preValue = string(preFS.Flag)
		out.payload.postValue = string(fs.Flag)
		out.payload.diff = notification.CalculateDiff(out.payload.preValue, out.payload.postValue)
	}
	return out, nil
}

// NotifyAfterCommit sends webhook/metrics after the outer transaction committed.
func (n SnapshotNotification) NotifyAfterCommit(
	flagID uint,
	updatedBy string,
	operation notification.Operation,
	componentType notification.ComponentType,
	componentID uint,
	componentKey string,
) {
	logFlagSnapshotUpdate(flagID, updatedBy)
	notification.SendNotification(notification.Notification{
		Operation:     operation,
		FlagID:        flagID,
		FlagKey:       n.payload.flagKey,
		ComponentType: componentType,
		ComponentID:   componentID,
		ComponentKey:  componentKey,
		PreValue:      n.payload.preValue,
		PostValue:     n.payload.postValue,
		Diff:          n.payload.diff,
		User:          updatedBy,
	})
}

// SaveFlagSnapshot saves the Flag Snapshot and sends a notification in its own transaction.
func SaveFlagSnapshot(db *gorm.DB, flagID uint, updatedBy string, operation notification.Operation, componentType notification.ComponentType, componentID uint, componentKey string) {
	tx := db.Begin()
	snap, err := WriteFlagSnapshotTx(tx, flagID, updatedBy)
	if err != nil {
		tx.Rollback()
		return
	}
	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		logrus.WithError(err).WithField("flagID", flagID).Error("failed to commit flag snapshot")
		return
	}
	snap.NotifyAfterCommit(flagID, updatedBy, operation, componentType, componentID, componentKey)
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