package handler

import (
	"fmt"
	"slices"
	"sync"
	"time"

	"github.com/go-openapi/runtime/middleware"
	"github.com/go-openapi/strfmt"
	"github.com/openflagr/flagr/pkg/config"
	"github.com/openflagr/flagr/pkg/datar"
	"github.com/openflagr/flagr/swagger_gen/models"
	datarapi "github.com/openflagr/flagr/swagger_gen/restapi/operations/datar"
	"github.com/sirupsen/logrus"
)

// ---------------------------------------------------------------------------
// Singleton engine lifecycle
// ---------------------------------------------------------------------------

var (
	singletonEngine   *datar.Engine
	singletonEngineMu sync.Mutex
)

// GetDatar returns the singleton datar.Engine.
// Creates the instance on first call, starting its flush loop.
// Returns nil if Datar is not enabled.
func GetDatar() *datar.Engine {
	singletonEngineMu.Lock()
	defer singletonEngineMu.Unlock()
	if singletonEngine != nil {
		return singletonEngine
	}
	if !config.Config.RecorderEnabled || !slices.Contains(config.Config.RecorderType, "datar") {
		return nil
	}
	singletonEngine = datar.New(
		getDB(),
		true,
		config.Config.RecorderDatarFlushInterval,
	)
	return singletonEngine
}

// ResetDatar clears the singleton for test isolation.
func ResetDatar() {
	singletonEngineMu.Lock()
	defer singletonEngineMu.Unlock()
	if singletonEngine != nil {
		singletonEngine.Shutdown()
		singletonEngine = nil
	}
}

const defaultLookbackDays = 7

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

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

// toSwaggerSummaryFlag converts an engine SummaryRow to a swagger model.
func toSwaggerSummaryFlag(r datar.SummaryRow) *models.DatarSummaryFlag {
	f := &models.DatarSummaryFlag{
		FlagID:         r.FlagID,
		FlagKey:        r.FlagKey,
		Enabled:        r.Enabled,
		Description:    r.Description,
		TotalEvalCount: r.TotalEvalCount,
	}
	if r.LastEvaluated != "" {
		if t, err := time.Parse(time.RFC3339, r.LastEvaluated); err == nil {
			f.LastEvaluatedAt = strfmt.DateTime(t)
		}
	}
	return f
}
// toSwaggerDay converts an engine DayEntry to a swagger model.
// Returns nil if the date string is unparseable.
func toSwaggerDay(d datar.DayEntry) *models.DatarDayEntry {
	t, err := time.Parse("2006-01-02", d.Day)
	if err != nil {
		logrus.WithError(err).WithField("date", d.Day).Warn("Datar: invalid date in time bucket")
		return nil
	}
	return &models.DatarDayEntry{Date: strfmt.Date(t), Count: d.Count}
}

// ---------------------------------------------------------------------------
// HTTP handlers
// ---------------------------------------------------------------------------

// HandleGetDatarSummary is the handler for GET /datar/summary.
func HandleGetDatarSummary(params datarapi.GetDatarSummaryParams) middleware.Responder {
	d := GetDatar()
	if d == nil {
		return datarapi.NewGetDatarSummaryDefault(503).WithPayload(
			datarError("Datar is not enabled"),
		)
	}

	from, to := parseTimeRange(params.From, params.To, defaultLookbackDays)

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
		return datarapi.NewGetDatarSummaryDefault(500).WithPayload(
			datarError("query failed: %s", err),
		)
	}

	flags := make([]*models.DatarSummaryFlag, len(rows))
	for i, r := range rows {
		flags[i] = toSwaggerSummaryFlag(r)
	}

	return datarapi.NewGetDatarSummaryOK().WithPayload(
		&models.DatarSummaryResponse{Flags: flags},
	)
}

// HandleGetDatarFlagSummary is the handler for GET /datar/flags/{flagID}/summary.
func HandleGetDatarFlagSummary(params datarapi.GetDatarFlagSummaryParams) middleware.Responder {
	d := GetDatar()
	if d == nil {
		return datarapi.NewGetDatarFlagSummaryDefault(503).WithPayload(
			datarError("Datar is not enabled"),
		)
	}

	from, to := parseTimeRange(params.From, params.To, defaultLookbackDays)

	summary, err := d.QueryFlagSummaryBreakdown(params.FlagID, from, to)
	if err != nil {
		logrus.WithError(err).Error("Datar: QueryFlagSummaryBreakdown failed")
		return datarapi.NewGetDatarFlagSummaryDefault(500).WithPayload(
			datarError("query failed: %s", err),
		)
	}

	variants := make([]*models.DatarVariantEntry, len(summary.Variants))
	for i, v := range summary.Variants {
		variants[i] = &models.DatarVariantEntry{VariantID: v.VariantID, Count: v.Count}
	}

	segs := make([]*models.DatarSegmentEntry, len(summary.Segments))
	for i, s := range summary.Segments {
		segs[i] = &models.DatarSegmentEntry{SegmentID: s.SegmentID, Count: s.Count}
	}

	days := make([]*models.DatarDayEntry, 0, len(summary.Days))
	for _, d := range summary.Days {
		if entry := toSwaggerDay(d); entry != nil {
			days = append(days, entry)
		}
	}

	return datarapi.NewGetDatarFlagSummaryOK().WithPayload(&models.DatarFlagSummaryResponse{
		FlagID:           summary.FlagID,
		TrafficByVariant: variants,
		TrafficBySegment: segs,
		TrafficByDay:     days,
	})
}
