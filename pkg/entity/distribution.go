//go:generate goqueryset -in distribution.go

package entity

import (
	"github.com/jinzhu/gorm"
)

// Distribution is the struct represents distribution under segment and links to variant
// gen:qs
type Distribution struct {
	gorm.Model
	SegmentID  uint
	VariantID  uint
	VariantKey string

	Percent uint   // Percent is an uint from 0 to 100, percent is always derived from Bitmap
	Bitmap  string `sql:"type:text"`
}
