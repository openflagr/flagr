package handler

import (
	"testing"
	"time"

	"github.com/openflagr/flagr/pkg/entity"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func jevTestResponse() *JevResponse {
	return &JevResponse{
		Model:   "jev-latest",
		Answers: map[string]JevAnswer{"a": {Type: entity.JevTypeNoul}},
	}
}

func TestJevAnswerCacheTTL(t *testing.T) {
	t.Parallel()

	c := &jevAnswerCache{entries: map[string]jevCacheEntry{}, maxSize: 10, ttl: time.Minute}
	c.Set("k", jevTestResponse())
	got, ok := c.Get("k")
	assert.True(t, ok)
	require.NotNil(t, got)
	assert.Contains(t, got.Answers, "a")

	expired := &jevAnswerCache{entries: map[string]jevCacheEntry{}, maxSize: 10, ttl: -time.Second}
	expired.Set("k", jevTestResponse())
	_, ok = expired.Get("k")
	assert.False(t, ok)
}

func TestJevAnswerCacheEviction(t *testing.T) {
	t.Parallel()

	c := &jevAnswerCache{entries: map[string]jevCacheEntry{}, maxSize: 2, ttl: time.Minute}
	c.Set("a", jevTestResponse())
	c.Set("b", jevTestResponse())
	c.Set("c", jevTestResponse())

	c.mu.Lock()
	size := len(c.entries)
	c.mu.Unlock()
	assert.LessOrEqual(t, size, 2)
}

func TestJevAnswerCacheDisabled(t *testing.T) {
	t.Parallel()

	c := &jevAnswerCache{entries: map[string]jevCacheEntry{}, maxSize: 0, ttl: time.Minute}
	c.Set("k", jevTestResponse())
	_, ok := c.Get("k")
	assert.False(t, ok)
}

func TestJevCacheKeyDeterministic(t *testing.T) {
	t.Parallel()

	q := map[string]entity.JevConstraintSpec{
		"a": {Name: "a", Type: entity.JevTypeNoul, Instructions: "x"},
	}
	k1, err := jevCacheKey(map[string]any{"a": 1, "b": 2}, q)
	require.NoError(t, err)
	k2, err := jevCacheKey(map[string]any{"b": 2, "a": 1}, q)
	require.NoError(t, err)
	assert.Equal(t, k1, k2)
}
