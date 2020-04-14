package entity

import (
	"fmt"

	"github.com/checkr/flagr/pkg/util"
	"github.com/jinzhu/gorm"
)

// Flag is the unit of flags
type Flag struct {
	gorm.Model

	Key         string `gorm:"type:varchar(64);unique_index:idx_flag_key"`
	Description string `sql:"type:text"`
	CreatedBy   string
	UpdatedBy   string
	Enabled     bool
	Segments    []Segment
	Variants    []Variant
	SnapshotID  uint
	Notes       string `sql:"type:text"`

	DataRecordsEnabled bool
	EntityType         string

	AuthorizedUsers  string `sql:"type:text"`
	AuthorizedGroups string `sql:"type:text"`

	FlagEvaluation FlagEvaluation `gorm:"-" json:"-"`

}

// FlagEvaluation is a struct that holds the necessary info for evaluation
type FlagEvaluation struct {
	VariantsMap map[uint]*Variant
}

// PreloadSegmentsVariants preloads segments and variants for flag
func PreloadSegmentsVariants(db *gorm.DB) *gorm.DB {
	return db.
		Preload("Segments", func(db *gorm.DB) *gorm.DB {
			return PreloadConstraintsDistribution(db).
				Order("rank ASC").
				Order("id ASC")
		}).
		Preload("Variants", func(db *gorm.DB) *gorm.DB {
			return db.Order("id ASC")
		})
}

// Preload preloads the segments and variants into flags
func (f *Flag) Preload(db *gorm.DB) error {
	return PreloadSegmentsVariants(db).First(f, f.ID).Error
}

// PrepareEvaluation prepares the information for evaluation
func (f *Flag) PrepareEvaluation() error {
	f.FlagEvaluation = FlagEvaluation{
		VariantsMap: make(map[uint]*Variant),
	}
	for i := range f.Segments {
		if err := f.Segments[i].PrepareEvaluation(); err != nil {
			return err
		}
	}
	for i := range f.Variants {
		f.FlagEvaluation.VariantsMap[f.Variants[i].ID] = &f.Variants[i]
	}
	return nil
}

// CreateFlagKey creates the key based on the given key
func CreateFlagKey(key string) (string, error) {
	if key == "" {
		key = util.NewSecureRandomKey()
	} else {
		ok, reason := util.IsSafeKey(key)
		if !ok {
			return "", fmt.Errorf("cannot create flag due to invalid key. reason: %s", reason)
		}
	}
	return key, nil
}

// CreateFlagEntityType creates the FlagEntityType if not exists
func CreateFlagEntityType(db *gorm.DB, key string) error {
	ok, reason := util.IsSafeKey(key)
	if !ok && key != "" {
		return fmt.Errorf("invalid DataRecordsEntityType. reason: %s", reason)
	}
	d := FlagEntityType{Key: key}
	return db.Where(d).FirstOrCreate(&d).Error
}
