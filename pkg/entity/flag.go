package entity

import (
	"fmt"

	"github.com/openflagr/flagr/pkg/util"
	"github.com/sirupsen/logrus"
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
	TagValues   []string // denormalized tag values for eval results
	// JevQuestions maps the `@jev.<name>` question name to its definition,
	// collected across segments in rank order. Empty when no Jev constraints exist.
	JevQuestions map[string]JevQuestion
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
	tagValues := make([]string, 0, len(f.Tags))
	for _, tag := range f.Tags {
		tagValues = append(tagValues, tag.Value)
	}
	f.FlagEvaluation = FlagEvaluation{
		VariantsMap:  make(map[uint]*Variant),
		TagValues:    tagValues,
		JevQuestions: make(map[string]JevQuestion),
	}
	for i := range f.Segments {
		if err := f.Segments[i].PrepareEvaluation(); err != nil {
			return err
		}
		f.collectJevQuestions(&f.Segments[i])
	}
	for i := range f.Variants {
		f.FlagEvaluation.VariantsMap[f.Variants[i].ID] = &f.Variants[i]
	}
	return nil
}

// collectJevQuestions records the segment's Jev constraints on the flag's
// evaluation state. Segments are visited in rank order, so the first
// definition of a question name wins if it appears in more than one segment.
func (f *Flag) collectJevQuestions(s *Segment) {
	for i := range s.Constraints {
		c := &s.Constraints[i]
		if !c.IsJev() {
			continue
		}
		name := c.JevName()
		q, err := c.JevQuestion()
		if err != nil {
			logrus.WithError(err).WithFields(logrus.Fields{
				"flagID":       f.ID,
				"segmentID":    s.ID,
				"constraintID": c.ID,
			}).Warn("skipping invalid jev constraint")
			continue
		}
		if _, exists := f.FlagEvaluation.JevQuestions[name]; exists {
			logrus.WithFields(logrus.Fields{
				"flagID":    f.ID,
				"segmentID": s.ID,
				"question":  name,
			}).Warn("duplicate jev question name; keeping the higher-priority definition")
			continue
		}
		f.FlagEvaluation.JevQuestions[name] = *q
	}
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
