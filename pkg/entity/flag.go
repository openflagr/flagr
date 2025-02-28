package entity

import (
	"fmt"

	"github.com/openflagr/flagr/pkg/util"
	"gorm.io/gorm"
)

// Flag is the unit of flags
type Flag struct {
	gorm.Model

	Key         string `gorm:"type:varchar(64);uniqueIndex:idx_flag_key"`
	Description string `gorm:"type:text"`
	CreatedBy   string
	UpdatedBy   string
	Enabled     bool
	Segments    []Segment
	Variants    []Variant
	Tags        []Tag `gorm:"many2many:flags_tags;"`
	SnapshotID  uint
	Notes       string `gorm:"type:text"`

	DataRecordsEnabled bool
	EntityType         string

	FlagEvaluation FlagEvaluation `gorm:"-" json:"-"`
}

// FlagEvaluation is a struct that holds the necessary info for evaluation
type FlagEvaluation struct {
	VariantsMap map[uint]*Variant
}

// Preloads just the tags
func PreloadFlagTags(db *gorm.DB) *gorm.DB {
	return db.Preload("Tags", func(db *gorm.DB) *gorm.DB {
		return db.Order("id")
	})
}

// PreloadSegmentsVariantsTags preloads segments, variants and tags for flag
func PreloadSegmentsVariantsTags(db *gorm.DB) *gorm.DB {
	return db.
		Preload("Segments", func(db *gorm.DB) *gorm.DB {
			return PreloadConstraintsDistribution(db).
				Order("segments.rank").
				Order("segments.id")
		}).
		Preload("Variants", func(db *gorm.DB) *gorm.DB {
			return db.Order("id")
		}).
		Preload("Tags", func(db *gorm.DB) *gorm.DB {
			return db.Order("id")
		})
}

// Preload preloads the segments, variants and tags into flags
func (f *Flag) Preload(db *gorm.DB) error {
	return PreloadSegmentsVariantsTags(db).First(f, f.Model.ID).Error
}

// PreloadTags preloads the tags into flags
func (f *Flag) PreloadTags(db *gorm.DB) error {
	return PreloadFlagTags(db).First(f, f.Model.ID).Error
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
		f.FlagEvaluation.VariantsMap[f.Variants[i].Model.ID] = &f.Variants[i]
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
