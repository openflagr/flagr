package entity

import "github.com/jinzhu/gorm"

// Experience is the struct of experience
type Experience struct {
	gorm.Model
	FlagID uint

	Key        string
	Attachment map[string]string
}
