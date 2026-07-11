package store

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"resumetake/models"
)

type UserPersistence struct {
	mu       sync.Mutex
	filePath string
}

func NewUserPersistence(filePath string) *UserPersistence {
	return &UserPersistence{filePath: filePath}
}

func (up *UserPersistence) Load() (map[string]*models.User, error) {
	// R53b-B1: callers that need atomic read-modify-write should use
	// loadLocked/saveLocked under up.mu. This public Load does not hold
	// the lock (for read-only callers like migration).
	return up.loadLocked()
}

// loadLocked reads the JSON file without holding up.mu. Callers must
// already hold up.mu if doing read-modify-write.
func (up *UserPersistence) loadLocked() (map[string]*models.User, error) {
	if up.filePath == "" {
		return make(map[string]*models.User), nil
	}

	b, err := os.ReadFile(up.filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return make(map[string]*models.User), nil
		}
		return nil, err
	}

	loaded := make(map[string]*models.PersistedUser)
	if err := json.Unmarshal(b, &loaded); err != nil {
		return nil, err
	}

	users := make(map[string]*models.User, len(loaded))
	for email, user := range loaded {
		users[email] = &models.User{
			ID:               user.ID,
			Email:            user.Email,
			Password:         user.Password,
			PasswordType:     user.PasswordType,
			Name:             user.Name,
			Token:            user.Token,
			UsageCount:       user.UsageCount,
			MaxFreeUsage:     user.MaxFreeUsage,
			Plan:             user.Plan,
			SubscriptionID:   user.SubscriptionID,
			CaptureID:        user.CaptureID,
			PurchasedTemplates: user.PurchasedTemplates,
			CreatedAt:        user.CreatedAt,
		}
	}

	return users, nil
}

func (up *UserPersistence) Save(users map[string]*models.User) error {
	up.mu.Lock()
	defer up.mu.Unlock()
	return up.saveLocked(users)
}

// saveLocked writes the users map to disk without holding up.mu. Callers must
// already hold up.mu (R53b-B1: enables atomic Load→modify→Save sequences).
func (up *UserPersistence) saveLocked(users map[string]*models.User) error {
	if up.filePath == "" {
		return nil
	}

	snapshot := make(map[string]*models.PersistedUser, len(users))
	for email, user := range users {
		snapshot[email] = &models.PersistedUser{
			ID:               user.ID,
			Email:            user.Email,
			Password:         user.Password,
			PasswordType:     user.PasswordType,
			Name:             user.Name,
			Token:            user.Token,
			UsageCount:       user.UsageCount,
			MaxFreeUsage:     user.MaxFreeUsage,
			Plan:             user.Plan,
			SubscriptionID:   user.SubscriptionID,
			CaptureID:        user.CaptureID,
			PurchasedTemplates: user.PurchasedTemplates,
			CreatedAt:        user.CreatedAt,
		}
	}

	if err := os.MkdirAll(filepath.Dir(up.filePath), 0755); err != nil {
		return err
	}

	b, err := json.MarshalIndent(snapshot, "", "  ")
	if err != nil {
		return err
	}

	tmp := up.filePath + ".tmp"
	if err := os.WriteFile(tmp, b, 0600); err != nil {
		return err
	}

	if err := os.Rename(tmp, up.filePath); err != nil {
		return fmt.Errorf("failed to rename temp file: %w", err)
	}

	return nil
}

// cloneUser returns a deep copy of the User so the store owns its own struct
// value rather than the caller's live *User pointer. Without this, a caller
// could mutate the User after SaveUser returns and corrupt the in-memory state
// observed by the next reader (R53b-B2).
func cloneUser(u *models.User) *models.User {
	if u == nil {
		return nil
	}
	return &models.User{
		ID:               u.ID,
		Email:            u.Email,
		Password:         u.Password,
		PasswordType:     u.PasswordType,
		Name:             u.Name,
		Token:            u.Token,
		UsageCount:       u.UsageCount,
		MaxFreeUsage:     u.MaxFreeUsage,
		Plan:             u.Plan,
		SubscriptionID:   u.SubscriptionID,
		CaptureID:        u.CaptureID,
		PurchasedTemplates: u.PurchasedTemplates,
		CreatedAt:        u.CreatedAt,
	}
}

