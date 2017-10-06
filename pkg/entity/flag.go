//go:generate goqueryset -in flag.go

package entity

import (
	"github.com/jinzhu/gorm"
)

// Flag is the unit of flags
// gen:qs
type Flag struct {
	gorm.Model
	Description string `sql:"type:text"`
	CreatedBy   string
	UpdatedBy   string

	Segments []Segment
	Variants []Variant
}

// Preload preloads the segments and variants into flags
func (f *Flag) Preload(db *gorm.DB) error {
	ss := []Segment{}
	segmentQuery := NewSegmentQuerySet(db)
	if err := segmentQuery.FlagIDEq(f.ID).OrderAscByRank().All(&ss); err != nil {
		return err
	}
	for i, s := range ss {
		if err := s.Preload(db); err != nil {
			return err
		}
		ss[i] = s
	}
	f.Segments = ss

	vs := []Variant{}
	variantQuery := NewVariantQuerySet(db)
	err := variantQuery.FlagIDEq(f.ID).OrderAscByID().All(&vs)
	if err != nil {
		return err
	}
	f.Variants = vs

	return nil
}
