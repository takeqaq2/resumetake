package services

import (
	"context"
	"crypto/sha256"
	"fmt"
	"log"
	"sync"
	"sync/atomic"
	"time"
)

type CacheEntry struct {
	Data      interface{}
	ExpiresAt time.Time
}

type ResponseCache struct {
	mu      sync.RWMutex
	entries map[string]*CacheEntry
	ttl     time.Duration
	maxSize int
	hits    atomic.Int64
	misses  atomic.Int64
}

var resumeCache = &ResponseCache{
	entries: make(map[string]*CacheEntry),
	ttl:     24 * time.Hour,
	maxSize: 1000,
}

func (c *ResponseCache) Get(key string) (interface{}, bool) {
	c.mu.RLock()
	entry, ok := c.entries[key]
	if ok && time.Now().Before(entry.ExpiresAt) {
		c.mu.RUnlock()
		c.hits.Add(1)
		return entry.Data, true
	}
	c.mu.RUnlock()

	if ok {
		// Entry exists but expired — remove under write lock. Re-check
		// identity under the write lock: between RUnlock and Lock another
		// goroutine may have called Set(key, newData) and replaced our
		// expired entry with a fresh one. Without this check, we would
		// delete the fresh entry and force a redundant AI call.
		c.mu.Lock()
		cur, stillThere := c.entries[key]
		if stillThere && cur == entry {
			delete(c.entries, key)
		}
		c.mu.Unlock()
	}
	c.misses.Add(1)
	return nil, false
}

func (c *ResponseCache) Set(key string, data interface{}) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Only evict when adding a new key — overwriting an existing key
	// doesn't increase the entry count, so evicting would needlessly
	// shrink the cache below maxSize.
	if _, exists := c.entries[key]; !exists && len(c.entries) >= c.maxSize {
		c.evictOldest()
	}

	c.entries[key] = &CacheEntry{
		Data:      data,
		ExpiresAt: time.Now().Add(c.ttl),
	}
}

func (c *ResponseCache) evictOldest() {
	var oldestKey string
	var oldestTime time.Time

	for key, entry := range c.entries {
		if oldestKey == "" || entry.ExpiresAt.Before(oldestTime) {
			oldestKey = key
			oldestTime = entry.ExpiresAt
		}
	}

	if oldestKey != "" {
		delete(c.entries, oldestKey)
		if len(oldestKey) > 8 {
			log.Printf("[Cache] Evicted oldest entry: %s", oldestKey[:8])
		} else {
			log.Printf("[Cache] Evicted oldest entry: %s", oldestKey)
		}
	}
}

func (c *ResponseCache) Cleanup() {
	c.mu.Lock()
	defer c.mu.Unlock()

	now := time.Now()
	for key, entry := range c.entries {
		if now.After(entry.ExpiresAt) {
			delete(c.entries, key)
		}
	}
}

func (c *ResponseCache) Stats() map[string]interface{} {
	c.mu.RLock()
	defer c.mu.RUnlock()

	hits := c.hits.Load()
	misses := c.misses.Load()
	total := hits + misses
	var hitRate float64
	if total > 0 {
		hitRate = float64(hits) / float64(total) * 100
	}

	return map[string]interface{}{
		"entries":     len(c.entries),
		"maxSize":     c.maxSize,
		"ttl":         c.ttl.String(),
		"hits":        hits,
		"misses":      misses,
		"hitRate":     fmt.Sprintf("%.1f%%", hitRate),
	}
}

// GenerateCacheKey produces a SHA256 cache key from all request parameters
// that affect the AI output. Including targetJob and moduleHints prevents
// cross-request result mismatches: without them, a user optimizing for
// "engineer" and another optimizing for "product manager" with the same
// resume/jobDesc/lang would hit the same cache entry and receive the wrong
// target-job-specific optimization.
// R57b-B3: cache key version. Bumped when system prompts or AI model
// change, so stale cached results from a prior deployment are never served.
// Without this, updating prompts.go or switching models leaves old results
// in cache for up to 24h (TTL), causing users to get optimization based on
// the previous prompt/model.
const cacheKeyVersion = "v2-r57b"

func GenerateCacheKey(resumeText, jobDesc, targetJob, moduleHints, lang string) string {
	h := sha256.New()
	h.Write([]byte(cacheKeyVersion))
	h.Write([]byte{0})
	for _, p := range []string{resumeText, jobDesc, targetJob, moduleHints, lang} {
		h.Write([]byte(p))
		h.Write([]byte{0}) // NUL delimiter prevents pipe-character collisions
	}
	return fmt.Sprintf("%x", h.Sum(nil))
}

func GetCachedResult(resumeText, jobDesc, targetJob, moduleHints, lang string) (interface{}, bool) {
	key := GenerateCacheKey(resumeText, jobDesc, targetJob, moduleHints, lang)
	return resumeCache.Get(key)
}

func SetCachedResult(resumeText, jobDesc, targetJob, moduleHints, lang string, result interface{}) {
	key := GenerateCacheKey(resumeText, jobDesc, targetJob, moduleHints, lang)
	resumeCache.Set(key, result)
}

func GetCacheStats() map[string]interface{} {
	return resumeCache.Stats()
}

// StartCleanup launches a background goroutine that purges expired cache
// entries every hour. It returns immediately. The goroutine exits when ctx
// is cancelled, allowing graceful shutdown (replaces the prior init() which
// could not be stopped).
func StartCleanup(ctx context.Context) {
	go func() {
		ticker := time.NewTicker(1 * time.Hour)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				resumeCache.Cleanup()
			}
		}
	}()
}
