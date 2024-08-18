package entity

import (
	"context"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
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

const AZ_DB_SCOPE = "https://ossrdbms-aad.database.windows.net/.default"

var (
	singletonDB   *gorm.DB
	singletonOnce sync.Once
)

// AutoMigrateTables stores the entity tables that we can auto migrate in gorm
var AutoMigrateTables = []interface{}{
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
				connStr := config.Config.DBConnectionStr
				if config.Config.AzureDBAuth {
					token, e := GetAzureToken(
						config.Config.AzurePostgresDBAuthTenant,
						config.Config.AzurePostgresDBAuthClientId,
						config.Config.AzurePostgresDBAuthClientSecret,
					)
					if e != nil {
						return e
					}
					connStr = config.Config.AzurePostgresDBConnectionString + " password=" + token
				}
				db, err = gorm.Open(postgres.Open(connStr), &gorm.Config{
					Logger: logger,
				})
			case "sqlite3":
				db, err = gorm.Open(sqlite.Open(config.Config.DBConnectionStr), &gorm.Config{
					Logger: logger,
				})
			case "mysql":
				connStr := config.Config.DBConnectionStr
				if config.Config.AzureDBAuth {
					token, e := GetAzureToken(
						config.Config.AzureMySQLDBAuthTenant,
						config.Config.AzureMySQLDBAuthClientId,
						config.Config.AzureMySQLDBAuthClientSecret,
					)
					if e != nil {
						return e
					}
					connStr = config.Config.AzurePostgresDBConnectionString + "; Password=" + token
				}
				db, err = gorm.Open(mysql.Open(connStr), &gorm.Config{
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

func GetAzureToken(tenantId string, clientId string, clientSecret string) (token string, err error) {

	// set background context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if config.Config.AzureAuthType == "system" {

		cred, e := azidentity.NewDefaultAzureCredential(nil)
		if e != nil {
			logrus.WithField("err", err).Warn("azure system user failed to get system identity")
			return "", err
		}
		t, e := cred.GetToken(ctx, policy.TokenRequestOptions{
			Scopes: []string{AZ_DB_SCOPE},
		})
		if e != nil {
			logrus.WithField("err", e).Warn("azure system user failed to get token")
			return "", e
		}
		token = t.Token
	}
	if config.Config.AzureAuthType == "user" {
		// For user-assigned identity.
		options := &azidentity.ManagedIdentityCredentialOptions{ID: azidentity.ClientID(clientId)}
		cred, e := azidentity.NewManagedIdentityCredential(options)
		if e != nil {
			logrus.WithField("err", e).Warn("azure managed user failed to get identity")
			return "", err
		}
		t, e := cred.GetToken(ctx, policy.TokenRequestOptions{
			Scopes: []string{AZ_DB_SCOPE},
		})
		if e != nil {
			logrus.WithField("err", e).Warn("azure managed user failed to get token")
			return "", e
		}
		token = t.Token
	}
	if config.Config.AzureAuthType == "service-principal" {
		// For service principal.
		cred, e := azidentity.NewClientSecretCredential(tenantId, clientId, clientSecret, &azidentity.ClientSecretCredentialOptions{})
		if e != nil {
			logrus.WithField("err", e).Warn("azure service-principal failed to get identity")
			return "", e
		}
		t, e := cred.GetToken(ctx, policy.TokenRequestOptions{
			Scopes: []string{AZ_DB_SCOPE},
		})
		if e != nil{
			logrus.WithField("err", e).Warn("azure service-principal failed to get token")
			return "", e

		}
		token = t.Token
	}

	return token, nil
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
