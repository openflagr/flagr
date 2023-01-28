package handler

import (
	"github.com/openflagr/flagr/pkg/entity"
	"github.com/openflagr/flagr/pkg/util"
)

// GenFixtureEvalCache generates a fixture
func GenFixtureEvalCache() *EvalCache {
	f1 := entity.GenFixtureFlag()

	f2 := entity.GenFixtureFlag()
	f2.Key = "flag_key_101"
	f2.ID = 101
	f2.Tags = []entity.Tag{{Value: "tag2"}}

	ec := &EvalCache{}
	ec.cache.Store(&cacheContainer{
		idCache: map[string]*entity.Flag{
			util.SafeString(f1.ID): &f1,
			util.SafeString(f2.ID): &f2,
		},
		keyCache: map[string]*entity.Flag{
			f1.Key: &f1,
			f2.Key: &f2,
		},
		tagCache: map[string]map[uint]*entity.Flag{
			"tag1": {f1.ID: &f1},
			"tag2": {f1.ID: &f1, f2.ID: &f2},
		},
	})

	return ec
}
