package services

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"log"
	"sync"
	"time"
)

type DedupeEntry struct {
	Key       string
	Token     string
	ExpiresAt time.Time
}

type RequestDedup struct {
	mu      sync.RWMutex
	entries map[string]*DedupeEntry
	ttl     time.Duration
}

var dedup = &RequestDedup{
	entries: make(map[string]*DedupeEntry),
	ttl:     5 * time.Minute,
}

// TryAcquire attempts to acquire a dedup lock for the given key. Returns
// (true, token, nil) if the lock was acquired, (false, "", nil) if a live
// entry already exists, or (false, "", err) if the CSPRNG is unavailable.
// The token must be passed to Release to prove ownership — this prevents a
// stale Release (from a request whose entry expired and was re-acquired by
// another request) from deleting the new owner's entry.
// R54b-B5: fail-closed on crypto/rand failure — previously fell back to a
// predictable timestamp token, which let an attacker guess another request's
// token and maliciously release its dedup lock. Now returns an error so the
// caller can surface a 503 instead of proceeding with a predictable token.
func (d *RequestDedup) TryAcquire(key string) (bool, string, error) {
	d.mu.Lock()
	defer d.mu.Unlock()

	if entry, ok := d.entries[key]; ok {
		if time.Now().Before(entry.ExpiresAt) {
			return false, "", nil
		}
		delete(d.entries, key)
	}

	token, err := generateDedupToken()
	if err != nil {
		return false, "", err
	}
	d.entries[key] = &DedupeEntry{
		Key:       key,
		Token:     token,
		ExpiresAt: time.Now().Add(d.ttl),
	}
	return true, token, nil
}

// Release removes the dedup entry for the given key, but only if the token
// matches the entry's token. If the token doesn't match (e.g. the original
// entry expired and was re-acquired by a different request), the entry is
// left untouched — deleting it would break the new owner's dedup protection.
func (d *RequestDedup) Release(key, token string) {
	if token == "" {
		return
	}
	d.mu.Lock()
	defer d.mu.Unlock()
	if entry, ok := d.entries[key]; ok && entry.Token == token {
		delete(d.entries, key)
	}
}

// generateDedupToken produces a random 16-byte hex token. Using crypto/rand
// ensures tokens are unpredictable — an attacker can't guess another
// request's token to maliciously release its dedup lock.
// R54b-B5: fail-closed — returns an error instead of a predictable timestamp
// fallback. CSPRNG failure indicates a serious system issue (e.g. /dev/urandom
// exhaustion); proceeding with a predictable token would let an attacker
// release other requests' dedup locks.
func generateDedupToken() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		log.Printf("[ERROR] crypto/rand failed for dedup token: %v", err)
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func (d *RequestDedup) Cleanup() {
	d.mu.Lock()
	defer d.mu.Unlock()

	now := time.Now()
	for key, entry := range d.entries {
		if now.After(entry.ExpiresAt) {
			delete(d.entries, key)
		}
	}
}

// GenerateDedupeKey produces a stable SHA-256 hash of the concatenated parts.
// Previously this used raw string concatenation with a "|" separator, which
// could collide when a part itself contained "|" (e.g. resume text with pipe
// characters): GenerateDedupeKey("uid", "a|b", "c") and ("uid", "a", "b|c")
// both produced "uid|a|b|c|". Hashing eliminates structural ambiguity and
// keeps the key length fixed regardless of input size.
func GenerateDedupeKey(parts ...string) string {
	h := sha256.New()
	for _, p := range parts {
		h.Write([]byte(p))
		h.Write([]byte{0}) // unambiguous separator (NUL cannot appear in text)
	}
	return hex.EncodeToString(h.Sum(nil))
}

func TryAcquireRequest(key string) (bool, string, error) {
	return dedup.TryAcquire(key)
}

func ReleaseRequest(key, token string) {
	dedup.Release(key, token)
}

// StartDedupCleanup launches the background goroutine that purges expired
// dedup entries every minute. It replaces the init() version which had no
// exit mechanism. Call StopDedupCleanup() (or cancel the passed context) to
// stop the goroutine on graceful shutdown.
func StartDedupCleanup(ctx context.Context) {
	go func() {
		ticker := time.NewTicker(1 * time.Minute)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				dedup.Cleanup()
			case <-ctx.Done():
				return
			}
		}
	}()
}
