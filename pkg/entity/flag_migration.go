package entity

import "gorm.io/gorm"

type FlagMigration struct {
	gorm.Model

	Name string `gorm:"type:varchar(64);uniqueIndex:idx_flag_migration_name"`
}
