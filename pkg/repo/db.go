package repo

import (
	"sync"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql" // mysql driver

	"github.com/checkr/flagr/pkg/config"
	"github.com/checkr/flagr/pkg/entity"
)

var (
	singletonDB   *gorm.DB
	singletonOnce sync.Once
)

// GetDB gets the db singleton
func GetDB() *gorm.DB {
	singletonOnce.Do(func() {
		db, err := gorm.Open(config.Config.DB.DBDriver, config.Config.DB.DBConnectionStr)
		if err != nil {
			panic(err)
		}
		db.AutoMigrate(
			entity.Constraint{},
			entity.Distribution{},
			entity.Flag{},
			entity.Segment{},
			entity.User{},
			entity.Variant{},
		)
		singletonDB = db
	})

	return singletonDB
}
