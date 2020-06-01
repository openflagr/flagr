package handler

import (
	"sort"
	"sync"
	"time"

	"github.com/checkr/flagr/pkg/config"
	"github.com/checkr/flagr/pkg/entity"
	"github.com/checkr/flagr/pkg/util"

	"github.com/sirupsen/logrus"
	"github.com/zhouzhuojie/withtimeout"
)

var (
	singletonEvalCache     *EvalCache
	singletonEvalCacheOnce sync.Once
)

type mapCache map[string]*entity.Flag
type multiMapCache map[string][]*entity.Flag

// EvalCache is the in-memory cache just for evaluation
type EvalCache struct {
	mapCacheLock sync.RWMutex
	idCache      mapCache
	keyCache     mapCache
	tagCache     multiMapCache

	refreshTimeout  time.Duration
	refreshInterval time.Duration
}

// GetEvalCache gets the EvalCache
var GetEvalCache = func() *EvalCache {
	singletonEvalCacheOnce.Do(func() {
		ec := &EvalCache{
			idCache:         make(map[string]*entity.Flag),
			keyCache:        make(map[string]*entity.Flag),
			tagCache:        make(map[string][]*entity.Flag),
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

func (ec *EvalCache) GetByTags(tags []string) []*entity.Flag {
	ec.mapCacheLock.RLock()
	defer ec.mapCacheLock.RUnlock()

	results := []*entity.Flag{}
	for _, t := range tags {
		s := util.SafeString(t)
		f, ok := ec.tagCache[s]
		if ok {
			results = append(results, f...)
		}
	}

	return FlattenFlags(results)
}

func FlattenFlags(list []*entity.Flag) []*entity.Flag {
	sort.Slice(list, func(i, j int) bool {
		return list[i].ID < list[j].ID
	})

	j := 0
	for i := 1; i < len(list); i++ {
		if list[j] == list[i] {
			continue
		}
		j++

		list[j] = list[i]
	}
	return list[:j+1]
}

// GetByFlagKeyOrID gets the flag by Key or ID
func (ec *EvalCache) GetByFlagKeyOrID(keyOrID interface{}) *entity.Flag {
	ec.mapCacheLock.RLock()
	defer ec.mapCacheLock.RUnlock()

	s := util.SafeString(keyOrID)
	f, ok := ec.idCache[s]
	if !ok {
		f = ec.keyCache[s]
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

		ec.mapCacheLock.Lock()
		defer ec.mapCacheLock.Unlock()

		ec.idCache = idCache
		ec.keyCache = keyCache
		ec.tagCache = tagCache
		return nil, err
	})

	return err
}
