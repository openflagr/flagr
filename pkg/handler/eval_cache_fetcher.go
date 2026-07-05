package handler

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"slices"

	"github.com/openflagr/flagr/pkg/config"
	"github.com/openflagr/flagr/pkg/entity"
	"github.com/openflagr/flagr/pkg/util"
	"github.com/openflagr/flagr/swagger_gen/restapi/operations/export"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// EvalCacheJSON is the JSON serialization format of EvalCache's flags
type EvalCacheJSON struct {
	Flags []entity.Flag
}

func (ec *EvalCache) export(query export.GetExportEvalCacheJSONParams) EvalCacheJSON {
	// Build lookup sets for id/key filters (O(1) lookups)
	var targetIDs map[int64]struct{}
	if len(query.Ids) > 0 {
		targetIDs = make(map[int64]struct{}, len(query.Ids))
		for _, id := range query.Ids {
			targetIDs[id] = struct{}{}
		}
	}

	var targetKeys map[string]struct{}
	if len(query.Keys) > 0 {
		targetKeys = make(map[string]struct{}, len(query.Keys))
		for _, key := range query.Keys {
			targetKeys[key] = struct{}{}
		}
	}

	// Build tag filter function
	var hasTags func(*entity.Flag) bool
	if len(query.Tags) > 0 {
		useAll := query.TagsOperator != nil && *query.TagsOperator == "ALL"
		if useAll {
			// ALL: flag must have every tag in query.Tags
			hasTags = func(f *entity.Flag) bool {
				for _, tag := range query.Tags {
					if !slices.ContainsFunc(f.Tags, func(t entity.Tag) bool {
						return t.Value == tag
					}) {
						return false
					}
				}
				return true
			}
		} else {
			// ANY: flag must have at least one tag in query.Tags
			hasTags = func(f *entity.Flag) bool {
				return slices.ContainsFunc(query.Tags, func(tag string) bool {
					return slices.ContainsFunc(f.Tags, func(t entity.Tag) bool {
						return t.Value == tag
					})
				})
			}
		}
	}

	ec.cacheMutex.RLock()
	defer ec.cacheMutex.RUnlock()

	idCache := ec.cache.idCache
	fs := make([]entity.Flag, 0, len(idCache))
	for _, f := range idCache {
		// ids filter: highest precedence, OR within group
		if targetIDs != nil {
			if _, ok := targetIDs[int64(f.ID)]; ok {
				ff := *f
				fs = append(fs, ff)
			}
			continue
		}

		// keys filter: second precedence, OR within group
		if targetKeys != nil {
			if _, ok := targetKeys[f.Key]; ok {
				ff := *f
				fs = append(fs, ff)
			}
			continue
		}

		// enabled filter
		if query.Enabled != nil && *query.Enabled != f.Enabled {
			continue
		}

		// tags filter
		if hasTags != nil && !hasTags(f) {
			continue
		}

		ff := *f
		fs = append(fs, ff)
	}
	return EvalCacheJSON{Flags: fs}
}

// loadAndBuildCaches fetches all flags from the configured fetcher and builds
// the three lookup caches (idCache, keyCache, tagCache) used by the EvalCache.
func (ec *EvalCache) loadAndBuildCaches() (idCache map[string]*entity.Flag, keyCache map[string]*entity.Flag, tagCache map[string]map[uint]*entity.Flag, err error) {
	fs, err := ec.getFetcher().fetch()
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
	return unmarshalFlags(b)
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
	return unmarshalFlags(b)
}

// unmarshalFlags parses JSON bytes into EvalCacheJSON and returns the flags.
// It auto-assigns IDs to any entities with zero IDs, which is essential for
// hand-edited JSON files where picking unique IDs for every entity is impractical.
//
// Validation warnings are logged but do not prevent loading — the system is
// lenient to allow incremental flag authoring. Validation errors, however,
// DO prevent loading: a flag definition with broken references or missing
// required fields would produce incorrect evaluation results.
func unmarshalFlags(b []byte) ([]entity.Flag, error) {
	ecj := &EvalCacheJSON{}
	if err := json.Unmarshal(b, ecj); err != nil {
		return nil, err
	}

	// Validate after parsing — operates on entity structs directly,
	// giving actionable warnings for hand-edited files.
	result := ValidateFlags(ecj.Flags)
	if !result.OK() {
		for _, e := range result.Errors {
			logrus.Errorf("flag validation error: %s", e)
		}
		return nil, fmt.Errorf("flag validation failed with %d error(s)", len(result.Errors))
	}
	for _, w := range result.Warnings {
		logrus.Warnf("flag validation warning: %s", w)
	}

	normalizeIDs(ecj.Flags)
	return ecj.Flags, nil
}

