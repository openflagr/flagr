package handler

import (
	"github.com/checkr/flagr/pkg/entity"
	"github.com/checkr/flagr/pkg/util"
)

// GenFixtureEvalCache generates a fixture
func GenFixtureEvalCache() *EvalCache {
	f := entity.GenFixtureFlag()
	return &EvalCache{
		idCache:  map[string]*entity.Flag{util.SafeString(f.ID): &f},
		keyCache: map[string]*entity.Flag{f.Key: &f},
	}
}
