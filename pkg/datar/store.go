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
	LastEvaluatedAt string `json:"lastEvaluatedAt"`
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
type RawEvent struct {
	VariantID int64
	SegmentID int64
	BucketHour string
	EvalCount  int32
}

func (s *Store) QueryFlagSummary(flagID int64, from, to time.Time) ([]RawEvent, error) {
	var rows []RawEvent
	err := s.db.Table("datar_hourly_events").
		Select("variant_id, segment_id, bucket_hour, eval_count").
		Where("flag_id = ? AND bucket_hour >= ? AND bucket_hour < ?", flagID, from, to).
		Scan(&rows).Error
	if err != nil {
		logrus.WithError(err).Error("Datar: QueryFlagSummary failed")
		return nil, err
	}
	return rows, nil
}

// SegmentDescriptions returns descriptions for the given segment IDs.
func (s *Store) SegmentDescriptions(ids []int64) (map[int64]string, error) {
	if len(ids) == 0 {
		return nil, nil
	}
	type segRow struct {
		ID          int64
		Description string
	}
	var rows []segRow
	err := s.db.Table("segments").
		Select("id, description").
		Where("id IN ?", ids).
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	descs := make(map[int64]string, len(rows))
	for _, r := range rows {
		descs[r.ID] = r.Description
	}
	return descs, nil
}
