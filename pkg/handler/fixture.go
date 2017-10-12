package handler

import "github.com/checkr/flagr/pkg/entity"

// GenFixtureEvalCache generates a fixture
func GenFixtureEvalCache() *EvalCache {
	f := entity.GenFixtureFlag()
	return &EvalCache{
		mapCache: map[uint]*entity.Flag{100: &f},
	}
}