// SaveUser persists a single user. The JSON file format does not support
// partial updates, so this rewrites the whole file. In production the SQLite
// backed DatabasePersistence.SaveUser performs a true single-row write; this
// implementation exists to satisfy the interface for tests/migration only.
// R53b-B1: atomic under up.mu (previously Load→modify→Save raced).
// R53b-B2: clones the caller's *User so later caller-side mutation cannot
// corrupt the stored snapshot.
func (up *UserPersistence) SaveUser(user *models.User) error {
	if user == nil {
		return nil
	}
	up.mu.Lock()
	defer up.mu.Unlock()
	loaded, err := up.loadLocked()
	if err != nil {
		return err
	}
	loaded[user.Email] = cloneUser(user)
	return up.saveLocked(loaded)
}

// UpdateUserUsage updates only usage_count for the given email. The JSON file
// format does not support partial updates, so this loads the full file, patches
// the field, and writes back. In production the SQLite backed
// DatabasePersistence.UpdateUserUsage performs a true targeted UPDATE; this
// implementation exists to satisfy the interface for tests only.
// R37b-B2: max_free_usage removed — see database/sqlite.go for rationale.
// R53b-B1: atomic under up.mu (previously Load→modify→Save raced).
func (up *UserPersistence) UpdateUserUsage(email string, usageCount int) error {
	up.mu.Lock()
	defer up.mu.Unlock()
	loaded, err := up.loadLocked()
	if err != nil {
		return err
	}
	u, ok := loaded[email]
	if !ok {
		return nil
	}
	u.UsageCount = usageCount
	return up.saveLocked(loaded)
}

// AdjustUserUsage performs an incremental update of usage_count.
// R52-B1: JSON file implementation for tests — loads, patches, saves.
// R53b-B1: atomic under up.mu (previously Load→modify→Save raced).
func (up *UserPersistence) AdjustUserUsage(email string, delta int) error {
	up.mu.Lock()
	defer up.mu.Unlock()
	loaded, err := up.loadLocked()
	if err != nil {
		return err
	}
	u, ok := loaded[email]
	if !ok {
		return nil
	}
	u.UsageCount += delta
	return up.saveLocked(loaded)
}

// UpdateUserToken updates only the token for the given email. JSON file
// implementation (test/migration only); SQLite backed DatabasePersistence
// performs a true targeted UPDATE. R37-B1.
// R53b-B1: atomic under up.mu (previously Load→modify→Save raced).
func (up *UserPersistence) UpdateUserToken(email, token string) error {
	up.mu.Lock()
	defer up.mu.Unlock()
	loaded, err := up.loadLocked()
	if err != nil {
		return err
	}
	u, ok := loaded[email]
	if !ok {
		return nil
	}
	u.Token = token
	return up.saveLocked(loaded)
}

// UpdateUserTokenAndPassword updates token plus (optionally) password and
// password_type. R37-B1. When password is empty, only token is updated.
// R53b-B1: atomic under up.mu (previously Load→modify→Save raced).
func (up *UserPersistence) UpdateUserTokenAndPassword(email, token, password, passwordType string) error {
	up.mu.Lock()
	defer up.mu.Unlock()
	loaded, err := up.loadLocked()
	if err != nil {
		return err
	}
	u, ok := loaded[email]
	if !ok {
		return nil
	}
	u.Token = token
	if password != "" {
		u.Password = password
		u.PasswordType = passwordType
	}
	return up.saveLocked(loaded)
}

// UpdateUserPlan updates plan, subscription_id, capture_id, and max_free_usage
// for the given email. R37-B1.
// R53b-B1: atomic under up.mu (previously Load→modify→Save raced).
func (up *UserPersistence) UpdateUserPlan(email, plan, subscriptionID, captureID string, maxFreeUsage int) error {
	up.mu.Lock()
	defer up.mu.Unlock()
	loaded, err := up.loadLocked()
	if err != nil {
		return err
	}
	u, ok := loaded[email]
	if !ok {
		return nil
	}
	u.Plan = plan
	u.SubscriptionID = subscriptionID
	u.CaptureID = captureID
	u.MaxFreeUsage = maxFreeUsage
	return up.saveLocked(loaded)
}

// UpdateUserTemplates updates only the purchased_templates list for the given
// email. JSON file implementation (test/migration only); SQLite backed
// DatabasePersistence performs a true targeted UPDATE. R37-B1.
// R53b-B1: atomic under up.mu (previously Load→modify→Save raced).
func (up *UserPersistence) UpdateUserTemplates(email string, templates []string) error {
	up.mu.Lock()
	defer up.mu.Unlock()
	loaded, err := up.loadLocked()
	if err != nil {
		return err
	}
	u, ok := loaded[email]
	if !ok {
		return nil
	}
	u.PurchasedTemplates = templates
	return up.saveLocked(loaded)
}
