package entity

import "github.com/jinzhu/gorm"

// Distribution is
type Distribution struct {
	gorm.Model
	SegmentID uint
	VariantID uint

	Rank    uint
	Percent uint // Percent is an uint from 0 to 100, percent is always derived from Bitmap
	Bitmap  string
}
