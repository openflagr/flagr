package entity

import (
	"github.com/jinzhu/gorm"
)

// Tag is a descriptive identifier given to ease searchability
type Tag struct {
	gorm.Model

	Value string  `sql:"type:text" gorm:"unique;not null"`
	Flags []*Flag `gorm:"many2many:flags_tags"`
}
