package handler

import (
	"sync"
	"time"

	"github.com/checkr/flagr/pkg/config"
	"github.com/checkr/flagr/pkg/entity"
	"github.com/checkr/flagr/pkg/util"
	"github.com/jinzhu/gorm"
	"github.com/sirupsen/logrus"
)

var (
	singletonEvalCache     *EvalCache
	singletonEvalCacheOnce sync.Once
)

// EvalCache is the in-memory cache just for evaluation
type EvalCache struct {
	mapCache     map[string]*entity.Flag
	mapCacheLock sync.RWMutex

	refreshTimeout  time.Duration
	refreshInterval time.Duration
}

// GetEvalCache gets the EvalCache
var GetEvalCache = func() *EvalCache {
	singletonEvalCacheOnce.Do(func() {
		ec := &EvalCache{
			mapCache:        make(map[string]*entity.Flag),
			refreshTimeout:  config.Config.EvalCacheRefreshTimeout,
			refreshInterval: config.Config.EvalCacheRefreshInterval,
		}
		singletonEvalCache = ec
	})
	return singletonEvalCache
}

// Start starts the polling of EvalCache
func (ec *EvalCache) Start() {
	err := ec.reloadMapCache()
	if err != nil {
		panic(err)
	}
	go func() {
		for range time.Tick(ec.refreshInterval) {
			err := ec.reloadMapCache()
			if err != nil {
				logrus.WithField("err", err).Error("reload evaluation cache error")
			}
		}
	}()
}

// GetByFlagKeyOrID gets the flag by Key or ID
func (ec *EvalCache) GetByFlagKeyOrID(keyOrID interface{}) *entity.Flag {
	ec.mapCacheLock.RLock()
	f := ec.mapCache[util.SafeString(keyOrID)]
	ec.mapCacheLock.RUnlock()
	return f
}

var fetchAllFlags = func() ([]entity.Flag, error) {
	// Use eager loading to avoid N+1 problem
	// doc: http://jinzhu.me/gorm/crud.html#preloading-eager-loading
	fs := []entity.Flag{}
	err := getDB().Preload("Segments", func(db *gorm.DB) *gorm.DB {
		return db.Preload("Distributions", func(db *gorm.DB) *gorm.DB {
			return db.Order("variant_id ASC")
		}).Preload("Constraints", func(db *gorm.DB) *gorm.DB {
			return db.Order("created_at ASC")
		}).Order("rank ASC").Order("id ASC")
	}).Preload("Variants").Find(&fs).Error
	return fs, err
}

func (ec *EvalCache) reloadMapCache() error {
	if config.Config.NewRelicEnabled {
		defer config.Global.NewrelicApp.StartTransaction("eval_cache_reload", nil, nil).End()
	}

	fs, err := fetchAllFlags()
	if err != nil {
		return err
	}
	m := make(map[string]*entity.Flag)
	for i := range fs {
		ptr := &fs[i]
		if ptr.ID != 0 {
			m[util.SafeString(ptr.ID)] = ptr
		}
		if ptr.Key != "" {
			m[ptr.Key] = ptr
		}
	}

	for _, f := range m {
		err := f.PrepareEvaluation()
		if err != nil {
			return err
		}
	}

	ec.mapCacheLock.Lock()
	ec.mapCache = m
	ec.mapCacheLock.Unlock()
	return nil
}
