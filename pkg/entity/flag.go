package entity

import (
	"fmt"
	"time"

	"github.com/Allen-Career-Institute/flagr/pkg/util"
	"gorm.io/gorm"
)

// FlagTag represents the join table between flags and tags with timestamps
type FlagTag struct {
	FlagID    uint `gorm:"primaryKey"`
	TagID     uint `gorm:"primaryKey"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

// TableName specifies the table name for FlagTag to match the many-to-many relationship
func (FlagTag) TableName() string {
	return "flags_tags"
}

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
	Tags        []Tag `gorm:"many2many:flags_tags;joinForeignKey:FlagID;joinReferences:TagID;"`
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
	return PreloadSegmentsVariantsTags(db).First(f, f.ID).Error
}

// PreloadTags preloads the tags into flags
func (f *Flag) PreloadTags(db *gorm.DB) error {
	return PreloadFlagTags(db).First(f, f.ID).Error
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

// AddTagToFlag adds a tag to a flag
func AddTagToFlag(db *gorm.DB, flagID, tagID uint) error {
	// Check if association already exists
	var existingAssoc FlagTag
	result := db.Where("flag_id = ? AND tag_id = ?", flagID, tagID).First(&existingAssoc)

	if result.Error != nil && result.Error != gorm.ErrRecordNotFound {
		return result.Error
	}

	if result.Error == gorm.ErrRecordNotFound {
		// Create new association
		flagTag := FlagTag{
			FlagID: flagID,
			TagID:  tagID,
		}
		return db.Create(&flagTag).Error
	}

	// Association already exists, no need to update
	return nil
}

// RemoveTagFromFlag removes a tag from a flag
func RemoveTagFromFlag(db *gorm.DB, flagID, tagID uint) error {
	return db.Where("flag_id = ? AND tag_id = ?", flagID, tagID).Delete(&FlagTag{}).Error
}
