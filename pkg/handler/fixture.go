package handler

import (
	"github.com/checkr/flagr/pkg/entity"
	"github.com/checkr/flagr/pkg/util"
)

// GenFixtureEvalCache generates a fixture
func GenFixtureEvalCache() *EvalCache {
	f := entity.GenFixtureFlag()
	return &EvalCache{
		mapCache: map[string]*entity.Flag{
			util.SafeString(f.ID): &f,
			f.Key: &f,
		},
	}
}
