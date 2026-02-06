package entity

import (
	"time"
)

// EvalContext represents the context we can evaluate and log
type EvalContext struct {
	EntityID      string
	EntityType    string
	EntityContext map[string]any

	EnableDebug bool
}

// EvalResult represents the result of the evaluation
type EvalResult struct {
	FlagID      uint
	SegmentID   uint
	VariantID   uint
	EvalContext EvalContext
	Timestamp   time.Time

	EvalDebugLog *EvalDebugLog
}

// EvalDebugLog is the debugging log of evaluation
type EvalDebugLog struct {
	SegmentDebugLogs []SegmentDebugLog
	Msg              string
}

// SegmentDebugLog is the segmebnt level of debugging logs
type SegmentDebugLog struct {
	SegmentID uint
	Msg       string
}
