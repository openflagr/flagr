package entity

import "github.com/jinzhu/gorm"

// User represents the User struct
type User struct {
	gorm.Model
	Email string `sql:"type:text"`
}
