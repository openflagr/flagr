package entity

import (
	"testing"

	"github.com/openflagr/flagr/pkg/config"
	"github.com/stretchr/testify/assert"
)

func setTestDBConfig(driver string, connectionStr string) (reset func()) {
	old := config.Config

	config.Config.DBDriver = driver
	config.Config.DBConnectionStr = connectionStr
	config.Config.DBConnectionRetryAttempts = 2

	return func() {
		config.Config = old
	}
}

func TestConnectDB(t *testing.T) {
	t.Run("happy code path", func(t *testing.T) {
		reset := setTestDBConfig("sqlite3", ":memory:")
		defer reset()

		db, err := connectDB()
		assert.NotNil(t, db)
		assert.NoError(t, err)
	})

	t.Run("error code path", func(t *testing.T) {
		reset := setTestDBConfig("mysql", "invalid")
		defer reset()

		_, err := connectDB()
		assert.Error(t, err)
	})
}

func TestGetDB(t *testing.T) {
	reset := setTestDBConfig("sqlite3", ":memory:")
	defer reset()

	db := GetDB()
	assert.NotNil(t, db)
}
