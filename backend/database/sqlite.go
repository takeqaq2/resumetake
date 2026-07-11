package database

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	_ "modernc.org/sqlite"

	"resumetake/models"
)

type Database struct {
	db        *sql.DB
	mu        sync.RWMutex
	path      string
	done      chan struct{}
	closeOnce sync.Once
}

func NewDatabase(path string) (*Database, error) {
	db, err := sql.Open("sqlite", path+"?_journal_mode=WAL&_busy_timeout=5000")
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// modernc.org/sqlite allows only one writer at a time in WAL mode.
	// Capping MaxOpenConns at 1 serializes writes (avoiding "database is
	// locked") while reads can still proceed via the shared cache. The
	// 5-minute conn lifetime keeps idle connections from going stale.
	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)
	db.SetConnMaxLifetime(5 * time.Minute)

	if err := db.Ping(); err != nil {
		db.Close() // R51b-B3: prevent connection leak on error path
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	d := &Database{db: db, path: path, done: make(chan struct{})}
	if err := d.migrate(); err != nil {
		db.Close() // R51b-B3: prevent connection leak on error path
		return nil, fmt.Errorf("failed to migrate database: %w", err)
	}

	// WAL checkpoint: without periodic checkpointing, the -wal file grows
	// unbounded under write traffic (user registrations, resume saves,
	// usage updates), eventually exhausting disk and degrading reads.
	// PASSIVE mode lets writers proceed concurrently and only merges what
	// is safe to merge. Running every 5 minutes keeps the -wal file small.
	// The done channel is closed in Close() so this goroutine exits
	// gracefully on shutdown instead of running on a closed DB.
	go func() {
		ticker := time.NewTicker(5 * time.Minute)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				// R56b-B2: hold RLock so this cannot race with Close().
				// Close() acquires the write lock and then calls
				// db.Close(); without coordination here, the checkpoint
				// could run on an already-closed *sql.DB and panic with
				// "sql: database is closed". RLock lets concurrent reads
				// (GetUser/GetAllUsers, which also take RLock) proceed
				// while serializing against Close()'s write lock.
				d.mu.RLock()
				select {
				case <-d.done:
					d.mu.RUnlock()
					return
				default:
				}
				_, err := db.Exec("PRAGMA wal_checkpoint(PASSIVE)")
				d.mu.RUnlock()
				if err != nil {
					log.Printf("[DB] wal_checkpoint failed: %v", err)
				}
			case <-d.done:
				return
			}
		}
	}()

	log.Printf("[DB] SQLite database initialized at %s", path)
	return d, nil
}

