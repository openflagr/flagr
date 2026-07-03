package handler

import (
	"github.com/openflagr/flagr/pkg/entity"
	"github.com/openflagr/flagr/pkg/util"
)

// GenFixtureEvalCache generates a fixture
func GenFixtureEvalCache() *EvalCache {
	f := entity.GenFixtureFlag()

	tagCache := make(map[string]map[uint]*entity.Flag)
	for _, tag := range f.Tags {
		tagCache[tag.Value] = map[uint]*entity.Flag{f.ID: &f}
	}

	ec := &EvalCache{
		cache: &cacheContainer{
			idCache:  map[string]*entity.Flag{util.SafeString(f.ID): &f},
			keyCache: map[string]*entity.Flag{f.Key: &f},
			tagCache: tagCache,
		},
	}

	return ec
}

// GenFixtureFlagWithTags generates a flag with specific attributes
func GenFixtureFlagWithTags(id uint, key string, enabled bool, tags []string) entity.Flag {
	f := entity.Flag{
		Key:     key,
		Enabled: enabled,
		Tags:    []entity.Tag{},
	}
	f.ID = id
	for _, tag := range tags {
		f.Tags = append(f.Tags, entity.Tag{Value: tag})
	}
	return f
}

// GenFixtureEvalCacheWithFlags generates an EvalCache with multiple flags
func GenFixtureEvalCacheWithFlags(flags []entity.Flag) *EvalCache {
	idCache := make(map[string]*entity.Flag)
	keyCache := make(map[string]*entity.Flag)
	tagCache := make(map[string]map[uint]*entity.Flag)

	for i := range flags {
		f := &flags[i]
		idCache[util.SafeString(f.ID)] = f
		keyCache[f.Key] = f
		for _, tag := range f.Tags {
			if tagCache[tag.Value] == nil {
				tagCache[tag.Value] = make(map[uint]*entity.Flag)
			}
			tagCache[tag.Value][f.ID] = f
		}
	}

	return &EvalCache{
		cache: &cacheContainer{
			idCache:  idCache,
			keyCache: keyCache,
			tagCache: tagCache,
		},
	}
}
