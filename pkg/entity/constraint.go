//go:generate goqueryset -in constraint.go

package entity

import "github.com/jinzhu/gorm"

// Constraint is the unit of constraints
// gen:qs
type Constraint struct {
	gorm.Model
	SegmentID uint

	Property string
	Operator string
	Value    string `sql:"type:text"`
}
