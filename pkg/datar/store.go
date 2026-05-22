package datar

import (
	"time"

	"github.com/openflagr/flagr/pkg/entity"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// Store handles DB operations for Datar aggregate data.
type Store struct {
	db        *gorm.DB
	upsertRef string // dialect-specific upsert reference ("excluded.*" or "VALUES(*)")
}

// initStore initializes a Store with dialect-aware upsert reference.
func initStore(db *gorm.DB) *Store {
	ref := "excluded.eval_count"
	if db.Name() == "mysql" {
		ref = "VALUES(eval_count)"
	}
	return &Store{db: db, upsertRef: ref}
}

// NewStore creates a Store using Flagr's main DB connection.
func NewStore() *Store {
	return initStore(entity.GetDB())
}

// NewTestStore creates a Store with the given DB connection for testing.
func NewTestStore(db *gorm.DB) *Store {
	return initStore(db)
}

// FlushAggregates writes the in-memory aggregate snapshot to the DB
// using additive UPSERT. Multiple instances can flush concurrently.
// All rows are written in a single transaction.
func (s *Store) FlushAggregates(agg map[FlushKey]int32) error {
	if len(agg) == 0 {
		return nil
	}
	now := time.Now()
	query := "INSERT INTO datar_hourly_events (flag_id, variant_id, segment_id, bucket_hour, eval_count) VALUES (?, ?, ?, ?, ?) ON CONFLICT(flag_id, variant_id, segment_id, bucket_hour) DO UPDATE SET eval_count = datar_hourly_events.eval_count + " + s.upsertRef + ", updated_at = ?"

	tx := s.db.Begin()
	for k, count := range agg {
		if err := tx.Exec(query, k.FlagID, k.VariantID, k.SegmentID, k.Hour, count, now).Error; err != nil {
			tx.Rollback()
			return err
		}
	}
	return tx.Commit().Error
}

// SummaryRow is a single flag's aggregate summary for the list page.
type SummaryRow struct {
	FlagID          int64      `json:"flagID"`
	FlagKey         string     `json:"flagKey"`
	Enabled         bool       `json:"enabled"`
	Description     string     `json:"description"`
	TotalEvalCount  int64      `json:"totalEvalCount"`
	LastEvaluatedAt *time.Time `json:"lastEvaluatedAt"`
}

// QuerySummary returns all flags with traffic totals in the given time range.
func (s *Store) QuerySummary(from, to time.Time, limit, offset int) ([]SummaryRow, error) {
	var rows []SummaryRow
	err := s.db.Raw(`
		SELECT f.id AS flag_id, f.key AS flag_key, f.enabled, f.description,
			COALESCE(SUM(e.eval_count), 0) AS total_eval_count,
			MAX(e.updated_at) AS last_evaluated_at
		FROM flags f
		LEFT JOIN datar_hourly_events e ON e.flag_id = f.id
			AND e.bucket_hour >= ? AND e.bucket_hour < ?
		WHERE f.deleted_at IS NULL
		GROUP BY f.id
		ORDER BY total_eval_count DESC
		LIMIT ? OFFSET ?
	`, from, to, limit, offset).Scan(&rows).Error
	if err != nil {
		logrus.WithError(err).Error("Datar: QuerySummary failed")
		return nil, err
	}
	return rows, nil
}

// TrafficPoint is one bucket in a time-series by variant.
type TrafficPoint struct {
	BucketHour string `json:"bucket"`
	VariantID  int64  `json:"variantID"`
	Count      int32  `json:"count"`
}

// QueryTraffic returns traffic grouped by variant per time bucket.
func (s *Store) QueryTraffic(flagID int64, from, to time.Time) ([]TrafficPoint, error) {
	var rows []TrafficPoint
	err := s.db.Raw(`
		SELECT e.bucket_hour, e.variant_id, CAST(SUM(e.eval_count) AS INTEGER) AS count
		FROM datar_hourly_events e
		WHERE e.flag_id = ? AND e.bucket_hour >= ? AND e.bucket_hour < ?
		GROUP BY e.bucket_hour, e.variant_id
		ORDER BY e.bucket_hour, e.variant_id
	`, flagID, from, to).Scan(&rows).Error
	if err != nil {
		logrus.WithError(err).Error("Datar: QueryTraffic failed")
		return nil, err
	}
	return rows, nil
}

// SegmentRow is one segment's traffic count.
type SegmentRow struct {
	SegmentID   int64  `json:"segmentID"`
	Description string `json:"description"`
	EvalCount   int64  `json:"evalCount"`
}

// QuerySegments returns segment distribution for a flag.
func (s *Store) QuerySegments(flagID int64, from, to time.Time) ([]SegmentRow, error) {
	var rows []SegmentRow
	err := s.db.Raw(`
		SELECT e.segment_id, COALESCE(s.description, '') AS description, CAST(SUM(e.eval_count) AS INTEGER) AS eval_count
		FROM datar_hourly_events e
		LEFT JOIN segments s ON s.id = e.segment_id
		WHERE e.flag_id = ? AND e.bucket_hour >= ? AND e.bucket_hour < ? AND e.segment_id > 0
		GROUP BY e.segment_id
		ORDER BY eval_count DESC
	`, flagID, from, to).Scan(&rows).Error
	if err != nil {
		logrus.WithError(err).Error("Datar: QuerySegments failed")
		return nil, err
	}
	return rows, nil
}

// TimeBucketRow is a daily time-series count.
type TimeBucketRow struct {
	Bucket string `json:"date"`
	Count  int64  `json:"count"`
}

// QueryTimeBuckets returns daily traffic counts for a flag.
func (s *Store) QueryTimeBuckets(flagID int64, from, to time.Time) ([]TimeBucketRow, error) {
	var rows []TimeBucketRow
	err := s.db.Raw(`
		SELECT DATE(e.bucket_hour) AS bucket, CAST(SUM(e.eval_count) AS INTEGER) AS count
		FROM datar_hourly_events e
		WHERE e.flag_id = ? AND e.bucket_hour >= ? AND e.bucket_hour < ?
		GROUP BY DATE(e.bucket_hour)
		ORDER BY bucket
	`, flagID, from, to).Scan(&rows).Error
	if err != nil {
		logrus.WithError(err).Error("Datar: QueryTimeBuckets failed")
		return nil, err
	}
	return rows, nil
}
