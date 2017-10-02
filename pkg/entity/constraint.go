package entity

import "github.com/jinzhu/gorm"

// Constraint is the unit of constraints
type Constraint struct {
	gorm.Model
	SegmentID uint

	Property string
	Operator string
	Value    string
}
