package entity

import "gorm.io/gorm"

// User represents the User struct
type User struct {
	gorm.Model
	Email string `gorm:"type:text"`
}
