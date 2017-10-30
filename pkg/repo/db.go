package repo

import (
	"sync"

	"github.com/checkr/flagr/pkg/config"
	"github.com/checkr/flagr/pkg/entity"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"    // mysql driver
	_ "github.com/jinzhu/gorm/dialects/postgres" // postgres driver
	_ "github.com/jinzhu/gorm/dialects/sqlite"   // sqlite driver
	"github.com/sirupsen/logrus"
)

var (
	singletonDB   *gorm.DB
	singletonOnce sync.Once

	autoMigrateTables = []interface{}{
		entity.Constraint{},
		entity.Distribution{},
		entity.Flag{},
		entity.Segment{},
		entity.User{},
		entity.Variant{},
	}
)

// GetDB gets the db singleton
func GetDB() *gorm.DB {
	singletonOnce.Do(func() {
		db, err := gorm.Open(config.Config.DBDriver, config.Config.DBConnectionStr)
		if err != nil {
			logrus.WithField("err", err).Error("failed to connect to db")
			panic(err)
		}
		db.SetLogger(logrus.StandardLogger())
		db.AutoMigrate(autoMigrateTables...)
		singletonDB = db
	})

	return singletonDB
}
