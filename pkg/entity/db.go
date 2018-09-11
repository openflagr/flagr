package entity

import (
	"os"
	"sync"

	_ "github.com/jinzhu/gorm/dialects/mysql"    // mysql driver
	_ "github.com/jinzhu/gorm/dialects/postgres" // postgres driver
	_ "github.com/jinzhu/gorm/dialects/sqlite"   // sqlite driver

	"github.com/checkr/flagr/pkg/config"
	"github.com/jinzhu/gorm"
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
		db.Debug().AutoMigrate(AutoMigrateTables...)
		singletonDB = db
	})

	return singletonDB
}

// NewSQLiteDB creates a new sqlite db
// useful for backup exports and unit tests
func NewSQLiteDB(filePath string) *gorm.DB {
	os.Remove(filePath)
	db, err := gorm.Open("sqlite3", filePath)
	if err != nil {
		logrus.WithField("err", err).Errorf("failed to connect to db:%s", filePath)
		panic(err)
	}
	db.SetLogger(logrus.StandardLogger())
	db.AutoMigrate(AutoMigrateTables...)
	return db
}

// NewTestDB creates a new test db
func NewTestDB() *gorm.DB {
	return NewSQLiteDB(":memory:")
}

// PopulateTestDB seeds the test db
func PopulateTestDB(flag Flag) *gorm.DB {
	testDB := NewTestDB()
	flag.Create(testDB)
	return testDB
}
