package notify

import (
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/openflagr/flagr/pkg/entity"
	"github.com/stretchr/testify/assert"
)

// TestIntegration implements a Notifier for testing purposes.
type TestIntegration struct {
	fakeErr   error
	callCount int
}

// Notify handles notifications for a TestIntegration
func (n *TestIntegration) Notify(f *entity.Flag, b itemAction, i itemType, s subject) error {
	n.callCount++
	return n.fakeErr
}

type roundTripFunc func(req *http.Request) *http.Response

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req), nil
}

//NewTestClient returns *http.Client with Transport replaced to avoid making real calls
func NewTestClient(fn roundTripFunc) *Client {
	return &Client{
		HTTPClient: &http.Client{
			Transport: roundTripFunc(fn),
		},
		RetryWaitMin:  1 * time.Second,
		RetryWaitMax:  10 * time.Second,
		AttemptsMax:   1, // Retries disabled deliberately so that unhappy paths dont take 15seconds!
		CheckForRetry: DefaultRetryPolicy,
		Backoff:       DefaultBackoff,
	}
}

func TestNotifyAll(t *testing.T) {
	t.Run("we return early when the flagID is not found", func(t *testing.T) {
		db := entity.NewTestDB()
		tmpDB, dbErr := db.DB()
		if dbErr != nil {
			t.Errorf("Failed to get database")
		}
		defer tmpDB.Close()
		All(db, 1, TOGGLED, FLAG, "")
	})

	t.Run("nothing bad happens if we have the flag, but no configured integrations", func(t *testing.T) {
		f := entity.GenFixtureFlag()
		db := entity.PopulateTestDB(f)
		tmpDB, dbErr := db.DB()
		if dbErr != nil {
			t.Errorf("Failed to get database")
		}
		defer tmpDB.Close()
		All(db, f.ID, TOGGLED, FLAG, "")
	})

	t.Run("nothing bad happens if an integration fails to deliver and returns an error", func(t *testing.T) {
		f := entity.GenFixtureFlag()
		db := entity.PopulateTestDB(f)
		tmpDB, dbErr := db.DB()
		if dbErr != nil {
			t.Errorf("Failed to get database")
		}
		defer tmpDB.Close()

		notifier := &TestIntegration{fakeErr: errors.New("failed to notify testcase")}
		integration := Integration{notifier: notifier, name: "test"}
		integrations = append(integrations, integration)
		All(db, f.ID, TOGGLED, FLAG, "")
		assert.Equal(t, 1, notifier.callCount)
	})
}
