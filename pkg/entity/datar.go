package entity

import "time"

// HourlyEvent represents one aggregate row of evaluation counts per hour.
// The natural key is (flag_id, variant_id, segment_id, bucket_hour).
// A surrogate auto-increment PK allows adding columns later without table rebuild.
type HourlyEvent struct {
	ID         uint      `gorm:"primaryKey;autoIncrement"`
	FlagID     int64     `gorm:"not null;uniqueIndex:idx_datar_hourly"`
	VariantID  int64     `gorm:"not null;default:0;uniqueIndex:idx_datar_hourly"`
	SegmentID  int64     `gorm:"not null;default:0;uniqueIndex:idx_datar_hourly"`
	BucketHour time.Time `gorm:"not null;type:datetime(3);uniqueIndex:idx_datar_hourly"`
	EvalCount  int32     `gorm:"not null;default:0"`
	UpdatedAt  time.Time `gorm:"autoUpdateTime"`
}

// TableName specifies the table name for GORM.
func (HourlyEvent) TableName() string {
	return "datar_hourly_events"
}
