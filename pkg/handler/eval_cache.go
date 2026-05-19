package handler

import (
	"maps"
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

	// lastSnapshotMaxID tracks the highest flag_snapshot ID seen on the last
	// successful reload. The cache short-circuits when this hasn't changed,
	// because every API mutation that affects eval data creates a snapshot.
	// lastSnapshotMaxID > 0 indicates at least one successful load has occurred.
	lastSnapshotMaxID uint
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
			maps.Copy(results, fSet)
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
			maps.Copy(results, fSet)
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
func (ec *EvalCache) GetByFlagKeyOrID(keyOrID any) *entity.Flag {
	s := util.SafeString(keyOrID)

	ec.cacheMutex.RLock()
	defer ec.cacheMutex.RUnlock()

	f, ok := ec.cache.idCache[s]
	if !ok {
		f = ec.cache.keyCache[s]
	}
	return f
}

// reloadMapCache reloads the evaluation cache from the database. It short-circuits
// when no new flag_snapshots have been created, since every API mutation that
// affects evaluation data (flags, segments, variants, constraints, distributions,
// tags) creates a flag_snapshot row.
func (ec *EvalCache) reloadMapCache() error {
	if config.Config.NewRelicEnabled {
		defer config.Global.NewrelicApp.StartTransaction("eval_cache_reload", nil, nil).End()
	}

	// Lightweight check: has any evaluation data changed since last reload?
	// Every mutation handler that affects the cache creates a flag_snapshot,
	// so a rising MAX(id) means the cache is stale.
	var currentMaxID uint
	if err := getDB().Model(&entity.FlagSnapshot{}).
		Select("COALESCE(MAX(id), 0)").
		Scan(&currentMaxID).Error; err != nil {
		logrus.WithField("err", err).Warn(
			"failed to query flag_snapshots MAX(id), falling back to full reload")
	} else if currentMaxID == ec.lastSnapshotMaxID && ec.lastSnapshotMaxID > 0 {
		return nil
	}

	_, _, err := withtimeout.Do(ec.refreshTimeout, func() (any, error) {
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

		return nil, nil
	})

	if err == nil {
		ec.lastSnapshotMaxID = currentMaxID
	}

	return err
}
