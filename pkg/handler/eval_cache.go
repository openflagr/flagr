package handler

import (
	"sync"
	"time"

	"github.com/openflagr/flagr/swagger_gen/models"

	"github.com/openflagr/flagr/pkg/config"
	"github.com/openflagr/flagr/pkg/entity"
	"github.com/openflagr/flagr/pkg/util"

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
	cache           *cacheContainer
	cacheMutex      sync.RWMutex
	refreshTimeout  time.Duration
	refreshInterval time.Duration
}

// GetEvalCache gets the EvalCache
var GetEvalCache = func() *EvalCache {
	singletonEvalCacheOnce.Do(func() {
		ec := &EvalCache{
			cache:           &cacheContainer{},
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

func (ec *EvalCache) GetByTags(tags []string, operator *string) []*entity.Flag {
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

	ec.cacheMutex.RLock()
	defer ec.cacheMutex.RUnlock()

	for _, t := range tags {
		fSet, ok := ec.cache.tagCache[t]
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

	ec.cacheMutex.RLock()
	defer ec.cacheMutex.RUnlock()

	for i, t := range tags {
		fSet, ok := ec.cache.tagCache[t]
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
	s := util.SafeString(keyOrID)

	ec.cacheMutex.RLock()
	defer ec.cacheMutex.RUnlock()

	f, ok := ec.cache.idCache[s]
	if !ok {
		f = ec.cache.keyCache[s]
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

		ec.cacheMutex.Lock()
		ec.cache = &cacheContainer{
			idCache:  idCache,
			keyCache: keyCache,
			tagCache: tagCache,
		}
		ec.cacheMutex.Unlock()

		return nil, err
	})

	return err
}
