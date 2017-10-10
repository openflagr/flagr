package handler

import (
	"sync"
	"time"

	"github.com/checkr/flagr/pkg/config"
	"github.com/checkr/flagr/pkg/entity"
	"github.com/checkr/flagr/pkg/repo"
	raven "github.com/getsentry/raven-go"
)

var (
	singletonEvalCache     *EvalCache
	singletonEvalCacheOnce sync.Once
)

// EvalCache is the in-memory cache just for evaluation
type EvalCache struct {
	mapCache     map[uint]*entity.Flag
	mapCacheLock sync.RWMutex

	refreshTimeout  time.Duration
	refreshInterval time.Duration

	redisCache *redisCache
}

// GetEvalCache gets the EvalCache
func GetEvalCache() *EvalCache {
	singletonEvalCacheOnce.Do(func() {
		ec := &EvalCache{
			mapCache:        make(map[uint]*entity.Flag),
			refreshTimeout:  config.Config.EvalCache.RefreshTimeout.Duration,
			refreshInterval: config.Config.EvalCache.RefreshInterval.Duration,
			redisCache:      &redisCache{},
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
				raven.CaptureError(err, map[string]string{
					"msg": "reload evaluation cache error",
					"err": err.Error(),
				})
			}
		}
	}()
}

// GetByFlagIDs gets the flags by flagIDs from the EvalCache
func (ec *EvalCache) GetByFlagIDs(flagIDs []uint) map[uint]*entity.Flag {
	m := make(map[uint]*entity.Flag)

	ec.mapCacheLock.RLock()
	for _, flagID := range flagIDs {
		f, ok := ec.mapCache[flagID]
		if ok {
			m[flagID] = f
		}
	}
	ec.mapCacheLock.RUnlock()
	return m
}

func (ec *EvalCache) reloadMapCache() error {
	fs := []entity.Flag{}
	q := entity.NewFlagQuerySet(repo.GetDB())
	if err := q.All(&fs); err != nil {
		return err
	}
	m := make(map[uint]*entity.Flag)
	for _, f := range fs {
		err := f.Preload(repo.GetDB())
		if err != nil {
			return err
		}
		m[f.ID] = &f
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

type redisCache struct {
}

func (rc *redisCache) GetFlags() ([]entity.Flag, error) {
	return nil, nil
}

func (rc *redisCache) SetFlags(fs []entity.Flag) error {
	return nil
}

func (rc *redisCache) Lock() error {
	return nil
}

func (rc *redisCache) Unlock() error {
	return nil
}
