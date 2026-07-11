package services

import (
	"testing"
	"time"
)

func TestCacheMaxSize(t *testing.T) {
	cache := &ResponseCache{
		entries: make(map[string]*CacheEntry),
		ttl:     1 * time.Hour,
		maxSize: 3,
	}

	// Small sleeps ensure each entry gets a strictly-later ExpiresAt so the
	// evictOldest selection is deterministic. Without this, entries set in
	// nanoseconds of each other have near-identical ExpiresAt and Go's
	// random map iteration order makes eviction non-deterministic.
	cache.Set("key1", "value1")
	time.Sleep(time.Millisecond)
	cache.Set("key2", "value2")
	time.Sleep(time.Millisecond)
	cache.Set("key3", "value3")

	if len(cache.entries) != 3 {
		t.Errorf("Expected 3 entries, got %d", len(cache.entries))
	}

	cache.Set("key4", "value4")

	if len(cache.entries) != 3 {
		t.Errorf("Expected 3 entries after eviction, got %d", len(cache.entries))
	}

	_, ok := cache.Get("key1")
	if ok {
		t.Error("Expected key1 to be evicted")
	}
}

func TestCacheHitRate(t *testing.T) {
	cache := &ResponseCache{
		entries: make(map[string]*CacheEntry),
		ttl:     1 * time.Hour,
		maxSize: 100,
	}

	cache.Set("key1", "value1")
	cache.Get("key1")
	cache.Get("key1")
	cache.Get("nonexistent")

	stats := cache.Stats()
	if stats["hits"] != int64(2) {
		t.Errorf("Expected 2 hits, got %v", stats["hits"])
	}
	if stats["misses"] != int64(1) {
		t.Errorf("Expected 1 miss, got %v", stats["misses"])
	}
}

func TestCacheStats(t *testing.T) {
	cache := &ResponseCache{
		entries: make(map[string]*CacheEntry),
		ttl:     1 * time.Hour,
		maxSize: 500,
	}

	stats := cache.Stats()
	if stats["maxSize"] != 500 {
		t.Errorf("Expected maxSize 500, got %v", stats["maxSize"])
	}
	if stats["entries"] != 0 {
		t.Errorf("Expected 0 entries, got %v", stats["entries"])
	}
}