func (d *Database) migrate() error {
	// Wrap all migration DDL in a transaction so partial failures don't
	// leave the schema in an inconsistent state. All statements are
	// idempotent (IF NOT EXISTS + column existence checks), so a retry
	// after a crash is safe.
	tx, err := d.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin migration transaction: %w", err)
	}
	defer tx.Rollback()

	queries := []string{
		`CREATE TABLE IF NOT EXISTS users (
			id TEXT PRIMARY KEY,
			email TEXT UNIQUE NOT NULL,
			password TEXT NOT NULL,
			password_type TEXT DEFAULT 'bcrypt',
			name TEXT,
			token TEXT,
			usage_count INTEGER DEFAULT 0,
			max_free_usage INTEGER DEFAULT 5,
			plan TEXT DEFAULT 'free',
			subscription_id TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS resumes (
			id TEXT PRIMARY KEY,
			owner_id TEXT DEFAULT '',
			title TEXT NOT NULL,
			content TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		// Note: no separate idx_users_email — the UNIQUE constraint on the
		// email column already creates an implicit index. A redundant index
		// wastes storage and write throughput.
		`CREATE INDEX IF NOT EXISTS idx_users_token ON users(token)`,
	}

	for _, q := range queries {
		if _, err := tx.Exec(q); err != nil {
			return fmt.Errorf("migration failed: %w", err)
		}
	}

	// Add owner_id column to resumes if it doesn't exist (backward-compatible migration).
	// This MUST run before creating idx_resumes_owner, otherwise index creation fails on
	// existing databases where the resumes table predates the owner_id column.
	var colCount int
	err = tx.QueryRow(`SELECT COUNT(*) FROM pragma_table_info('resumes') WHERE name='owner_id'`).Scan(&colCount)
	if err != nil {
		return fmt.Errorf("failed to inspect resumes schema: %w", err)
	}
	if colCount == 0 {
		if _, err := tx.Exec(`ALTER TABLE resumes ADD COLUMN owner_id TEXT DEFAULT ''`); err != nil {
			return fmt.Errorf("failed to add owner_id column: %w", err)
		}
		log.Printf("[DB] Added owner_id column to resumes table")
	}

	// R33-H1: add capture_id column to users if it doesn't exist. PayPal
	// refund webhooks use capture ID as parent_id, but we previously only
	// stored the order ID in subscription_id — making refund lookups fail
	// and users retain paid plans after refunding.
	var capColCount int
	err = tx.QueryRow(`SELECT COUNT(*) FROM pragma_table_info('users') WHERE name='capture_id'`).Scan(&capColCount)
	if err != nil {
		return fmt.Errorf("failed to inspect users schema: %w", err)
	}
	if capColCount == 0 {
		if _, err := tx.Exec(`ALTER TABLE users ADD COLUMN capture_id TEXT DEFAULT ''`); err != nil {
			return fmt.Errorf("failed to add capture_id column: %w", err)
		}
		log.Printf("[DB] Added capture_id column to users table")
	}

	// PurchasedTemplates migration: stores the list of individually-purchased
	// resume templates (e.g. ["modern","creative"]) as a JSON array in TEXT.
	// Without this column, template purchases could not be persisted and the
	// "purchased" badge would never show after a restart. Mirrors the
	// capture_id backward-compatible migration above.
	var tplColCount int
	err = tx.QueryRow(`SELECT COUNT(*) FROM pragma_table_info('users') WHERE name='purchased_templates'`).Scan(&tplColCount)
	if err != nil {
		return fmt.Errorf("failed to inspect users schema: %w", err)
	}
	if tplColCount == 0 {
		if _, err := tx.Exec(`ALTER TABLE users ADD COLUMN purchased_templates TEXT DEFAULT '[]'`); err != nil {
			return fmt.Errorf("failed to add purchased_templates column: %w", err)
		}
		log.Printf("[DB] Added purchased_templates column to users table")
	}

	// Backward-compatible migration for template_captures (maps a template
	// purchase's PayPal capture ID -> template ID, used to reverse refunds).
	// Existing rows without the column keep the DEFAULT '[]' after ALTER.
	var tplCapColCount int
	err = tx.QueryRow(`SELECT COUNT(*) FROM pragma_table_info('users') WHERE name='template_captures'`).Scan(&tplCapColCount)
	if err != nil {
		return fmt.Errorf("failed to inspect users schema: %w", err)
	}
	if tplCapColCount == 0 {
		if _, err := tx.Exec(`ALTER TABLE users ADD COLUMN template_captures TEXT DEFAULT '[]'`); err != nil {
			return fmt.Errorf("failed to add template_captures column: %w", err)
		}
		log.Printf("[DB] Added template_captures column to users table")
	}

	// Now safe to create the owner index — owner_id is guaranteed to exist.
	if _, err := tx.Exec(`CREATE INDEX IF NOT EXISTS idx_resumes_owner ON resumes(owner_id)`); err != nil {
		return fmt.Errorf("failed to create resumes owner index: %w", err)
	}

	return tx.Commit()
}

func (d *Database) Close() error {
	// R49-B6: sync.Once makes Close idempotent. Without it, a second Close
	// call (e.g. from graceful shutdown + a defer in main) would panic on
	// close(d.done) — "close of closed channel". sync.Once guarantees the
	// channel is closed exactly once; subsequent calls are no-ops.
	var err error
	d.closeOnce.Do(func() {
		close(d.done)
		d.mu.Lock()
		err = d.db.Close()
		d.mu.Unlock()
	})
	return err
}

// Ping verifies the database connection is actually responsive, not just that
// the pointer is non-nil. AdminHealth previously reported "database: true"
// whenever h.db != nil, which masks a closed/stalled connection from
// monitoring. A 3s context bound prevents a hung DB from stalling the health
// endpoint itself (R53b-B3).
func (d *Database) Ping() error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	return d.db.PingContext(ctx)
}

type dbUser struct {
	ID               string
	Email            string
	Password         string
	PasswordType     string
	Name             string
	Token            string
	UsageCount       int
	MaxFreeUsage     int
	Plan             string
	SubscriptionID   string
	CaptureID        string
	PurchasedTemplates string
	TemplateCaptures   string
	// R50-B4: CreatedAt removed — scanUser uses a local variable since the
	// DB timestamp needs parsing before assignment to models.User.CreatedAt.
}

func (d *Database) scanUser(row interface{ Scan(...interface{}) error }) (*models.User, error) {
	var u dbUser
	var createdAt string
	err := row.Scan(&u.ID, &u.Email, &u.Password, &u.PasswordType, &u.Name, &u.Token,
		&u.UsageCount, &u.MaxFreeUsage, &u.Plan, &u.SubscriptionID, &u.CaptureID, &u.PurchasedTemplates, &u.TemplateCaptures, &createdAt)
	if err != nil {
		return nil, err
	}

	t, err := time.Parse(time.RFC3339, createdAt)
	if err != nil {
		// Fallback: SQLite CURRENT_TIMESTAMP format "2006-01-02 15:04:05"
		// (used when rows are inserted via DEFAULT instead of Go code).
		t, err = time.Parse("2006-01-02 15:04:05", createdAt)
	}
	if err != nil {
		log.Printf("[DB] unparseable created_at %q: %v", createdAt, err)
		t = time.Now()
	}

	// Parse the purchased_templates JSON array. An empty/malformed value
	// yields an empty (non-nil-safe) slice so callers can call includes.
	var purchased []string
	if strings.TrimSpace(u.PurchasedTemplates) != "" {
		if jErr := json.Unmarshal([]byte(u.PurchasedTemplates), &purchased); jErr != nil {
			log.Printf("[DB] unparseable purchased_templates %q: %v", u.PurchasedTemplates, jErr)
			purchased = []string{}
		}
	}
	if purchased == nil {
		purchased = []string{}
	}

	// Parse the template_captures JSON array (capture_id -> template_id).
	var tplCaps []models.TemplateCapture
	if strings.TrimSpace(u.TemplateCaptures) != "" {
		if jErr := json.Unmarshal([]byte(u.TemplateCaptures), &tplCaps); jErr != nil {
			log.Printf("[DB] unparseable template_captures %q: %v", u.TemplateCaptures, jErr)
			tplCaps = []models.TemplateCapture{}
		}
	}
	if tplCaps == nil {
		tplCaps = []models.TemplateCapture{}
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
		PurchasedTemplates: purchased,
		TemplateCaptures:   tplCaps,
		CreatedAt:        t,
	}, nil
}

func (d *Database) GetUser(email string) (*models.User, error) {
	d.mu.RLock()
	defer d.mu.RUnlock()

	row := d.db.QueryRow(
		`SELECT id, email, password, password_type, name, token, usage_count, max_free_usage, plan, subscription_id, capture_id, purchased_templates, template_captures, created_at
		 FROM users WHERE email = ?`, email)

	user, err := d.scanUser(row)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return user, err
}

func (d *Database) GetUserByToken(token string) (*models.User, error) {
	d.mu.RLock()
	defer d.mu.RUnlock()

	row := d.db.QueryRow(
		`SELECT id, email, password, password_type, name, token, usage_count, max_free_usage, plan, subscription_id, capture_id, purchased_templates, template_captures, created_at
		 FROM users WHERE token = ?`, token)

	user, err := d.scanUser(row)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return user, err
}

func (d *Database) GetAllUsers() ([]*models.User, error) {
	d.mu.RLock()
	defer d.mu.RUnlock()

	rows, err := d.db.Query(
		`SELECT id, email, password, password_type, name, token, usage_count, max_free_usage, plan, subscription_id, capture_id, purchased_templates, template_captures, created_at
		 FROM users`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []*models.User
	for rows.Next() {
		user, err := d.scanUser(rows)
		if err != nil {
			return nil, err
		}
		users = append(users, user)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return users, nil
}

func (d *Database) SaveUser(user *models.User) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	// Serialize purchased_templates to a JSON array for storage. A nil slice
	// becomes "[]" so the column is always valid JSON.
	tplJSON, err := json.Marshal(user.PurchasedTemplates)
	if err != nil {
		tplJSON = []byte("[]")
	}
	// Serialize template_captures (capture_id -> template_id) the same way.
	capJSON, err := json.Marshal(user.TemplateCaptures)
	if err != nil {
		capJSON = []byte("[]")
	}

	_, err = d.db.Exec(
		`INSERT OR REPLACE INTO users (id, email, password, password_type, name, token, usage_count, max_free_usage, plan, subscription_id, capture_id, purchased_templates, template_captures, created_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		user.ID, user.Email, user.Password, user.PasswordType, user.Name, user.Token,
		user.UsageCount, user.MaxFreeUsage, user.Plan, user.SubscriptionID, user.CaptureID,
		string(tplJSON), string(capJSON),
		user.CreatedAt.Format(time.RFC3339))
	return err
}

