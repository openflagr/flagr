package repo

import (
	"sync"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"    // mysql driver
	_ "github.com/jinzhu/gorm/dialects/postgres" // postgres driver

	"github.com/checkr/flagr/pkg/config"
	"github.com/checkr/flagr/pkg/entity"
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
		db, err := gorm.Open(config.Config.DB.DBDriver, config.Config.DB.DBConnectionStr)
		if err != nil {
			panic(err)
		}
		db.AutoMigrate(autoMigrateTables...)
		singletonDB = db
	})

	return singletonDB
}
