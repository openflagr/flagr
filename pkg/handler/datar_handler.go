package handler

import (
	"fmt"
	"sort"

	"time"

	"github.com/go-openapi/runtime/middleware"
	"github.com/go-openapi/strfmt"
	"github.com/openflagr/flagr/swagger_gen/models"
	"github.com/openflagr/flagr/swagger_gen/restapi/operations/datar"
	"github.com/sirupsen/logrus"
)

func datarError(msg string, args ...interface{}) *models.Error {
	m := fmt.Sprintf(msg, args...)
	return &models.Error{Message: &m}
}

func parseTimeRange(from, to *strfmt.DateTime, defaultDays int) (time.Time, time.Time) {
	now := time.Now().UTC()
	end := now
	start := now.Add(-time.Duration(defaultDays) * 24 * time.Hour)
	if from != nil {
		start = time.Time(*from).UTC()
	}
	if to != nil {
		end = time.Time(*to).UTC()
	}
	return start, end
}

// HandleGetDatarSummary is the handler for GET /datar/summary.
func HandleGetDatarSummary(params datar.GetDatarSummaryParams) middleware.Responder {
	d := GetDatar()
	if d == nil {
		return datar.NewGetDatarSummaryDefault(503).WithPayload(
			datarError("Datar is not enabled"),
		)
	}

	from, to := parseTimeRange(params.From, params.To, 7)

	limit := 100
	if params.Limit != nil {
		limit = int(*params.Limit)
	}
	offset := 0
	if params.Offset != nil {
		offset = int(*params.Offset)
	}

	rows, err := d.QuerySummary(from, to, limit, offset)
	if err != nil {
		logrus.WithError(err).Error("Datar: QuerySummary failed")
		return datar.NewGetDatarSummaryDefault(500).WithPayload(
			datarError("query failed: %s", err),
		)
	}
	flags := make([]*models.DatarSummaryFlag, len(rows))
	for i, r := range rows {
		flag := &models.DatarSummaryFlag{
			FlagID:         r.FlagID,
			FlagKey:        r.FlagKey,
			Enabled:        r.Enabled,
			Description:    r.Description,
			TotalEvalCount: r.TotalEvalCount,
		}
		if r.LastEvaluated != "" {
			t, err := time.Parse(time.RFC3339, r.LastEvaluated)
			if err == nil {
				flag.LastEvaluatedAt = strfmt.DateTime(t)
			}
		}
		flags[i] = flag
	}

	return datar.NewGetDatarSummaryOK().WithPayload(
		&models.DatarSummaryResponse{Flags: flags},
	)
}

func HandleGetDatarFlagSummary(params datar.GetDatarFlagSummaryParams) middleware.Responder {
	d := GetDatar()
	if d == nil {
		return datar.NewGetDatarFlagSummaryDefault(503).WithPayload(
			datarError("Datar is not enabled"),
		)
	}

	from, to := parseTimeRange(params.From, params.To, 7)
	flagID := params.FlagID

	rows, err := d.QueryFlagSummary(flagID, from, to)
	if err != nil {
		logrus.WithError(err).Error("Datar: QueryFlagSummary failed")
		return datar.NewGetDatarFlagSummaryDefault(500).WithPayload(
			datarError("query failed: %s", err),
		)
	}

	// Aggregate raw rows into three bucket types in a single pass.
	variantTotals := make(map[int64]int64)
	segIDs := make(map[int64]int64)
	dayCounts := make(map[string]int64)

	for _, r := range rows {
		variantTotals[r.VariantID] += int64(r.EvalCount)

		if r.SegmentID > 0 {
			segIDs[r.SegmentID] += int64(r.EvalCount)
		}

		// BucketHour from the driver is RFC 3339; extract YYYY-MM-DD.
		if len(r.BucketHour) >= 10 {
			dayCounts[r.BucketHour[:10]] += int64(r.EvalCount)
		}
	}

	// Variant entries sorted by count descending.
	variants := make([]*models.DatarVariantEntry, 0, len(variantTotals))
	for id, count := range variantTotals {
		variants = append(variants, &models.DatarVariantEntry{
			VariantID: id,
			Count:     count,
		})
	}
	sort.Slice(variants, func(i, j int) bool {
		return variants[i].Count > variants[j].Count
	})

	// Segment entries sorted by count descending.
	segs := make([]*models.DatarSegmentEntry, 0, len(segIDs))
	for id, count := range segIDs {
		segs = append(segs, &models.DatarSegmentEntry{
			SegmentID: id,
			Count:     count,
		})
	}
	sort.Slice(segs, func(i, j int) bool {
		return segs[i].Count > segs[j].Count
	})
	// Day entries sorted by date.
	days := make([]*models.DatarDayEntry, 0, len(dayCounts))
	for dateStr, count := range dayCounts {
		t, err := time.Parse("2006-01-02", dateStr)
		if err != nil {
			logrus.WithError(err).WithField("date", dateStr).Warn("Datar: invalid date in time bucket")
			continue
		}
		days = append(days, &models.DatarDayEntry{
			Date:  strfmt.Date(t),
			Count: count,
		})
	}
	sort.Slice(days, func(i, j int) bool {
		return time.Time(days[i].Date).Before(time.Time(days[j].Date))
	})

	resp := &models.DatarFlagSummaryResponse{
		FlagID:           flagID,
		TrafficByVariant: variants,
		TrafficBySegment: segs,
		TrafficByDay:     days,
	}

	return datar.NewGetDatarFlagSummaryOK().WithPayload(resp)
}
