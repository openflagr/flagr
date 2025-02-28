package handler

import (
	"fmt"
	"io"
	"net/http"
	"os"

	"encoding/json"

	"github.com/openflagr/flagr/pkg/config"
	"github.com/openflagr/flagr/pkg/entity"
	"github.com/openflagr/flagr/pkg/util"
	"gorm.io/gorm"
)

// EvalCacheJSON is the JSON serialization format of EvalCache's flags
type EvalCacheJSON struct {
	Flags []entity.Flag
}

func (ec *EvalCache) export() EvalCacheJSON {
	ec.cacheMutex.RLock()
	defer ec.cacheMutex.RUnlock()

	idCache := ec.cache.idCache
	fs := make([]entity.Flag, 0, len(idCache))
	for _, f := range idCache {
		ff := *f
		fs = append(fs, ff)
	}
	return EvalCacheJSON{Flags: fs}
}

func (ec *EvalCache) fetchAllFlags() (idCache map[string]*entity.Flag, keyCache map[string]*entity.Flag, tagCache map[string]map[uint]*entity.Flag, err error) {
	fs, err := fetchAllFlags()
	if err != nil {
		return nil, nil, nil, err
	}

	idCache = make(map[string]*entity.Flag)
	keyCache = make(map[string]*entity.Flag)
	tagCache = make(map[string]map[uint]*entity.Flag)

	for i := range fs {
		f := &fs[i]
		if err := f.PrepareEvaluation(); err != nil {
			return nil, nil, nil, err
		}

		if f.ID != 0 {
			idCache[util.SafeString(f.ID)] = f
		}
		if f.Key != "" {
			keyCache[f.Key] = f
		}
		if f.Tags != nil {
			for _, s := range f.Tags {
				if tagCache[s.Value] == nil {
					tagCache[s.Value] = make(map[uint]*entity.Flag)
				}
				tagCache[s.Value][f.ID] = f
			}
		}
	}
	return idCache, keyCache, tagCache, nil
}

type evalCacheFetcher interface {
	fetch() ([]entity.Flag, error)
}

func newFetcher() (evalCacheFetcher, error) {
	if !config.Config.EvalOnlyMode {
		return &dbFetcher{db: getDB()}, nil
	}

	switch config.Config.DBDriver {
	case "json_file":
		return &jsonFileFetcher{filePath: config.Config.DBConnectionStr}, nil
	case "json_http":
		return &jsonHTTPFetcher{url: config.Config.DBConnectionStr}, nil
	default:
		return nil, fmt.Errorf(
			"failed to create evaluation cache fetcher. DBDriver:%s is not supported",
			config.Config.DBDriver,
		)
	}
}

var fetchAllFlags = func() ([]entity.Flag, error) {
	fetcher, err := newFetcher()
	if err != nil {
		return nil, err
	}
	return fetcher.fetch()
}

type jsonFileFetcher struct {
	filePath string
}

func (ff *jsonFileFetcher) fetch() ([]entity.Flag, error) {
	b, err := os.ReadFile(ff.filePath)
	if err != nil {
		return nil, err
	}
	ecj := &EvalCacheJSON{}
	err = json.Unmarshal(b, ecj)
	if err != nil {
		return nil, err
	}
	return ecj.Flags, nil
}

type jsonHTTPFetcher struct {
	url string
}

func (hf *jsonHTTPFetcher) fetch() ([]entity.Flag, error) {
	client := http.Client{Timeout: config.Config.EvalCacheRefreshTimeout}
	res, err := client.Get(hf.url)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	b, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	ecj := &EvalCacheJSON{}
	err = json.Unmarshal(b, ecj)
	if err != nil {
		return nil, err
	}
	return ecj.Flags, nil
}

type dbFetcher struct {
	db *gorm.DB
}

func (df *dbFetcher) fetch() ([]entity.Flag, error) {
	// Use eager loading to avoid N+1 problem
	// doc: http://jinzhu.me/gorm/crud.html#preloading-eager-loading
	fs := []entity.Flag{}
	err := entity.PreloadSegmentsVariantsTags(df.db).Find(&fs).Error
	return fs, err
}
