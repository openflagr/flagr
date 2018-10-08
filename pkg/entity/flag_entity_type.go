//go:generate goqueryset -in flag_entity_type.go

package entity

import "github.com/jinzhu/gorm"

// FlagEntityType is the entity_type that will overwrite into evaluation logs.
// gen:qs
type FlagEntityType struct {
	gorm.Model
	Key string `gorm:"type:varchar(64);unique_index:flag_entity_type_key"`
}
