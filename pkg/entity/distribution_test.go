package entity

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCRC32(t *testing.T) {
	num1 := crc32Num("entity1", "salt1")
	num2 := crc32Num("entity2", "salt1")
	num3 := crc32Num("entity1", "salt1")
	assert.Equal(t, num1, num3)
	assert.NotEqual(t, num1, num2)
}

func TestBucketByNum(t *testing.T) {
	t.Run("normal cases", func(t *testing.T) {
		d := DistributionArray{
			VariantIDs:          []uint{1111, 2222, 3333},
			PercentsAccumulated: []int{340, 340 + 330, 340 + 330 + 330},
		}

		var vID uint
		var index int

		vID, index = d.bucketByNum(0)
		assert.Equal(t, vID, uint(1111))
		assert.Equal(t, index, 0)

		vID, index = d.bucketByNum(1)
		assert.Equal(t, vID, uint(1111))
		assert.Equal(t, index, 0)

		vID, index = d.bucketByNum(339)
		assert.Equal(t, vID, uint(1111))
		assert.Equal(t, index, 0)

		vID, index = d.bucketByNum(340)
		assert.Equal(t, vID, uint(2222))
		assert.Equal(t, index, 1)

		vID, index = d.bucketByNum(340 + 330 - 1)
		assert.Equal(t, vID, uint(2222))
		assert.Equal(t, index, 1)

		vID, index = d.bucketByNum(340 + 330)
		assert.Equal(t, vID, uint(3333))
		assert.Equal(t, index, 2)

		vID, index = d.bucketByNum(999)
		assert.Equal(t, vID, uint(3333))
		assert.Equal(t, index, 2)
	})

	t.Run("0/1000 cases", func(t *testing.T) {
		d := DistributionArray{
			VariantIDs:          []uint{1111, 2222},
			PercentsAccumulated: []int{0, 1000},
		}
		var vID uint
		var index int

		vID, index = d.bucketByNum(0)
		assert.Equal(t, vID, uint(2222))
		assert.Equal(t, index, 1)

		vID, index = d.bucketByNum(1)
		assert.Equal(t, vID, uint(2222))
		assert.Equal(t, index, 1)

		vID, index = d.bucketByNum(999)
		assert.Equal(t, vID, uint(2222))
		assert.Equal(t, index, 1)
	})

	t.Run("1000/0 cases", func(t *testing.T) {
		d := DistributionArray{
			VariantIDs:          []uint{1111, 2222},
			PercentsAccumulated: []int{1000, 1000},
		}
		var vID uint
		var index int

		vID, index = d.bucketByNum(0)
		assert.Equal(t, vID, uint(1111))
		assert.Equal(t, index, 0)

		vID, index = d.bucketByNum(1)
		assert.Equal(t, vID, uint(1111))
		assert.Equal(t, index, 0)

		vID, index = d.bucketByNum(999)
		assert.Equal(t, vID, uint(1111))
		assert.Equal(t, index, 0)
	})

	t.Run("single variant case", func(t *testing.T) {
		d := DistributionArray{
			VariantIDs:          []uint{1111},
			PercentsAccumulated: []int{1000},
		}
		var vID uint
		var index int

		vID, index = d.bucketByNum(0)
		assert.Equal(t, vID, uint(1111))
		assert.Equal(t, index, 0)

		vID, index = d.bucketByNum(1)
		assert.Equal(t, vID, uint(1111))
		assert.Equal(t, index, 0)
	})
}

func TestRollout(t *testing.T) {
	d := DistributionArray{
		VariantIDs:          []uint{1111, 2222},
		PercentsAccumulated: []int{500, 1000},
	}
	assert.Equal(t, d.rollout(uint(0), uint(100), 0), true)
	assert.Equal(t, d.rollout(uint(0), uint(50), 0), true)
	assert.Equal(t, d.rollout(uint(0), uint(1), 0), true)
	assert.Equal(t, d.rollout(uint(0), uint(0), 0), false)

	assert.Equal(t, d.rollout(uint(0), uint(50), 0), true)
	assert.Equal(t, d.rollout(uint(249), uint(50), 0), true)
	assert.Equal(t, d.rollout(uint(250), uint(50), 0), false)
	assert.Equal(t, d.rollout(uint(499), uint(50), 0), false)

	assert.Equal(t, d.rollout(uint(500), uint(50), 1), true)
	assert.Equal(t, d.rollout(uint(749), uint(50), 1), true)
	assert.Equal(t, d.rollout(uint(750), uint(50), 1), false)
	assert.Equal(t, d.rollout(uint(999), uint(50), 1), false)

	assert.Equal(t, d.rollout(uint(0), uint(34), 0), true)
	assert.Equal(t, d.rollout(uint(500*0.34-1), uint(34), 0), true)
	assert.Equal(t, d.rollout(uint(500*0.34), uint(34), 0), false)
}

func TestRolloutWithEntity(t *testing.T) {
	t.Run("normal distributions cases", func(t *testing.T) {
		d := DistributionArray{
			VariantIDs:          []uint{1111, 2222},
			PercentsAccumulated: []int{500, 1000},
		}
		var vID *uint
		var msg string

		vID, msg = d.Rollout("", "salt", uint(0))
		assert.Nil(t, vID)
		assert.Contains(t, msg, "no")

		vID, msg = d.Rollout("entity123", "salt", uint(0))
		assert.Nil(t, vID)
		assert.Contains(t, msg, "no")

		vID, msg = d.Rollout("entity123", "salt", uint(100))
		assert.NotNil(t, vID)
		assert.Contains(t, msg, "yes")

		vID, msg = d.Rollout("entity123", "salt", uint(1))
		assert.Nil(t, vID)
		assert.Contains(t, msg, "no")
	})

	t.Run("empty distributions cases", func(t *testing.T) {
		d := DistributionArray{
			VariantIDs:          []uint{},
			PercentsAccumulated: []int{},
		}
		var vID *uint
		var msg string

		vID, msg = d.Rollout("entity123", "salt", uint(100))
		assert.Nil(t, vID)
		assert.Contains(t, msg, "no")
	})
}
