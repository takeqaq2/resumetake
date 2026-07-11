package services

import (
	"testing"
	"time"
)

func TestDedupTryAcquire(t *testing.T) {
	d := &RequestDedup{
		entries: make(map[string]*DedupeEntry),
		ttl:     1 * time.Minute,
	}

	ok, token1, _ := d.TryAcquire("key1")
	if !ok {
		t.Error("Expected first acquire to succeed")
	}
	if token1 == "" {
		t.Error("Expected non-empty token")
	}

	if ok2, _, _ := d.TryAcquire("key1"); ok2 {
		t.Error("Expected second acquire to fail")
	}

	d.Release("key1", token1)

	if ok, _, _ := d.TryAcquire("key1"); !ok {
		t.Error("Expected acquire after release to succeed")
	}
}

func TestDedupExpiry(t *testing.T) {
	d := &RequestDedup{
		entries: make(map[string]*DedupeEntry),
		ttl:     10 * time.Millisecond,
	}

	d.TryAcquire("key1")

	time.Sleep(20 * time.Millisecond)

	if ok, _, _ := d.TryAcquire("key1"); !ok {
		t.Error("Expected acquire after expiry to succeed")
	}
}

func TestDedupStaleRelease(t *testing.T) {
	d := &RequestDedup{
		entries: make(map[string]*DedupeEntry),
		ttl:     10 * time.Millisecond,
	}

	// Request A acquires the key.
	_, tokenA, _ := d.TryAcquire("key1")

	// Entry expires; request B re-acquires the same key with a new token.
	time.Sleep(20 * time.Millisecond)
	okB, tokenB, _ := d.TryAcquire("key1")
	if !okB {
		t.Fatal("Expected re-acquire after expiry to succeed")
	}
	if tokenB == tokenA {
		t.Error("Expected different tokens for different acquisitions")
	}

	// Request A's stale Release must NOT delete request B's entry.
	d.Release("key1", tokenA)

	// Request B's entry should still exist — the stale release was rejected.
	if ok, _, _ := d.TryAcquire("key1"); ok {
		t.Error("Stale release deleted new owner's entry — dedup protection broken")
	}

	// Request B's Release with the correct token should work.
	d.Release("key1", tokenB)
	if ok, _, _ := d.TryAcquire("key1"); !ok {
		t.Error("Expected acquire after correct release to succeed")
	}
}

func TestDedupCleanup(t *testing.T) {
	d := &RequestDedup{
		entries: make(map[string]*DedupeEntry),
		ttl:     10 * time.Millisecond,
	}

	d.TryAcquire("key1")
	d.TryAcquire("key2")

	time.Sleep(20 * time.Millisecond)

	d.Cleanup()

	d.mu.RLock()
	count := len(d.entries)
	d.mu.RUnlock()

	if count != 0 {
		t.Errorf("Expected 0 entries after cleanup, got %d", count)
	}
}
