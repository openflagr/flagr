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

	summary, err := d.QueryFlagSummaryBreakdown(params.FlagID, from, to)
	if err != nil {
		logrus.WithError(err).Error("Datar: QueryFlagSummaryBreakdown failed")
		return datar.NewGetDatarFlagSummaryDefault(500).WithPayload(
			datarError("query failed: %s", err),
		)
	}

	// Convert engine types to swagger models.
	variants := make([]*models.DatarVariantEntry, len(summary.Variants))
	for i, v := range summary.Variants {
		variants[i] = &models.DatarVariantEntry{
			VariantID: v.VariantID,
			Count:     v.Count,
		}
	}

	segs := make([]*models.DatarSegmentEntry, len(summary.Segments))
	for i, s := range summary.Segments {
		segs[i] = &models.DatarSegmentEntry{
			SegmentID: s.SegmentID,
			Count:     s.Count,
		}
	}

	days := make([]*models.DatarDayEntry, 0, len(summary.Days))
	for _, d := range summary.Days {
		t, err := time.Parse("2006-01-02", d.Date)
		if err != nil {
			logrus.WithError(err).WithField("date", d.Date).Warn("Datar: invalid date in time bucket")
			continue
		}
		days = append(days, &models.DatarDayEntry{
			Date:  strfmt.Date(t),
			Count: d.Count,
		})
	}

	resp := &models.DatarFlagSummaryResponse{
		FlagID:           summary.FlagID,
		TrafficByVariant: variants,
		TrafficBySegment: segs,
		TrafficByDay:     days,
	}

	return datar.NewGetDatarFlagSummaryOK().WithPayload(resp)
}
