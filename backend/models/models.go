package models

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/hex"
	"fmt"
	"sync"
	"time"
)

const MaxResumes = 5000

type User struct {
	ID             string    `json:"id"`
	Email          string    `json:"email"`
	Password       string    `json:"-"`
	PasswordType   string    `json:"-"`
	Name           string    `json:"name"`
	Token          string    `json:"token"`
	UsageCount     int       `json:"usage_count"`
	MaxFreeUsage   int       `json:"max_free_usage"`
	Plan           string    `json:"plan"`
	SubscriptionID string    `json:"subscription_id,omitempty"`
	CaptureID      string    `json:"capture_id,omitempty"`
	CreatedAt      time.Time `json:"created_at"`
}

type PersistedUser struct {
	ID             string    `json:"id"`
	Email          string    `json:"email"`
	Password       string    `json:"password"`
	PasswordType   string    `json:"password_type,omitempty"`
	Name           string    `json:"name"`
	Token          string    `json:"token"`
	UsageCount     int       `json:"usage_count"`
	MaxFreeUsage   int       `json:"max_free_usage"`
	Plan           string    `json:"plan"`
	SubscriptionID string    `json:"subscription_id,omitempty"`
	CaptureID      string    `json:"capture_id,omitempty"`
	CreatedAt      time.Time `json:"created_at"`
}

type Resume struct {
	ID        string                 `json:"id"`
	OwnerID   string                 `json:"owner_id,omitempty"`
	Title     string                 `json:"title"`
	Content   map[string]interface{} `json:"content"`
	CreatedAt string                 `json:"created_at"`
	UpdatedAt string                 `json:"updated_at"`
}

type VerificationCode struct {
	Code      string
	CreatedAt time.Time
	Attempts  int
}

type AIProvider struct {
	Name    string
	BaseURL string
	Model   string
	APIKey  string
}

type GroqRequest struct {
	Model       string        `json:"model"`
	Messages    []GroqMessage `json:"messages"`
	MaxTokens   int           `json:"max_tokens,omitempty"`
	Temperature float64       `json:"temperature,omitempty"`
}

type GroqMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type GroqResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
	// R46-B2: All OpenAI-compatible providers return usage.prompt_tokens /
	// usage.completion_tokens. Without parsing them, logAICall always logs
	// InputTokens=0 OutputTokens=0, making cost auditing impossible.
	Usage *struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
	} `json:"usage,omitempty"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

// GenerateToken returns a cryptographically secure random token. If the
// CSPRNG fails, it returns an error (fail-closed) rather than falling back
// to a predictable hash — token security must not be weaker than the
// verification code path (see GenerateVerificationCode in email.go).
func GenerateToken(email string) (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("failed to generate secure token: %w", err)
	}
	return hex.EncodeToString(b), nil
}

type Store struct {
	mu        sync.RWMutex
	resumes   map[string]*Resume
	insertSeq map[string]int64
	seq       int64
}

func NewStore() *Store {
	return &Store{resumes: make(map[string]*Resume), insertSeq: make(map[string]int64)}
}

// cloneContent returns a shallow copy of the Content map's top-level keys.
// R56-B2: without this, Save/Get/GetAll return shallow struct clones that
// share the same Content map pointer — a caller mutating the returned
// Content in-place causes a concurrent map write panic (fatal, unrecoverable)
// if another goroutine reads the store's copy simultaneously.
func cloneContent(src map[string]interface{}) map[string]interface{} {
	if src == nil {
		return nil
	}
	dst := make(map[string]interface{}, len(src))
	for k, v := range src {
		dst[k] = v
	}
	return dst
}

func (s *Store) Save(resume *Resume) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, exists := s.resumes[resume.ID]; !exists && len(s.resumes) >= MaxResumes {
		oldest := ""
		var oldestSeq int64 = -1
		for k, seq := range s.insertSeq {
			if oldestSeq == -1 || seq < oldestSeq {
				oldestSeq = seq
				oldest = k
			}
		}
		if oldest != "" {
			delete(s.resumes, oldest)
			delete(s.insertSeq, oldest)
		}
	}
	// R34-L7: update insertSeq on every Save (not just first insert) so
	// frequently-updated resumes are not evicted — makes the cache true
	// LRU rather than FIFO.
	// R49-B4: store a shallow clone (like UserStore.Save does) so the
	// caller's pointer is decoupled from the store's internal state.
	// R56-B2: deep copy the Content map to prevent concurrent map write
	// panics if a caller mutates the returned Content in-place.
	s.seq++
	s.insertSeq[resume.ID] = s.seq
	cp := *resume
	cp.Content = cloneContent(resume.Content)
	s.resumes[resume.ID] = &cp
}

func (s *Store) Get(id string) (*Resume, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	r, ok := s.resumes[id]
	if !ok {
		return nil, false
	}
	// R34-L7: touch access order so frequently-read resumes survive
	// eviction — true LRU semantics.
	// R49-B4: return a shallow clone (consistent with GetAll) so callers
	// can't mutate the store's internal entry by accident. Matches the
	// pattern in UserStore.GetByEmail/GetByToken which all return clones.
	// R56-B2: deep copy Content map — see cloneContent comment.
	s.seq++
	s.insertSeq[id] = s.seq
	cp := *r
	cp.Content = cloneContent(r.Content)
	return &cp, true
}

func (s *Store) Delete(id string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.resumes[id]; ok {
		delete(s.resumes, id)
		delete(s.insertSeq, id)
		return true
	}
	return false
}

func (s *Store) Count() int64 {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return int64(len(s.resumes))
}

// GetAll returns a slice of all resumes currently in the in-memory store.
// Returns shallow clones (not pointers to the live map entries) so callers
// can iterate safely without holding the lock. Used by the List handler's
// in-memory fallback.
func (s *Store) GetAll() []*Resume {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make([]*Resume, 0, len(s.resumes))
	for _, r := range s.resumes {
		cp := *r // shallow copy — prevents callers from mutating live entries
		cp.Content = cloneContent(r.Content) // R56-B2
		result = append(result, &cp)
	}
	return result
}

type UserStore struct {
	mu               sync.RWMutex
	users            map[string]*User
	byToken          map[string]*User
	bySubscriptionID map[string]*User
	byCaptureID      map[string]*User
}

func NewUserStore() *UserStore {
	return &UserStore{
		users:            make(map[string]*User),
		byToken:          make(map[string]*User),
		bySubscriptionID: make(map[string]*User),
		byCaptureID:      make(map[string]*User),
	}
}

// cloneUser returns a deep copy of the user struct. The UserStore stores
// *User pointers in its internal maps and mutates them in place under the
// write lock (via UpdateUser). Returning the live pointer to callers would
// expose them to data races: a handler reading user.UsageCount on the live
// pointer while another goroutine's UpdateUser mutates it concurrently.
// All getters therefore return clones so callers operate on a stable snapshot.
func cloneUser(u *User) *User {
	if u == nil {
		return nil
	}
	cp := *u
	return &cp
}

func (us *UserStore) GetByEmail(email string) (*User, bool) {
	us.mu.RLock()
	defer us.mu.RUnlock()
	u, ok := us.users[email]
	return cloneUser(u), ok
}

func (us *UserStore) GetByToken(token string) (*User, bool) {
	if token == "" {
		return nil, false
	}
	us.mu.RLock()
	defer us.mu.RUnlock()
	u, ok := us.byToken[token]
	if !ok {
		return nil, false
	}
	// R56-B3: removed ConstantTimeCompare — it was dead code. The map lookup
	// already matched token to u.Token (map key == u.Token), and the RLock
	// prevents concurrent mutation, so u.Token == token is guaranteed. The
	// constant-time comparison provided no real defense: the timing difference
	// between map hit vs. miss already occurred before this line. Token
	// security relies on 256-bit entropy (brute-force infeasible) + the
	// AuthMiddleware rate limiter.
	return cloneUser(u), true
}

func (us *UserStore) Count() int {
	us.mu.RLock()
	defer us.mu.RUnlock()
	return len(us.users)
}

func (us *UserStore) Save(user *User) {
	us.mu.Lock()
	defer us.mu.Unlock()
	if old, exists := us.users[user.Email]; exists {
		if old.Token != "" {
			delete(us.byToken, old.Token)
		}
		if old.SubscriptionID != "" {
			delete(us.bySubscriptionID, old.SubscriptionID)
		}
		if old.CaptureID != "" {
			delete(us.byCaptureID, old.CaptureID)
		}
	}
	// Store a clone so the caller's pointer is decoupled from the store's
	// internal state — subsequent mutations by the caller won't corrupt the
	// store, and the store's in-place mutations (UpdateUser) won't race with
	// callers still holding the original pointer.
	stored := cloneUser(user)
	us.users[user.Email] = stored
	if stored.Token != "" {
		us.byToken[stored.Token] = stored
	}
	if stored.SubscriptionID != "" {
		us.bySubscriptionID[stored.SubscriptionID] = stored
	}
	if stored.CaptureID != "" {
		us.byCaptureID[stored.CaptureID] = stored
	}
}

// GetByCaptureID returns the user associated with the given PayPal capture ID
// in O(1). Used by the PayPal REFUND webhook handler — for refund events,
// parent_id is the capture ID (not the order ID stored in SubscriptionID).
func (us *UserStore) GetByCaptureID(captureID string) (*User, bool) {
	if captureID == "" {
		return nil, false
	}
	us.mu.RLock()
	defer us.mu.RUnlock()
	u, ok := us.byCaptureID[captureID]
	return cloneUser(u), ok
}

// GetBySubscriptionID returns the user associated with the given PayPal
// subscription/order ID in O(1). Used by the PayPal webhook handler to avoid
// an O(n) scan of all users on every PAYMENT.CAPTURE.COMPLETED event.
func (us *UserStore) GetBySubscriptionID(subID string) (*User, bool) {
	if subID == "" {
		return nil, false
	}
	us.mu.RLock()
	defer us.mu.RUnlock()
	u, ok := us.bySubscriptionID[subID]
	return cloneUser(u), ok
}

func (us *UserStore) GetAll() map[string]*User {
	us.mu.RLock()
	defer us.mu.RUnlock()
	users := make(map[string]*User, len(us.users))
	for k, v := range us.users {
		users[k] = cloneUser(v)
	}
	return users
}

// UpdateUser atomically mutates the user identified by email under the write
// lock, then returns a clone of the updated user. Use this instead of Get ->
// mutate -> Save to avoid lost updates when concurrent requests touch the same
// user (e.g. UsageCount increments or Plan upgrades). If the user does not
// exist, the mutate callback is not invoked and (nil, false) is returned.
func (us *UserStore) UpdateUser(email string, mutate func(*User)) (*User, bool) {
	us.mu.Lock()
	defer us.mu.Unlock()
	u, ok := us.users[email]
	if !ok {
		return nil, false
	}
	oldToken := u.Token
	oldSubID := u.SubscriptionID
	oldCapID := u.CaptureID
	mutate(u)
	// Reindex Token, SubscriptionID, and CaptureID if the mutate callback changed them.
	if u.Token != oldToken {
		if oldToken != "" {
			delete(us.byToken, oldToken)
		}
		if u.Token != "" {
			us.byToken[u.Token] = u
		}
	}
	if u.SubscriptionID != oldSubID {
		if oldSubID != "" {
			delete(us.bySubscriptionID, oldSubID)
		}
		if u.SubscriptionID != "" {
			us.bySubscriptionID[u.SubscriptionID] = u
		}
	}
	if u.CaptureID != oldCapID {
		if oldCapID != "" {
			delete(us.byCaptureID, oldCapID)
		}
		if u.CaptureID != "" {
			us.byCaptureID[u.CaptureID] = u
		}
	}
	return cloneUser(u), true
}

// IncrementUsage atomically increments the user's UsageCount and returns the
// new value. Safe for concurrent requests on the same user.
func (us *UserStore) IncrementUsage(email string) (int, bool) {
	u, ok := us.UpdateUser(email, func(u *User) { u.UsageCount++ })
	if !ok {
		return 0, false
	}
	return u.UsageCount, true
}

// CheckAndIncrementUsage atomically checks the free-user quota and increments
// UsageCount if allowed. This prevents the TOCTOU race where concurrent
// requests each pass the `UsageCount >= MaxFreeUsage` check before any of
// them increments. For paid users (plan != "free"), always increments.
// Returns (newUsage, true) if allowed; (currentUsage, false) if quota
// exceeded or user not found.
func (us *UserStore) CheckAndIncrementUsage(email string) (int, bool) {
	us.mu.Lock()
	defer us.mu.Unlock()
	u, ok := us.users[email]
	if !ok {
		return 0, false
	}
	if u.Plan == "free" && u.UsageCount >= u.MaxFreeUsage {
		return u.UsageCount, false
	}
	u.UsageCount++
	return u.UsageCount, true
}

// DecrementUsage reverses a previous CheckAndIncrementUsage when the AI call
// fails, so users are not charged for failed attempts. Clamped at 0.
func (us *UserStore) DecrementUsage(email string) {
	us.UpdateUser(email, func(u *User) {
		if u.UsageCount > 0 {
			u.UsageCount--
		}
	})
}

// SaveIfAbsent atomically inserts a new user only if no user with the same
// email already exists. Returns true if the user was created, false if a user
// with that email already exists (in which case the store is left unchanged).
// Use this instead of GetByEmail -> Save in Register to eliminate the
// check-then-act race where two concurrent registrations for the same email
// both pass the "not exists" check and the second Save overwrites the first
// (account takeover via password overwrite).
func (us *UserStore) SaveIfAbsent(user *User) bool {
	us.mu.Lock()
	defer us.mu.Unlock()
	if _, exists := us.users[user.Email]; exists {
		return false
	}
	stored := cloneUser(user)
	us.users[user.Email] = stored
	if stored.Token != "" {
		us.byToken[stored.Token] = stored
	}
	if stored.SubscriptionID != "" {
		us.bySubscriptionID[stored.SubscriptionID] = stored
	}
	if stored.CaptureID != "" {
		us.byCaptureID[stored.CaptureID] = stored
	}
	return true
}

func (us *UserStore) Load(users map[string]*User) {
	us.mu.Lock()
	defer us.mu.Unlock()
	us.users = make(map[string]*User, len(users))
	us.byToken = make(map[string]*User, len(users))
	us.bySubscriptionID = make(map[string]*User, len(users))
	us.byCaptureID = make(map[string]*User, len(users))
	for _, u := range users {
		stored := cloneUser(u)
		us.users[stored.Email] = stored
		if stored.Token != "" {
			us.byToken[stored.Token] = stored
		}
		if stored.SubscriptionID != "" {
			us.bySubscriptionID[stored.SubscriptionID] = stored
		}
		if stored.CaptureID != "" {
			us.byCaptureID[stored.CaptureID] = stored
		}
	}
}

// SeenEventsStore provides in-memory idempotency for webhook events. PayPal
// delivers webhooks at-least-once and supports retries; without dedup, a
// replayed or duplicated signed event would be reprocessed (plan upgrade
// writes, notifications, etc.). Entries expire after 96h — PayPal's max retry
// window is 72h, but a small buffer prevents an edge race where a retry
// arrives just as the 72h TTL expires (entry purged → reprocessed).
type SeenEventsStore struct {
	mu        sync.RWMutex
	events    map[string]time.Time
	done      chan struct{}
	closeOnce sync.Once
}

const seenEventTTL = 96 * time.Hour

func NewSeenEventsStore() *SeenEventsStore {
	s := &SeenEventsStore{events: make(map[string]time.Time), done: make(chan struct{})}
	go func() {
		ticker := time.NewTicker(1 * time.Hour)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				s.purgeExpired()
			case <-s.done:
				return
			}
		}
	}()
	return s
}

func (s *SeenEventsStore) Close() {
	// R49-B6: sync.Once makes Close idempotent — prevents "close of closed
	// channel" panic if Close is called twice (e.g. graceful shutdown + defer).
	s.closeOnce.Do(func() { close(s.done) })
}

func (s *SeenEventsStore) purgeExpired() {
	s.mu.Lock()
	defer s.mu.Unlock()
	now := time.Now()
	for id, ts := range s.events {
		if now.Sub(ts) > seenEventTTL {
			delete(s.events, id)
		}
	}
}

// MarkSeen returns true if the event ID was newly recorded (first sighting),
// false if the event ID has already been seen within the TTL window.
// R44-B1: Empty eventID returns false (treat as "already seen") to prevent
// dedup bypass — returning true for empty IDs would allow repeated processing
// of events with missing IDs.
func (s *SeenEventsStore) MarkSeen(eventID string) bool {
	if eventID == "" {
		return false
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, exists := s.events[eventID]; exists {
		return false
	}
	s.events[eventID] = time.Now()
	return true
}

// UnmarkSeen removes an event ID from the seen set. Used when a webhook
// handler reserved the event ID via MarkSeen but processing failed (returned
// 500) — without unmarking, PayPal's retry would be silently deduplicated
// and the user would never get their plan upgrade.
func (s *SeenEventsStore) UnmarkSeen(eventID string) {
	if eventID == "" {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.events, eventID)
}

type VerificationStore struct {
	mu        sync.RWMutex
	codes     map[string]*VerificationCode
	cooldowns map[string]time.Time // R27-H1: cooldown for registered emails (no code stored)
	done      chan struct{}
	closeOnce sync.Once
}

func NewVerificationStore() *VerificationStore {
	vs := &VerificationStore{codes: make(map[string]*VerificationCode), cooldowns: make(map[string]time.Time), done: make(chan struct{})}
	// Start a background goroutine to purge expired codes every 5 minutes.
	// Without this, codes that are requested but never verified accumulate
	// indefinitely (memory leak exploitable via mass send-code requests).
	// The done channel is closed in Close() for graceful shutdown.
	go func() {
		ticker := time.NewTicker(5 * time.Minute)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				vs.purgeExpired()
			case <-vs.done:
				return
			}
		}
	}()
	return vs
}

func (vs *VerificationStore) Close() {
	// R49-B6: sync.Once makes Close idempotent — prevents "close of closed
	// channel" panic on double-close.
	vs.closeOnce.Do(func() { close(vs.done) })
}

func (vs *VerificationStore) purgeExpired() {
	vs.mu.Lock()
	defer vs.mu.Unlock()
	for email, vc := range vs.codes {
		if time.Since(vc.CreatedAt) > 5*time.Minute {
			delete(vs.codes, email)
		}
	}
	// R27-H1: also purge expired cooldown-only entries.
	for email, ts := range vs.cooldowns {
		if time.Since(ts) > 5*time.Minute {
			delete(vs.cooldowns, email)
		}
	}
}

// resendCooldown gates how often a user can request a verification code.
const resendCooldown = 60 * time.Second

// AcquireResendSlot atomically checks the resend cooldown and reserves a
// slot. R52b-B1: the old CanResend (RLock) + Save/SetCooldown (Lock) were
// separate critical sections — concurrent SendCode requests for the same
// email could all pass CanResend before any sets the cooldown, bypassing the
// 60s per-email rate limit and sending N emails to the victim. This method
// performs the check-and-reserve under a single Lock, closing the TOCTOU
// window. Returns true if the slot was acquired (caller may proceed),
// false if on cooldown.
// R56-B4: CanResend and SetCooldown were removed — they are superseded by
// this method and had no callers. Keeping them exported risked future
// developers reintroducing the TOCTOU race.
func (vs *VerificationStore) AcquireResendSlot(email string) bool {
	vs.mu.Lock()
	defer vs.mu.Unlock()
	if vc, ok := vs.codes[email]; ok && time.Since(vc.CreatedAt) < resendCooldown {
		return false
	}
	if ts, ok := vs.cooldowns[email]; ok && time.Since(ts) < resendCooldown {
		return false
	}
	// Reserve the slot immediately so concurrent requests see the cooldown.
	// The actual code is saved later via Save(); the cooldown timestamp is
	// what gates subsequent calls.
	vs.cooldowns[email] = time.Now()
	return true
}

// ReleaseResendSlot removes the cooldown entry set by AcquireResendSlot.
// R54-B1: used when SMTP send fails — without this, a transient SMTP
// failure locks the user out for 60s (the cooldown was set before the
// send attempt). Releasing allows immediate retry once SMTP recovers.
func (vs *VerificationStore) ReleaseResendSlot(email string) {
	vs.mu.Lock()
	defer vs.mu.Unlock()
	delete(vs.cooldowns, email)
}

func (vs *VerificationStore) Save(email, code string) {
	vs.mu.Lock()
	defer vs.mu.Unlock()
	vs.codes[email] = &VerificationCode{Code: code, CreatedAt: time.Now()}
}

func (vs *VerificationStore) Verify(email, code string) bool {
	vs.mu.Lock()
	defer vs.mu.Unlock()
	vc, ok := vs.codes[email]
	if !ok {
		return false
	}
	if time.Since(vc.CreatedAt) > 5*time.Minute {
		delete(vs.codes, email)
		return false
	}
	if subtle.ConstantTimeCompare([]byte(vc.Code), []byte(code)) != 1 {
		vc.Attempts++
		if vc.Attempts >= 5 {
			delete(vs.codes, email)
		}
		return false
	}
	delete(vs.codes, email)
	return true
}

var AllowedUploadExt = map[string]bool{
	".txt": true,
	".md":  true,
	".pdf": true,
}

const (
	MaxUploadBytes  = 1 * 1024 * 1024
	MaxPDFBytes     = 2 * 1024 * 1024
	MaxPDFPages     = 3
	MaxPDFTextChars = 15000
	MaxResumeChars  = 10000
)
