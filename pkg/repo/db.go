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
)

// GetDB gets the db singleton
func GetDB() *gorm.DB {
	singletonOnce.Do(func() {
		db, err := gorm.Open(config.Config.DBDriver, config.Config.DBConnectionStr)
		if err != nil {
			if config.Config.DBConnectionDebug {
				logrus.WithField("err", err).Fatal("failed to connect to db")
			} else {
				logrus.Fatal("failed to connect to db")
			}
		}
		db.SetLogger(logrus.StandardLogger())
		db.AutoMigrate(entity.AutoMigrateTables...)
		singletonDB = db
	})

	return singletonDB
}
