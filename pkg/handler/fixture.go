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
			idCache:  map[string]*entity.Flag{util.SafeString(f.Model.ID): &f},
			keyCache: map[string]*entity.Flag{f.Key: &f},
			tagCache: tagCache,
		},
	}

	return ec
}
