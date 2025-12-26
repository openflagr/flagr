package handler

import (
	"net/http"
	"testing"
	"time"

	"github.com/openflagr/flagr/pkg/config"
	"github.com/openflagr/flagr/pkg/entity"
	"github.com/openflagr/flagr/pkg/notification"
	"github.com/openflagr/flagr/pkg/util"
	"github.com/openflagr/flagr/swagger_gen/models"
	"github.com/openflagr/flagr/swagger_gen/restapi/operations/flag"
	"github.com/prashantv/gostub"
	"github.com/stretchr/testify/assert"
)

func TestHandlerNotifications(t *testing.T) {
	db := entity.NewTestDB()
	defer gostub.StubFunc(&getDB, db).Reset()

	mockNotifier := notification.NewMockNotifier()
	notification.SetNotifier(mockNotifier)
	defer notification.SetNotifier(nil)

	c := NewCRUD()

	t.Run("CreateFlag sends notification", func(t *testing.T) {
		mockNotifier.ClearSent()
		params := flag.CreateFlagParams{
			HTTPRequest: &http.Request{},
			Body: &models.CreateFlagRequest{
				Description: util.StringPtr("test flag"),
				Key:         "test_flag_notif",
			},
		}
		c.CreateFlag(params)

		// Notifications are sent in a goroutine, so we might need a small wait or check repeatedly
		assert.Eventually(t, func() bool {
			return len(mockNotifier.GetSentNotifications()) > 0
		}, 1*time.Second, 10*time.Millisecond)

		sent := mockNotifier.GetSentNotifications()
		assert.Len(t, sent, 1)
		assert.Equal(t, notification.OperationCreate, sent[0].Operation)
		assert.Equal(t, notification.EntityTypeFlag, sent[0].EntityType)
		assert.Equal(t, "test_flag_notif", sent[0].EntityKey)
		// Privacy by default
		assert.Empty(t, sent[0].PreValue)
		assert.Empty(t, sent[0].PostValue)
		assert.Empty(t, sent[0].Diff)
	})

	t.Run("PutFlag sends notification", func(t *testing.T) {
		f := entity.GenFixtureFlag()
		db.Create(&f)
		mockNotifier.ClearSent()

		params := flag.PutFlagParams{
			FlagID: int64(f.ID),
			Body: &models.PutFlagRequest{
				Description: util.StringPtr("updated description"),
			},
			HTTPRequest: &http.Request{},
		}
		c.PutFlag(params)

		assert.Eventually(t, func() bool {
			return len(mockNotifier.GetSentNotifications()) > 0
		}, 1*time.Second, 10*time.Millisecond)

		sent := mockNotifier.GetSentNotifications()
		assert.Len(t, sent, 1)
		assert.Equal(t, notification.OperationUpdate, sent[0].Operation)
		assert.Equal(t, f.Key, sent[0].EntityKey)
		// Privacy by default
		assert.Empty(t, sent[0].PreValue)
		assert.Empty(t, sent[0].PostValue)
		assert.Empty(t, sent[0].Diff)
	})

	t.Run("PutFlag with detailed diff enabled", func(t *testing.T) {
		stubs := gostub.Stub(&config.Config.NotificationDetailedDiffEnabled, true)
		defer stubs.Reset()

		f := entity.GenFixtureFlag()
		f.ID = 0 // Allow DB to assign new ID
		f.Key = "detailed_diff_flag"
		db.Create(&f)
		mockNotifier.ClearSent()

		// First update to create first snapshot
		params1 := flag.PutFlagParams{
			FlagID: int64(f.ID),
			Body: &models.PutFlagRequest{
				Description: util.StringPtr("first update"),
			},
			HTTPRequest: &http.Request{},
		}
		c.PutFlag(params1)

		// Second update to trigger diff calculation
		params2 := flag.PutFlagParams{
			FlagID: int64(f.ID),
			Body: &models.PutFlagRequest{
				Description: util.StringPtr("second update"),
			},
			HTTPRequest: &http.Request{},
		}
		c.PutFlag(params2)

		assert.Eventually(t, func() bool {
			return len(mockNotifier.GetSentNotifications()) >= 2
		}, 1*time.Second, 10*time.Millisecond)

		sent := mockNotifier.GetSentNotifications()
		assert.Len(t, sent, 2)
		// Second notification should have a diff
		assert.NotEmpty(t, sent[1].Diff)
		assert.Contains(t, sent[1].Diff, "-  \"Description\": \"first update\"")
		assert.Contains(t, sent[1].Diff, "+  \"Description\": \"second update\"")
	})
}
