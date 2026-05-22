package datar

import "time"

// FlushKey identifies one aggregate dimension set.
// Used as the map key in the in-memory aggregator.
type FlushKey struct {
	FlagID    int64
	VariantID int64
	SegmentID int64
	Hour      time.Time
}