// UpdateUserUsage performs a targeted UPDATE of only usage_count for the
// given email. Unlike SaveUser (which does INSERT OR REPLACE on ALL columns),
// this does NOT overwrite token, plan, max_free_usage, or other fields that
// may have been concurrently modified by Login, webhook, etc.
// R37b-B2: max_free_usage removed — persistUsage's stale snapshot could
// overwrite a concurrent plan upgrade's new max_free_usage value. Since
// max_free_usage only changes via SaveUser (plan upgrades/downgrades),
// it should never be persisted through this path.
func (d *Database) UpdateUserUsage(email string, usageCount int) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	_, err := d.db.Exec(
		"UPDATE users SET usage_count = ? WHERE email = ?",
		usageCount, email)
	return err
}

// AdjustUserUsage performs an incremental UPDATE of usage_count.
// R52-B1: eliminates the stale-write race in persistUsage — concurrent
// goroutines each do usage_count = usage_count + delta, which is
// commutative and order-independent, unlike absolute writes.
func (d *Database) AdjustUserUsage(email string, delta int) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	_, err := d.db.Exec(
		"UPDATE users SET usage_count = usage_count + ? WHERE email = ?",
		delta, email)
	return err
}

// UpdateUserToken performs a targeted UPDATE of only the token column.
// R37-B1: auth paths (VerifyLogin existing user / Logout) previously used
// SaveUser (INSERT OR REPLACE on ALL columns), whose stale snapshot could
// overwrite a concurrent persistUsage increment of usage_count. This method
// only touches token, eliminating the clobber window.
func (d *Database) UpdateUserToken(email, token string) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	_, err := d.db.Exec(
		"UPDATE users SET token = ? WHERE email = ?",
		token, email)
	return err
}

