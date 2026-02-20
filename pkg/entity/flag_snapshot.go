package entity

import (
	"fmt"

	"encoding/json"

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

// SaveFlagSnapshot saves the Flag Snapshot
func SaveFlagSnapshot(db *gorm.DB, flagID uint, updatedBy string, operation notification.Operation) {
	tx := db.Begin()
	f := &Flag{}
	// Use Unscoped to include soft-deleted flags (needed for delete notifications)
	if err := tx.Unscoped().First(f, flagID).Error; err != nil {
		logrus.WithFields(logrus.Fields{
			"err":    err,
			"flagID": flagID,
		}).Error("failed to find the flag when SaveFlagSnapshot")
		return
	}
	f.Preload(tx)

	b, err := json.Marshal(f)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"err":    err,
			"flagID": flagID,
		}).Error("failed to marshal the flag into JSON when SaveFlagSnapshot")
		return
	}

	fs := FlagSnapshot{FlagID: f.ID, UpdatedBy: updatedBy, Flag: b}
	if err := tx.Create(&fs).Error; err != nil {
		logrus.WithFields(logrus.Fields{
			"err":    err,
			"flagID": f.Model.ID,
		}).Error("failed to save FlagSnapshot")
		tx.Rollback()
		return
	}

	f.UpdatedBy = updatedBy
	f.SnapshotID = fs.ID

	// Use Unscoped to ensure we can update soft-deleted flags (e.g., after delete)
	if err := tx.Unscoped().Save(f).Error; err != nil {
		logrus.WithFields(logrus.Fields{
			"err":            err,
			"flagID":         f.Model.ID,
			"flagSnapshotID": fs.Model.ID,
		}).Error("failed to save Flag's UpdatedBy and SnapshotID")
		tx.Rollback()
		return
	}

	preFS := &FlagSnapshot{}
	// Find the most recent snapshot before the current one (use Unscoped to include any soft-deleted)
	tx.Unscoped().Where("flag_id = ? AND id < ?", flagID, fs.ID).Order("id desc").First(preFS)

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return
	}

	preValue := ""
	postValue := ""
	diff := ""

	if config.Config.NotificationDetailedDiffEnabled {
		preValue = string(preFS.Flag)
		postValue = string(fs.Flag)
		diff = notification.CalculateDiff(preValue, postValue)
	}

	logFlagSnapshotUpdate(flagID, updatedBy)
	notification.SendFlagNotification(
		operation,
		flagID,
		f.Key,
		f.Description,
		preValue,
		postValue,
		diff,
		updatedBy,
	)
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
