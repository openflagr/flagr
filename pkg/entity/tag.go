package entity

import (
	"github.com/jinzhu/gorm"
)

// Tag is a descriptive identifier given to ease searchability
type Tag struct {
	gorm.Model

	Value string  `sql:"type:varchar(64);unique_index:idx_tag_value"`
	Flags []*Flag `gorm:"many2many:flags_tags;"`
}
