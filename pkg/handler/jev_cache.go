package handler

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"sync"
	"time"

	"github.com/openflagr/flagr/pkg/config"
	"github.com/openflagr/flagr/pkg/entity"
)

// jevCacheEntry is a cached System One answer set.
type jevCacheEntry struct {
	answers map[string]JevAnswer
	expires time.Time
}

// jevAnswerCache is a bounded TTL cache keyed by (model, state, questions).
// It keeps repeated evaluations of the same entity from re-calling System One.
type jevAnswerCache struct {
	mu      sync.Mutex
	entries map[string]jevCacheEntry
	maxSize int
	ttl     time.Duration
}

var (
	singletonJevCache     *jevAnswerCache
	singletonJevCacheOnce sync.Once
)

// GetJevCache returns the process-wide Jev answer cache.
var GetJevCache = func() *jevAnswerCache {
	singletonJevCacheOnce.Do(func() {
		singletonJevCache = newJevAnswerCache()
	})
	return singletonJevCache
}

func newJevAnswerCache() *jevAnswerCache {
	return &jevAnswerCache{
		entries: make(map[string]jevCacheEntry),
		maxSize: config.Config.JevCacheSize,
		ttl:     config.Config.JevCacheTTL,
	}
}

// ResetJevCache drops the process-wide cache. Test helper.
func ResetJevCache() {
	singletonJevCacheOnce = sync.Once{}
	singletonJevCache = nil
}

func (c *jevAnswerCache) Get(key string) (map[string]JevAnswer, bool) {
	if c == nil || c.maxSize <= 0 || key == "" {
		return nil, false
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	e, ok := c.entries[key]
	if !ok {
		return nil, false
	}
	if time.Now().After(e.expires) {
		delete(c.entries, key)
		return nil, false
	}
	return e.answers, true
}

func (c *jevAnswerCache) Set(key string, answers map[string]JevAnswer) {
	if c == nil || c.maxSize <= 0 || key == "" || len(answers) == 0 {
		return
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	if len(c.entries) >= c.maxSize {
		c.evictLocked()
	}
	c.entries[key] = jevCacheEntry{answers: answers, expires: time.Now().Add(c.ttl)}
}

func (c *jevAnswerCache) evictLocked() {
	now := time.Now()
	for k, e := range c.entries {
		if now.After(e.expires) {
			delete(c.entries, k)
		}
	}
	for k := range c.entries {
		if len(c.entries) < c.maxSize {
			break
		}
		delete(c.entries, k)
	}
}

// jevCacheKey hashes the model, state, and questions into a stable key.
func jevCacheKey(state any, questions map[string]entity.JevConstraintSpec) (string, error) {
	stateJSON, err := json.Marshal(state)
	if err != nil {
		return "", err
	}
	questionsJSON, err := json.Marshal(questions)
	if err != nil {
		return "", err
	}
	h := sha256.New()
	h.Write([]byte(config.Config.JevModel))
	h.Write([]byte{0})
	h.Write(stateJSON)
	h.Write([]byte{0})
	h.Write(questionsJSON)
	return hex.EncodeToString(h.Sum(nil)), nil
}
