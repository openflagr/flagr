package handler

import (
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/openflagr/flagr/pkg/config"
	"github.com/openflagr/flagr/pkg/entity"
	"github.com/openflagr/flagr/pkg/notification"
	"github.com/openflagr/flagr/swagger_gen/models"
	"github.com/openflagr/flagr/swagger_gen/restapi/operations/flag"
	"github.com/prashantv/gostub"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	notificationWaitTimeout = time.Second
	notificationPollEvery   = 10 * time.Millisecond
)

// waitForNotification waits for an async SendNotification delivery that matches.
// Delivery order is not guaranteed: SendNotification dispatches in a goroutine.
func waitForNotification(t *testing.T, mock *notification.MockNotifier, match func(notification.Notification) bool) notification.Notification {
	t.Helper()
	var found notification.Notification
	require.Eventually(t, func() bool {
		for _, n := range mock.GetSentNotifications() {
			if match(n) {
				found = n
				return true
			}
		}
		return false
	}, notificationWaitTimeout, notificationPollEvery)
	return found
}

func TestHandlerNotifications(t *testing.T) {
	db, cleanup := handlerTestDB(t)
	defer cleanup()

	mockNotifier := notification.NewMockNotifier()
	// Use gostub to set notifiers and reset via defer
	stubs := gostub.Stub(&notification.Notifiers, []notification.Notifier{mockNotifier})
	defer stubs.Reset()

	c := NewCRUD()

	t.Run("CreateFlag sends notification", func(t *testing.T) {
		mockNotifier.ClearSent()
		params := flag.CreateFlagParams{
			HTTPRequest: &http.Request{},
			Body: &models.CreateFlagRequest{
				Description: new("test flag"),
				Key:         "test_flag_notif",
			},
		}
		c.CreateFlag(params)

		sent := waitForNotification(t, mockNotifier, func(n notification.Notification) bool {
			return n.Operation == notification.OperationCreate && n.FlagKey == "test_flag_notif"
		})
		assert.Empty(t, sent.PreValue)
		assert.Empty(t, sent.PostValue)
		assert.Empty(t, sent.Diff)
	})

	t.Run("PutFlag sends notification", func(t *testing.T) {
		f := entity.GenFixtureFlag()
		db.Create(&f)
		mockNotifier.ClearSent()

		params := flag.PutFlagParams{
			FlagID: int64(f.ID),
			Body: &models.PutFlagRequest{
				Description: new("updated description"),
			},
			HTTPRequest: &http.Request{},
		}
		c.PutFlag(params)

		sent := waitForNotification(t, mockNotifier, func(n notification.Notification) bool {
			return n.Operation == notification.OperationUpdate && n.FlagKey == f.Key
		})
		assert.Empty(t, sent.PreValue)
		assert.Empty(t, sent.PostValue)
		assert.Empty(t, sent.Diff)
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
				Description: new("first update"),
			},
			HTTPRequest: &http.Request{},
		}
		c.PutFlag(params1)
		waitForNotification(t, mockNotifier, func(n notification.Notification) bool {
			return n.Operation == notification.OperationUpdate && n.FlagKey == f.Key
		})

		params2 := flag.PutFlagParams{
			FlagID: int64(f.ID),
			Body: &models.PutFlagRequest{
				Description: new("second update"),
			},
			HTTPRequest: &http.Request{},
		}
		c.PutFlag(params2)

		const (
			diffRemoved = `-  "Description": "first update"`
			diffAdded   = `+  "Description": "second update"`
		)
		sent := waitForNotification(t, mockNotifier, func(n notification.Notification) bool {
			return n.FlagKey == f.Key &&
				strings.Contains(n.Diff, diffRemoved) &&
				strings.Contains(n.Diff, diffAdded)
		})
		assert.NotEmpty(t, sent.Diff)
	})

	t.Run("DeleteFlag sends notification", func(t *testing.T) {
		f := entity.Flag{Key: "delete_notif_flag", Description: "to delete", Enabled: true}
		assert.NoError(t, db.Create(&f).Error)
		mockNotifier.ClearSent()

		params := flag.DeleteFlagParams{
			FlagID:      int64(f.ID),
			HTTPRequest: &http.Request{},
		}
		c.DeleteFlag(params)

		waitForNotification(t, mockNotifier, func(n notification.Notification) bool {
			return n.Operation == notification.OperationDelete && n.FlagKey == f.Key
		})
	})

	t.Run("RestoreFlag sends notification", func(t *testing.T) {
		f := entity.Flag{Key: "restore_notif_flag", Description: "restore", Enabled: true}
		assert.NoError(t, db.Create(&f).Error)
		db.Delete(&f)
		mockNotifier.ClearSent()

		params := flag.RestoreFlagParams{
			FlagID:      int64(f.ID),
			HTTPRequest: &http.Request{},
		}
		c.RestoreFlag(params)

		waitForNotification(t, mockNotifier, func(n notification.Notification) bool {
			return n.Operation == notification.OperationRestore && n.FlagKey == f.Key
		})
	})

	t.Run("SetFlagEnabledState sends notification", func(t *testing.T) {
		f := entity.Flag{Key: "enabled_notif_flag", Description: "enabled", Enabled: true}
		assert.NoError(t, db.Create(&f).Error)
		mockNotifier.ClearSent()

		params := flag.SetFlagEnabledParams{
			FlagID: int64(f.ID),
			Body: &models.SetFlagEnabledRequest{
				Enabled: new(false),
			},
			HTTPRequest: &http.Request{},
		}
		c.SetFlagEnabledState(params)

		sent := waitForNotification(t, mockNotifier, func(n notification.Notification) bool {
			return n.Operation == notification.OperationUpdate && n.FlagID == f.ID
		})
		assert.Equal(t, f.Key, sent.FlagKey)
	})

	t.Run("DuplicateFlag sends notification", func(t *testing.T) {
		mockNotifier.ClearSent()
		createParams := flag.CreateFlagParams{
			HTTPRequest: &http.Request{},
			Body: &models.CreateFlagRequest{
				Description: new("dup notif source"),
				Key:         "dup_notif_src",
			},
		}
		createRes := c.CreateFlag(createParams)
		createOK := createRes.(*flag.CreateFlagOK)
		require.NotNil(t, createOK.Payload)
		sourceID := uint(createOK.Payload.ID)

		waitForNotification(t, mockNotifier, func(n notification.Notification) bool {
			return n.Operation == notification.OperationCreate && n.FlagID == sourceID
		})
		mockNotifier.ClearSent()

		dupRes := c.DuplicateFlag(flag.DuplicateFlagParams{
			FlagID:      createOK.Payload.ID,
			HTTPRequest: &http.Request{},
		})
		dupOK, ok := dupRes.(*flag.DuplicateFlagOK)
		require.True(t, ok, "duplicate failed: %T", dupRes)
		require.NotNil(t, dupOK.Payload)
		assert.NotEqual(t, createOK.Payload.ID, dupOK.Payload.ID)

		dupFlagID := uint(dupOK.Payload.ID)
		dupNotif := waitForNotification(t, mockNotifier, func(n notification.Notification) bool {
			return n.Operation == notification.OperationCreate && n.FlagID == dupFlagID
		})
		assert.NotEqual(t, "dup_notif_src", dupNotif.FlagKey)
		assert.NotEmpty(t, dupNotif.FlagKey)
	})
}
