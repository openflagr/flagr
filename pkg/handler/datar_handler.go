package handler

import (
	"fmt"
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

	rows, err := d.Store().QuerySummary(from, to, limit, offset)
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
		if r.LastEvaluatedAt != "" {
			t, err := time.Parse(time.RFC3339, r.LastEvaluatedAt)
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

// HandleGetDatarFlagSummary is the handler for GET /datar/flags/{flagID}/summary.
func HandleGetDatarFlagSummary(params datar.GetDatarFlagSummaryParams) middleware.Responder {
	d := GetDatar()
	if d == nil {
		return datar.NewGetDatarFlagSummaryDefault(503).WithPayload(
			datarError("Datar is not enabled"),
		)
	}

	from, to := parseTimeRange(params.From, params.To, 7)
	flagID := params.FlagID

	// Get variant traffic.
	trafficRows, err := d.Store().QueryTraffic(flagID, from, to)
	if err != nil {
		logrus.WithError(err).Error("Datar: QueryTraffic failed")
		return datar.NewGetDatarFlagSummaryDefault(500).WithPayload(
			datarError("query failed: %s", err),
		)
	}

	// Get segment distribution.
	segRows, err := d.Store().QuerySegments(flagID, from, to)
	if err != nil {
		logrus.WithError(err).Error("Datar: QuerySegments failed")
		return datar.NewGetDatarFlagSummaryDefault(500).WithPayload(
			datarError("query failed: %s", err),
		)
	}

	// Get daily time buckets.
	dayRows, err := d.Store().QueryTimeBuckets(flagID, from, to)
	if err != nil {
		logrus.WithError(err).Error("Datar: QueryTimeBuckets failed")
		return datar.NewGetDatarFlagSummaryDefault(500).WithPayload(
			datarError("query failed: %s", err),
		)
	}

	// Aggregate variant totals from traffic rows.
	variantTotals := make(map[string]int64)
	for _, p := range trafficRows {
		vk := fmt.Sprintf("%d", p.VariantID)
		variantTotals[vk] += int64(p.Count)
	}

	// Day entries.
	days := make([]*models.DatarDayEntry, len(dayRows))
	for i, dr := range dayRows {
		t, err := time.Parse("2006-01-02", dr.Bucket)
		if err != nil {
			logrus.WithError(err).WithField("date", dr.Bucket).Warn("Datar: invalid date in time bucket")
			continue
		}
		days[i] = &models.DatarDayEntry{
			Date:  strfmt.Date(t),
			Count: dr.Count,
		}
	}

	// Segment entries.
	segs := make([]*models.DatarSegmentEntry, len(segRows))
	for i, sr := range segRows {
		segs[i] = &models.DatarSegmentEntry{
			SegmentID:   sr.SegmentID,
			Description: sr.Description,
			EvalCount:   sr.EvalCount,
		}
	}

	resp := &models.DatarFlagSummaryResponse{
		FlagID:           flagID,
		TrafficByVariant: variantTotals,
		TrafficBySegment: segs,
		TrafficByDay:     days,
	}

	return datar.NewGetDatarFlagSummaryOK().WithPayload(resp)
}
