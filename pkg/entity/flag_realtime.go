package entity

import (
	"sync"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/sirupsen/logrus"
)

// FlagRealtime is the unit that tracks the realtime information of a flag
type FlagRealtime struct {
	FlagID     uint      `gorm:"primary_key"`
	LastEvalAt time.Time `gorm:"index:idx_flagrealtime_lastevalat"`
}

// FlagRealtimeRepo is an in-memory cache repository of FlagRealtime
type FlagRealtimeRepo struct {
	sync.RWMutex

	db           *gorm.DB
	m            map[uint]FlagRealtime
	syncInterval time.Duration
}

// NewFlagRealtimeRepo creates a new NewFlagRealtimeRepo
func NewFlagRealtimeRepo(db *gorm.DB, syncInterval time.Duration) *FlagRealtimeRepo {
	return &FlagRealtimeRepo{
		db:           db,
		syncInterval: syncInterval,
		m:            make(map[uint]FlagRealtime),
	}
}

// Start will begin a goroutine to write to DB every syncInterval
func (frr *FlagRealtimeRepo) Start() {
	for range time.Tick(frr.syncInterval) {
		err := frr.store()
		if err != nil {
			logrus.WithField("err", err).Error("failed to store FlagRealtimeRepo")
		}
	}
}

// Update updates its cache based on LastEvalAt
func (frr *FlagRealtimeRepo) Update(fr FlagRealtime) {
	if fr.FlagID == 0 {
		return
	}

	frr.Lock()
	defer frr.Unlock()

	old, ok := frr.m[fr.FlagID]
	if !ok {
		frr.m[fr.FlagID] = fr
		return
	}

	if fr.LastEvalAt.After(old.LastEvalAt) {
		frr.m[fr.FlagID] = fr
	}
	return
}

func (frr *FlagRealtimeRepo) store() error {
	frr.RLock()
	defer frr.RUnlock()

	for _, fr := range frr.m {
		fr := fr
		t := fr.LastEvalAt
		if err := frr.db.FirstOrCreate(&fr, fr.FlagID).Error; err != nil {
			return err
		}

		if t.After(fr.LastEvalAt) {
			fr.LastEvalAt = t
			if err := frr.db.Save(&fr).Error; err != nil {
				return err
			}
		}
	}
	return nil
}
