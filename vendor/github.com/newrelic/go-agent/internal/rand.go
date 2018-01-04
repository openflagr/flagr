package internal

import (
	"math/rand"
	"sync"
	"time"
)

var (
	seededRand = struct {
		sync.Mutex
		*rand.Rand
	}{
		Rand: rand.New(rand.NewSource(int64(time.Now().UnixNano()))),
	}
)

// RandUint64 returns a random uint64.
//
// IMPORTANT! The default rand package functions are not used, since we want to
// minimize the chance that different Go processes duplicate the same
// transaction id.  (Note that the rand top level functions "use a default
// shared Source that produces a deterministic sequence of values each time a
// program is run" (and we don't seed the shared Source to avoid changing
// customer apps' behavior)).
func RandUint64() uint64 {
	seededRand.Lock()
	defer seededRand.Unlock()

	u1 := seededRand.Uint32()
	u2 := seededRand.Uint32()
	return (uint64(u1) << 32) | uint64(u2)
}

// RandUint32 returns a random uint32.
func RandUint32() uint32 {
	seededRand.Lock()
	defer seededRand.Unlock()

	return seededRand.Uint32()
}
