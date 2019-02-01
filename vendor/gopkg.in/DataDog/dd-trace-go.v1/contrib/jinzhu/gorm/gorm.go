// Package gorm provides helper functions for tracing the jinzhu/gorm package (https://github.com/jinzhu/gorm).
package gorm

import (
	sqltraced "gopkg.in/DataDog/dd-trace-go.v1/contrib/database/sql"

	"github.com/jinzhu/gorm"
)

// Open opens a new (traced) database connection. The used dialect must be formerly registered
// using (gopkg.in/DataDog/dd-trace-go.v1/contrib/database/sql).Register.
func Open(dialect, source string) (*gorm.DB, error) {
	db, err := sqltraced.Open(dialect, source)
	if err != nil {
		return nil, err
	}
	return gorm.Open(dialect, db)
}
