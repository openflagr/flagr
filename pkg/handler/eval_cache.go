// nolint: errcheck
package handler

import (
	"sync"
	"sync/atomic"
	"time"

	"github.com/Allen-Career-Institute/flagr/swagger_gen/models"

	"github.com/Allen-Career-Institute/flagr/pkg/config"
	"github.com/Allen-Career-Institute/flagr/pkg/entity"
	"github.com/Allen-Career-Institute/flagr/pkg/util"

	"github.com/sirupsen/logrus"
	"github.com/zhouzhuojie/withtimeout"
)

var (
	singletonEvalCache     *EvalCache
	singletonEvalCacheOnce sync.Once
)

type cacheContainer struct {
	idCache  map[string]*entity.Flag
	keyCache map[string]*entity.Flag
	tagCache map[string]map[uint]*entity.Flag
}

// EvalCache is the in-memory cache just for evaluation
type EvalCache struct {
	cache           atomic.Value
	refreshTimeout  time.Duration
	refreshInterval time.Duration
	isInitialized   atomic.Bool // Track if cache loaded successfully
}

// GetEvalCache gets the EvalCache
var GetEvalCache = func() *EvalCache {
	singletonEvalCacheOnce.Do(func() {
		ec := &EvalCache{
			refreshTimeout:  config.Config.EvalCacheRefreshTimeout,
			refreshInterval: config.Config.EvalCacheRefreshInterval,
		}
		ec.isInitialized.Store(true)
		singletonEvalCache = ec
	})
	return singletonEvalCache
}

// Start starts the polling of EvalCache
func (ec *EvalCache) Start() {
	// Initial load attempt
	err := ec.reloadMapCache()
	if err != nil {
		// Log error instead of panic
		logrus.WithError(err).Error("initial cache load failed - feature flag evaluations will be disabled")
		ec.isInitialized.Store(false)
	}

	// Background refresh
	go func() {
		for range time.Tick(ec.refreshInterval) {
			err := ec.reloadMapCache()
			if err != nil {
				logrus.WithField("err", err).Error("reload evaluation cache error")
			} else {
				// Enable evaluations if cache load succeeds
				wasInitialized := ec.isInitialized.Load()
				ec.isInitialized.Store(true)

				if !wasInitialized {
					logrus.Info("cache successfully reloaded - feature flag evaluations are now enabled")
				}
			}
		}
	}()
}

func (ec *EvalCache) GetByTags(tags []string, operator *string) []*entity.Flag {
	if !ec.isInitialized.Load() {
		logrus.Error("cache not initialized - returning empty list for flag evaluation")
		return []*entity.Flag{}
	}

	var results map[uint]*entity.Flag

	if operator == nil || *operator == models.EvaluationBatchRequestFlagTagsOperatorANY {
		results = ec.getByTagsANY(tags)
	}

	if operator != nil && *operator == models.EvaluationBatchRequestFlagTagsOperatorALL {
		results = ec.getByTagsALL(tags)
	}

	values := make([]*entity.Flag, 0, len(results))
	for _, f := range results {
		values = append(values, f)
	}

	return values
}

func (ec *EvalCache) getByTagsANY(tags []string) map[uint]*entity.Flag {
	results := map[uint]*entity.Flag{}
	cache := ec.cache.Load().(*cacheContainer)

	for _, t := range tags {
		fSet, ok := cache.tagCache[t]
		if ok {
			for fID, f := range fSet {
				results[fID] = f
			}
		}
	}
	return results
}

func (ec *EvalCache) getByTagsALL(tags []string) map[uint]*entity.Flag {
	results := map[uint]*entity.Flag{}
	cache := ec.cache.Load().(*cacheContainer)

	for i, t := range tags {
		fSet, ok := cache.tagCache[t]
		if !ok {
			// no flags
			return map[uint]*entity.Flag{}
		}

		if i == 0 {
			// store all the flags
			for fID, f := range fSet {
				results[fID] = f
			}
		} else {
			for fID := range results {
				if _, ok := fSet[fID]; !ok {
					delete(results, fID)
				}
			}

			// no flags left
			if len(results) == 0 {
				return results
			}
		}
	}

	return results
}

// GetByFlagKeyOrID gets the flag by Key or ID
func (ec *EvalCache) GetByFlagKeyOrID(keyOrID interface{}) *entity.Flag {
	if !ec.isInitialized.Load() {
		logrus.Error("cache not initialized - returning nil for flag evaluation")
		return nil
	}

	cacheVal := ec.cache.Load()
	if cacheVal == nil {
		logrus.Error("cache is nil - returning nil for flag evaluation")
		return nil
	}

	cache, ok := cacheVal.(*cacheContainer)
	if !ok {
		logrus.Error("invalid cache type - returning nil for flag evaluation")
		return nil
	}

	s := util.SafeString(keyOrID)
	f, ok := cache.idCache[s]
	if !ok {
		f = cache.keyCache[s]
	}
	return f
}

func (ec *EvalCache) reloadMapCache() error {
	if config.Config.NewRelicEnabled {
		defer config.Global.NewrelicApp.StartTransaction("eval_cache_reload", nil, nil).End()
	}

	_, _, err := withtimeout.Do(ec.refreshTimeout, func() (interface{}, error) {
		idCache, keyCache, tagCache, err := ec.fetchAllFlags()
		if err != nil {
			return nil, err
		}

		ec.cache.Store(&cacheContainer{
			idCache:  idCache,
			keyCache: keyCache,
			tagCache: tagCache,
		})
		return nil, err
	})

	return err
}
