package discord

import (
	"sync"
	"time"
)

// SeedCache manages a cache of recently seen seeds to prevent duplicate notifications
type SeedCache struct {
	cache map[string]time.Time
	mutex sync.RWMutex
	ttl   time.Duration
}

// NewSeedCache creates a new seed cache with the specified TTL
func NewSeedCache(ttl time.Duration) *SeedCache {
	cache := &SeedCache{
		cache: make(map[string]time.Time),
		ttl:   ttl,
	}

	// Start a goroutine to periodically clean expired entries
	go cache.cleanupLoop()

	return cache
}

// cleanupLoop periodically removes expired entries from the cache
func (sc *SeedCache) cleanupLoop() {
	ticker := time.NewTicker(sc.ttl / 2)
	defer ticker.Stop()

	for range ticker.C {
		sc.cleanup()
	}
}

// cleanup removes expired entries from the cache
func (sc *SeedCache) cleanup() {
	sc.mutex.Lock()
	defer sc.mutex.Unlock()

	now := time.Now()
	for key, timestamp := range sc.cache {
		if now.Sub(timestamp) > sc.ttl {
			delete(sc.cache, key)
		}
	}
}

// HasSeen checks if a seed has been seen recently
func (sc *SeedCache) HasSeen(seedKey string) bool {
	sc.mutex.RLock()
	defer sc.mutex.RUnlock()

	_, exists := sc.cache[seedKey]
	return exists
}

// MarkSeen marks a seed as seen
func (sc *SeedCache) MarkSeen(seedKey string) {
	sc.mutex.Lock()
	defer sc.mutex.Unlock()

	sc.cache[seedKey] = time.Now()
}
