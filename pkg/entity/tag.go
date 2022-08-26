package entity

import (
	"gorm.io/gorm"
)

// Tag is a descriptive identifier given to ease searchability
type Tag struct {
	gorm.Model

	Value string  `gorm:"type:varchar(64);uniqueIndex:idx_tag_value"`
	Flags []*Flag `gorm:"many2many:flags_tags;"`
}
