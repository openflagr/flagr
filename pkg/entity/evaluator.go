package entity

import (
	"time"

	"github.com/zhouzhuojie/conditions"
)

// EvalContext represents the context we can evaluate and log
type EvalContext struct {
	EntityID      string
	EntityType    string
	EntityContext map[string]interface{}

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

// SegmentEvaluation is a struct that holds the necessary info for evaluation
type SegmentEvaluation struct {
	Conditions      []conditions.Expr
	DistributionMap map[uint]uint
}

// PrepareEvaluation prepares the segment for evaluation by parsing constraints
// and denormalize distributions
func (s *Segment) PrepareEvaluation() error {
	se := SegmentEvaluation{
		Conditions:      make([]conditions.Expr, len(s.Constraints), len(s.Constraints)),
		DistributionMap: make(map[uint]uint),
	}

	for i, c := range s.Constraints {
		expr, err := c.ToExpr()
		if err != nil {
			return err
		}
		se.Conditions[i] = expr
	}
	for _, d := range s.Distributions {
		se.DistributionMap[d.VariantID] = d.Percent
	}

	s.SegmentEvaluation = se
	return nil
}
