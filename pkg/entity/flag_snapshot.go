package entity

import (
	"fmt"

	"encoding/json"

	"github.com/openflagr/flagr/pkg/config"
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
func SaveFlagSnapshot(db *gorm.DB, flagID uint, updatedBy string) {
	tx := db.Begin()
	f := &Flag{}
	if err := tx.First(f, flagID).Error; err != nil {
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

	fs := FlagSnapshot{FlagID: f.Model.ID, UpdatedBy: updatedBy, Flag: b}
	if err := tx.Create(&fs).Error; err != nil {
		logrus.WithFields(logrus.Fields{
			"err":    err,
			"flagID": f.Model.ID,
		}).Error("failed to save FlagSnapshot")
		tx.Rollback()
		return
	}

	f.UpdatedBy = updatedBy
	f.SnapshotID = fs.Model.ID

	if err := tx.Save(f).Error; err != nil {
		logrus.WithFields(logrus.Fields{
			"err":            err,
			"flagID":         f.Model.ID,
			"flagSnapshotID": fs.Model.ID,
		}).Error("failed to save Flag's UpdatedBy and SnapshotID")
		tx.Rollback()
		return
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
	}

	logFlagSnapshotUpdate(flagID, updatedBy)
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