// UpdateUserTokenAndPassword updates token plus (optionally) password and
// password_type in a single UPDATE. Used by Login, which rotates the token
// and may also upgrade a legacy SHA256 password to bcrypt. When password is
// empty, only the token is updated (password columns untouched).
func (d *Database) UpdateUserTokenAndPassword(email, token, password, passwordType string) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	var err error
	if password != "" {
		_, err = d.db.Exec(
			"UPDATE users SET token = ?, password = ?, password_type = ? WHERE email = ?",
			token, password, passwordType, email)
	} else {
		_, err = d.db.Exec(
			"UPDATE users SET token = ? WHERE email = ?",
			token, email)
	}
	return err
}

// UpdateUserPlan performs a targeted UPDATE of plan-related columns only
// (plan, subscription_id, capture_id, max_free_usage). R37-B1: payment paths
// (CapturePayPalOrder / webhook upgrade / webhook refund downgrade) previously
// used SaveUser, whose stale snapshot could overwrite a concurrent
// persistUsage increment of usage_count. This method only touches plan fields.
func (d *Database) UpdateUserPlan(email, plan, subscriptionID, captureID string, maxFreeUsage int) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	_, err := d.db.Exec(
		"UPDATE users SET plan = ?, subscription_id = ?, capture_id = ?, max_free_usage = ? WHERE email = ?",
		plan, subscriptionID, captureID, maxFreeUsage, email)
	return err
}

