package entity

import (
	"os"
	"sync"
	"time"

	sqlite "github.com/glebarez/sqlite" // sqlite driver with pure go
	mysql "gorm.io/driver/mysql"        // mysql driver
	postgres "gorm.io/driver/postgres"  // postgres driver

	retry "github.com/avast/retry-go"
	"github.com/openflagr/flagr/pkg/config"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
	gorm_logger "gorm.io/gorm/logger"
)

var (
	singletonDB   *gorm.DB
	singletonOnce sync.Once
)

// AutoMigrateTables stores the entity tables that we can auto migrate in gorm
var AutoMigrateTables = []any{
	Flag{},
	Constraint{},
	Distribution{},
	FlagSnapshot{},
	Segment{},
	User{},
	Variant{},
	Tag{},
	FlagEntityType{},
}

func connectDB() (db *gorm.DB, err error) {
	logger := &Logger{
		LogLevel:                  gorm_logger.Info,
		SlowThreshold:             time.Millisecond,
		IgnoreRecordNotFoundError: false,
	}

	err = retry.Do(
		func() error {
			switch config.Config.DBDriver {
			case "postgres":
				db, err = gorm.Open(postgres.Open(config.Config.DBConnectionStr), &gorm.Config{
					Logger: logger,
				})
			case "sqlite3":
				db, err = gorm.Open(sqlite.Open(config.Config.DBConnectionStr), &gorm.Config{
					Logger: logger,
				})
			case "mysql":
				db, err = gorm.Open(mysql.Open(config.Config.DBConnectionStr), &gorm.Config{
					Logger: logger,
				})
			}
			return err
		},
		retry.Attempts(config.Config.DBConnectionRetryAttempts),
		retry.Delay(config.Config.DBConnectionRetryDelay),
	)
	return db, err
}

// GetDB gets the db singleton
func GetDB() *gorm.DB {
	singletonOnce.Do(func() {
		db, err := connectDB()
		if err != nil {
			if config.Config.DBConnectionDebug {
				logrus.WithField("err", err).Fatal("failed to connect to db")
			} else {
				logrus.Fatal("failed to connect to db")
			}
		}
		db.AutoMigrate(AutoMigrateTables...)
		singletonDB = db
	})

	return singletonDB
}

// NewSQLiteDB creates a new sqlite db
// useful for backup exports and unit tests
func NewSQLiteDB(filePath string) *gorm.DB {
	os.Remove(filePath)

	db, err := gorm.Open(sqlite.Open(filePath), &gorm.Config{})
	if err != nil {
		logrus.WithField("err", err).Errorf("failed to connect to db:%s", filePath)
		panic(err)
	}
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
	testDB.Create(&flag)
	return testDB
}
