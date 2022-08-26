package entity

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSegmentPrepareEvaluation(t *testing.T) {
	t.Run("happy code path", func(t *testing.T) {
		s := GenFixtureSegment()
		assert.NoError(t, s.PrepareEvaluation())
		assert.NotNil(t, s.SegmentEvaluation.ConditionsExpr)
		assert.NotNil(t, s.SegmentEvaluation.DistributionArray)
	})

	t.Run("error code path", func(t *testing.T) {
		s := GenFixtureSegment()
		s.SegmentEvaluation = SegmentEvaluation{}
		s.Constraints[0].Value = `"CA"]` // invalid value
		assert.Error(t, s.PrepareEvaluation())
		assert.Empty(t, s.SegmentEvaluation.ConditionsExpr)
		assert.Empty(t, s.SegmentEvaluation.DistributionArray.VariantIDs)
		assert.Empty(t, s.SegmentEvaluation.DistributionArray.PercentsAccumulated)
	})
}

func TestSegmentPreload(t *testing.T) {
	t.Run("happy code path", func(t *testing.T) {
		s := GenFixtureSegment()
		f := GenFixtureFlag()
		db := PopulateTestDB(f)

		tmpDB, dbErr := db.DB()
		if dbErr != nil {
			t.Errorf("Failed to get database")
		}

		defer tmpDB.Close()

		err := s.Preload(db)
		assert.NoError(t, err)
	})
}