// UpdateUserTemplates performs a targeted UPDATE of only the
// purchased_templates and template_captures columns. R37-B1: payment paths
// (CaptureTemplateOrder / refund webhook) must not clobber concurrent
// usage_count increments, so we touch only these two columns. Both slices are
// stored as JSON arrays.
func (d *Database) UpdateUserTemplates(email string, templates []string, captures []models.TemplateCapture) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	tplJSON, err := json.Marshal(templates)
	if err != nil {
		tplJSON = []byte("[]")
	}
	capJSON, err := json.Marshal(captures)
	if err != nil {
		capJSON = []byte("[]")
	}

	_, err = d.db.Exec(
		"UPDATE users SET purchased_templates = ?, template_captures = ? WHERE email = ?",
		string(tplJSON), string(capJSON), email)
	return err
}

func (d *Database) DeleteUser(email string) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	_, err := d.db.Exec("DELETE FROM users WHERE email = ?", email)
	return err
}

type dbResume struct {
	ID        string
	OwnerID   string
	Title     string
	Content   string
	CreatedAt string
	UpdatedAt string
}

func (d *Database) GetResume(id string) (*models.Resume, error) {
	d.mu.RLock()
	defer d.mu.RUnlock()

	var r dbResume
	err := d.db.QueryRow(
		`SELECT id, owner_id, title, content, created_at, updated_at FROM resumes WHERE id = ?`, id).
		Scan(&r.ID, &r.OwnerID, &r.Title, &r.Content, &r.CreatedAt, &r.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	var content map[string]interface{}
	if err := json.Unmarshal([]byte(r.Content), &content); err != nil {
		log.Printf("[ERROR] sqlite.GetResume: corrupted content JSON for resume %s: %v", r.ID, err)
		return nil, fmt.Errorf("corrupted resume content for %s: %w", r.ID, err)
	}

	return &models.Resume{
		ID:        r.ID,
		OwnerID:   r.OwnerID,
		Title:     r.Title,
		Content:   content,
		CreatedAt: r.CreatedAt,
		UpdatedAt: r.UpdatedAt,
	}, nil
}

func (d *Database) SaveResume(resume *models.Resume) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	content, err := json.Marshal(resume.Content)
	if err != nil {
		return fmt.Errorf("failed to marshal resume content: %w", err)
	}
	if resume.CreatedAt == "" {
		resume.CreatedAt = time.Now().Format(time.RFC3339)
	}
	if resume.UpdatedAt == "" {
		resume.UpdatedAt = resume.CreatedAt
	}
	_, err = d.db.Exec(
		`INSERT OR REPLACE INTO resumes (id, owner_id, title, content, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?)`,
		resume.ID, resume.OwnerID, resume.Title, string(content), resume.CreatedAt, resume.UpdatedAt)
	return err
}

func (d *Database) DeleteResume(id string) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	_, err := d.db.Exec("DELETE FROM resumes WHERE id = ?", id)
	return err
}

// CountResumesByOwner returns the number of resumes owned by a given user.
// Used to enforce per-user resume limits in the Create endpoint, preventing
// a single user from exhausting SQLite storage or evicting other users'
// resumes from the in-memory LRU store.
func (d *Database) CountResumesByOwner(ownerID string) (int, error) {
	d.mu.RLock()
	defer d.mu.RUnlock()

	var count int
	err := d.db.QueryRow("SELECT COUNT(*) FROM resumes WHERE owner_id = ?", ownerID).Scan(&count)
	if err != nil {
		return 0, err
	}
	return count, nil
}

