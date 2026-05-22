package datar

import (
	"time"

	"github.com/openflagr/flagr/pkg/entity"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// Store handles DB operations for Datar aggregate data.
type Store struct {
	db        *gorm.DB
	upsertRef string // "EXCLUDED" for PG/SQLite, "VALUES" for MySQL
}

// initStore initializes a Store with dialect-aware upsert reference prefix.
func initStore(db *gorm.DB) *Store {
	ref := "EXCLUDED"
	if db.Name() == "mysql" {
		ref = "VALUES"
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
// All rows are written in a single batch via GORM's Clauses + Create.
func (s *Store) FlushAggregates(agg map[FlushKey]int32) error {
	if len(agg) == 0 {
		return nil
	}
	now := time.Now()
	records := make([]entity.HourlyEvent, 0, len(agg))
	for k, count := range agg {
		records = append(records, entity.HourlyEvent{
			FlagID:     k.FlagID,
			VariantID:  k.VariantID,
			SegmentID:  k.SegmentID,
			BucketHour: k.Hour,
			EvalCount:  count,
			UpdatedAt:  now,
		})
	}
	return s.db.Clauses(clause.OnConflict{
		Columns: []clause.Column{
			{Name: "flag_id"},
			{Name: "variant_id"},
			{Name: "segment_id"},
			{Name: "bucket_hour"},
		},
		DoUpdates: clause.Set{{
			Column: clause.Column{Name: "eval_count"},
			Value:  gorm.Expr("datar_hourly_events.eval_count + " + s.upsertRef + ".eval_count"),
		}, {
			Column: clause.Column{Name: "updated_at"},
			Value:  gorm.Expr(s.upsertRef + ".updated_at"),
		}},
	}).Create(&records).Error
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

