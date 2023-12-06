package entity

import (
	"fmt"
	"hash/crc32"
	"sort"

	"gorm.io/gorm"
)

const (
	// TotalBucketNum represents how many buckets we can use to determine the consistent hashing
	// distribution and rollout
	TotalBucketNum uint = 1000

	// PercentMultiplier implies that the multiplier between percentage (100) and TotalBucketNum
	PercentMultiplier uint = TotalBucketNum / uint(100)
)

// Distribution is the struct represents distribution under segment and links to variant
type Distribution struct {
	gorm.Model
	SegmentID  uint `gorm:"index:idx_distribution_segmentid"`
	VariantID  uint `gorm:"index:idx_distribution_variantid"`
	VariantKey string

	Percent uint   // Percent is an uint from 0 to 100, percent is always derived from Bitmap
	Bitmap  string `gorm:"type:text" json:"-"`
}

// DistributionArray is useful for faster evalution
type DistributionArray struct {
	VariantIDs          []uint
	PercentsAccumulated []int // useful for binary search to find the rollout variant
}

// DistributionDebugLog is useful for making debug logs
type DistributionDebugLog struct {
	BucketNum         uint
	DistributionArray DistributionArray
	VariantID         uint
	RolloutPercent    uint
}

// Rollout rolls out the entity based on the rolloutPercent
func (d DistributionArray) Rollout(entityID string, salt string, rolloutPercent uint) (variantID *uint, msg string) {
	if entityID == "" {
		return nil, "rollout no. empty entityID"
	}

	if rolloutPercent == uint(0) {
		return nil, "rollout no. 0% rolloutPercent"
	}

	if len(d.VariantIDs) == 0 || len(d.PercentsAccumulated) == 0 {
		return nil, "rollout no. there's no distribution set"
	}

	num := crc32Num(entityID, salt)
	vID, index := d.bucketByNum(num)
	log := fmt.Sprintf("%+v", DistributionDebugLog{
		BucketNum:         num,
		DistributionArray: d,
		VariantID:         vID,
		RolloutPercent:    rolloutPercent,
	})

	if d.rollout(num, rolloutPercent, index) {
		return &vID, "rollout yes. " + log
	}
	return nil, "rollout no. " + log
}

func (d DistributionArray) bucketByNum(bucketNum uint) (variantID uint, index int) {
	index = sort.SearchInts(d.PercentsAccumulated, int(bucketNum)+1)
	return d.VariantIDs[index], index
}

func (d DistributionArray) rollout(bucketNum uint, rolloutPercent uint, index int) bool {
	if rolloutPercent == uint(0) {
		return false
	}
	if rolloutPercent == uint(100) {
		return true
	}

	min := 0
	max := d.PercentsAccumulated[index]
	r := 0
	if index != 0 {
		min = d.PercentsAccumulated[index-1]
	}
	if max-min-1 > 0 {
		r = max - min - 1
	}
	return 100*(bucketNum-uint(min)) <= uint(r)*rolloutPercent
}

func crc32Num(entityID string, salt string) uint {
	// crc32 is good in terms of uniform distribution
	// http://michiel.buddingh.eu/distribution-of-hash-values
	return uint(crc32.ChecksumIEEE([]byte(salt+entityID))) % TotalBucketNum
}