// setIfZeroAndBumpNext evaluates *target: if zero, sets it to next and
// returns next+1 (to advance the counter); otherwise returns next unchanged.
// Used by normalizeIDs to auto-assign sequential IDs to zero-valued entities.
func setIfZeroAndBumpNext(target *uint, next uint) uint {
	if *target == 0 {
		*target = next
		return next + 1
	}
	return next
}

// normalizeIDs assigns sequential IDs to any entities with zero IDs.
// This allows hand-edited JSON files to omit IDs entirely — the system
// auto-generates unique ones. Entities with explicit non-zero IDs are
// left untouched.
//
// All entity types use global counters (not per-flag) to match the
// behavior of a real database where every table has its own auto-increment.
// This also means IDs are stable if a flag is later migrated to a DB backend.
//
// Invariants:
//   - Every entity type has globally unique IDs
//   - Distribution.VariantID matches a Variant.ID in the same flag
//   - Segment.FlagID, Constraint.SegmentID, Distribution.SegmentID are set
func normalizeIDs(flags []entity.Flag) {
	// Pass 1: find the max existing ID per type so we never collide
	var nextFlagID, nextVariantID, nextSegmentID, nextConstraintID, nextDistributionID, nextTagID uint = 1, 1, 1, 1, 1, 1
	for i := range flags {
		if flags[i].ID >= nextFlagID {
			nextFlagID = flags[i].ID + 1
		}
		for _, v := range flags[i].Variants {
			if v.ID >= nextVariantID {
				nextVariantID = v.ID + 1
			}
		}
		for _, s := range flags[i].Segments {
			if s.ID >= nextSegmentID {
				nextSegmentID = s.ID + 1
			}
			for _, c := range s.Constraints {
				if c.ID >= nextConstraintID {
					nextConstraintID = c.ID + 1
				}
			}
			for _, d := range s.Distributions {
				if d.ID >= nextDistributionID {
					nextDistributionID = d.ID + 1
				}
			}
		}
		for _, t := range flags[i].Tags {
			if t.ID >= nextTagID {
				nextTagID = t.ID + 1
			}
		}
	}

	// Pass 2: assign IDs where missing and fix parent references
	for i := range flags {
		nextFlagID = setIfZeroAndBumpNext(&flags[i].ID, nextFlagID)
		for j := range flags[i].Variants {
			nextVariantID = setIfZeroAndBumpNext(&flags[i].Variants[j].ID, nextVariantID)
		}
		for j := range flags[i].Segments {
			nextSegmentID = setIfZeroAndBumpNext(&flags[i].Segments[j].ID, nextSegmentID)
			flags[i].Segments[j].FlagID = flags[i].ID
			for k := range flags[i].Segments[j].Constraints {
				nextConstraintID = setIfZeroAndBumpNext(&flags[i].Segments[j].Constraints[k].ID, nextConstraintID)
				flags[i].Segments[j].Constraints[k].SegmentID = flags[i].Segments[j].ID
			}
			for k := range flags[i].Segments[j].Distributions {
				d := &flags[i].Segments[j].Distributions[k]
				nextDistributionID = setIfZeroAndBumpNext(&d.ID, nextDistributionID)
				d.SegmentID = flags[i].Segments[j].ID
				// Resolve VariantID from VariantKey when VariantID is missing.
				// This lets hand-edited files omit numeric variant IDs entirely —
				// just set "VariantKey": "control" and the link is resolved.
				if d.VariantID == 0 && d.VariantKey != "" {
					found := false
					for _, v := range flags[i].Variants {
						if v.Key == d.VariantKey {
							d.VariantID = v.ID
							found = true
							break
						}
					}
					if !found {
						logrus.Warnf("flag %q: distribution references unknown variant key %q", flags[i].Key, d.VariantKey)
					}
				}
			}
		}
		for j := range flags[i].Tags {
			nextTagID = setIfZeroAndBumpNext(&flags[i].Tags[j].ID, nextTagID)
		}
	}
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