// ListResumesByOwner returns all resumes owned by a given user, ordered by
// most recently updated first. Excludes the content field — the List endpoint
// only needs metadata, and loading 50 resumes × 100KB content = 5MB would
// waste memory and slow the response. Callers needing content should use
// GetResume(id) for individual resumes.
func (d *Database) ListResumesByOwner(ownerID string) ([]*models.Resume, error) {
	d.mu.RLock()
	defer d.mu.RUnlock()

	rows, err := d.db.Query(
		"SELECT id, owner_id, title, created_at, updated_at FROM resumes WHERE owner_id = ? ORDER BY updated_at DESC",
		ownerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var resumes []*models.Resume
	for rows.Next() {
		var r models.Resume
		if err := rows.Scan(&r.ID, &r.OwnerID, &r.Title, &r.CreatedAt, &r.UpdatedAt); err != nil {
			return nil, err
		}
		resumes = append(resumes, &r)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return resumes, nil
}

// SaveResumeWithLimit atomically checks the owner's resume count and inserts
// if under the limit, within a single transaction. This prevents the TOCTOU
// race where two concurrent Create requests both pass CountResumesByOwner
// before either inserts — resulting in 51 resumes when the limit is 50.
// Returns ErrResumeLimitReached if the limit is exceeded.
var ErrResumeLimitReached = fmt.Errorf("resume limit reached")

func (d *Database) SaveResumeWithLimit(resume *models.Resume, maxCount int) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	tx, err := d.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	var count int
	if err := tx.QueryRow("SELECT COUNT(*) FROM resumes WHERE owner_id = ? AND id != ?", resume.OwnerID, resume.ID).Scan(&count); err != nil {
		return fmt.Errorf("failed to count resumes: %w", err)
	}
	if count >= maxCount {
		return ErrResumeLimitReached
	}

	content, err := json.Marshal(resume.Content)
	if err != nil {
		return fmt.Errorf("failed to marshal resume content: %w", err)
	}
	_, err = tx.Exec(
		`INSERT INTO resumes (id, owner_id, title, content, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?)`,
		resume.ID, resume.OwnerID, resume.Title, string(content), resume.CreatedAt, resume.UpdatedAt)
	if err != nil {
		return fmt.Errorf("failed to insert resume: %w", err)
	}

	return tx.Commit()
}

func (d *Database) GetAllResumes() ([]*models.Resume, error) {
	d.mu.RLock()
	defer d.mu.RUnlock()

	rows, err := d.db.Query("SELECT id, owner_id, title, content, created_at, updated_at FROM resumes")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	resumes := make([]*models.Resume, 0)
	for rows.Next() {
		var r dbResume
		if err := rows.Scan(&r.ID, &r.OwnerID, &r.Title, &r.Content, &r.CreatedAt, &r.UpdatedAt); err != nil {
			return nil, err
		}
		var content map[string]interface{}
		if err := json.Unmarshal([]byte(r.Content), &content); err != nil {
			log.Printf("[ERROR] sqlite.GetAllResumes: corrupted content JSON for resume %s: %v", r.ID, err)
			continue
		}
		resumes = append(resumes, &models.Resume{
			ID:        r.ID,
			OwnerID:   r.OwnerID,
			Title:     r.Title,
			Content:   content,
			CreatedAt: r.CreatedAt,
			UpdatedAt: r.UpdatedAt,
		})
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return resumes, nil
}

func (d *Database) MigrateFromJSON(users map[string]*models.User) error {
	if len(users) == 0 {
		return nil
	}

	log.Printf("[DB] Migrating %d users from JSON to SQLite", len(users))

	// Acquire the write lock to coordinate with concurrent SaveUser/SaveResume
	// callers — every other write method on Database takes d.mu, so this must
	// too, even though it currently only runs at startup.
	d.mu.Lock()
	defer d.mu.Unlock()

	tx, err := d.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare(
		`INSERT OR IGNORE INTO users (id, email, password, password_type, name, token, usage_count, max_free_usage, plan, subscription_id, capture_id, created_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, user := range users {
		passwordType := user.PasswordType
		if passwordType == "" {
			if strings.HasPrefix(user.Password, "$2a$") || strings.HasPrefix(user.Password, "$2b$") {
				passwordType = "bcrypt"
			} else {
				passwordType = "sha256"
			}
		}

		_, err := stmt.Exec(
			user.ID, user.Email, user.Password, passwordType, user.Name, user.Token,
			user.UsageCount, user.MaxFreeUsage, user.Plan, user.SubscriptionID,
			user.CaptureID, user.CreatedAt.Format(time.RFC3339))
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}
