//go:generate goqueryset -in user.go

package entity

import "github.com/jinzhu/gorm"

// User represents the User struct
// gen:qs
type User struct {
	gorm.Model
	Email string `sql:"type:text"`
}
