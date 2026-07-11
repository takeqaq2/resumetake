package handlers

import (
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"log"
	"strings"
	"time"
	"unicode"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"resumetake/models"
	"resumetake/services"
)

type UserPersistence interface {
	Save(users map[string]*models.User) error
	// SaveUser persists a single user. Prefer this over Save when only one
	// user changed — Save re-writes the entire user table on every mutation.
	// R37-B1: prefer the targeted Update* methods below for existing-user
	// mutations — SaveUser does INSERT OR REPLACE on ALL columns and can
	// clobber a concurrent persistUsage increment with a stale snapshot.
	SaveUser(user *models.User) error
	// UpdateUserUsage performs a targeted UPDATE of only usage_count.
	// Prefer this over SaveUser for persistUsage to avoid clobbering
	// concurrent token/plan/max_free_usage updates with a stale snapshot.
	UpdateUserUsage(email string, usageCount int) error
	// AdjustUserUsage performs an incremental UPDATE (usage_count = usage_count + delta).
	// R52-B1: prefer this over UpdateUserUsage for persistUsage — the absolute-write
	// approach reads the in-memory UsageCount and writes it to DB, but concurrent
	// requests can persist stale values (e.g. A reads 5, B reads 6, A writes 5 last).
	// Incremental UPDATE is atomic at the DB level and order-independent.
	AdjustUserUsage(email string, delta int) error
	// UpdateUserToken performs a targeted UPDATE of only the token column.
	// Use for VerifyLogin (existing user) and Logout.
	UpdateUserToken(email, token string) error
	// UpdateUserTokenAndPassword updates token plus (optionally) password
	// and password_type. Use for Login (token rotation + SHA256→bcrypt
	// upgrade). When password is empty, only token is updated.
	UpdateUserTokenAndPassword(email, token, password, passwordType string) error
	// UpdateUserPlan performs a targeted UPDATE of plan, subscription_id,
	// capture_id, and max_free_usage. Use for CapturePayPalOrder and webhook
	// upgrade/downgrade — avoids clobbering concurrent usage_count increments.
	UpdateUserPlan(email, plan, subscriptionID, captureID string, maxFreeUsage int) error
	// UpdateUserTemplates performs a targeted UPDATE of only the
	// purchased_templates and template_captures columns. Use for
	// CaptureTemplateOrder and the template-refund webhook — avoids
	// clobbering concurrent usage_count/plan increments.
	UpdateUserTemplates(email string, templates []string, captures []models.TemplateCapture) error
}

// dummyBcryptHash is used to equalize timing on the Login "user not found"
// path. Without it, a missing email returns instantly while a wrong-password
// attempt pays the bcrypt cost (~100ms), allowing email enumeration by timing.
// Comparing against this dummy hash makes both paths perform one bcrypt op.
var dummyBcryptHash = func() []byte {
	h, err := bcrypt.GenerateFromPassword([]byte("timing-equalizer"), bcrypt.DefaultCost)
	if err != nil {
		// R54-B3: fail-fast — if bcrypt init fails (extreme OOM), the
		// timing equalization is silently lost, re-enabling email
		// enumeration. A panic makes this detectable at startup.
		panic("failed to initialize dummyBcryptHash: " + err.Error())
	}
	return h
}()

type AuthHandler struct {
	userStore         *models.UserStore
	verificationStore *models.VerificationStore
	persistence       UserPersistence
}

func NewAuthHandler(userStore *models.UserStore, verificationStore *models.VerificationStore, persistence UserPersistence) *AuthHandler {
	return &AuthHandler{
		userStore:         userStore,
		verificationStore: verificationStore,
		persistence:       persistence,
	}
}

func (h *AuthHandler) SendCode(c *fiber.Ctx) error {
	var body struct {
		Email string `json:"email"`
	}
	if err := c.BodyParser(&body); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "INVALID_BODY", "message": "Invalid request body"})
	}

	body.Email = strings.TrimSpace(strings.ToLower(body.Email))
	if len(body.Email) > 254 {
		return c.Status(400).JSON(fiber.Map{"error": "VALIDATION_ERROR", "message": "Email too long"})
	}
	if !services.IsValidEmail(body.Email) {
		return c.Status(400).JSON(fiber.Map{"error": "VALIDATION_ERROR", "message": "Invalid email format"})
	}

	// Use the same generic message regardless of whether the email is
	// registered to prevent email enumeration via response differences.
	const genericMsg = "If this email is available, a verification code has been sent."

	// Rate limit: one code per email per 60 seconds to prevent email bombing.
	// R52b-B1: AcquireResendSlot atomically checks cooldown AND reserves the
	// slot under a single Lock. The previous CanResend (RLock) + Save (Lock)
	// had a TOCTOU window: concurrent requests could all pass CanResend before
	// any set the cooldown, sending N emails to the victim.
	if !h.verificationStore.AcquireResendSlot(body.Email) {
		return c.Status(429).JSON(fiber.Map{"error": "RATE_LIMITED", "message": "Please wait before requesting another code"})
	}

	if _, exists := h.userStore.GetByEmail(body.Email); exists {
		// Timing equalization: the "not registered" path below performs code
		// generation + SMTP send (~100-500ms). Without equivalent work here,
		// an attacker can enumerate registered emails by response latency.
		// A bcrypt hash adds ~100ms CPU cost, narrowing the timing gap.
		// R50-B1: log bcrypt failures — if GenerateFromPassword fails (e.g.
		// OOM under load), the timing equalization is lost and the endpoint
		// becomes a timing oracle again. Logging makes this detectable.
		if _, err := bcrypt.GenerateFromPassword([]byte(body.Email), bcrypt.DefaultCost); err != nil {
			log.Printf("[WARN] timing equalization bcrypt failed for send-code: %v", err)
		}
		// Cooldown already set by AcquireResendSlot above — no need for
		// SetCooldown here. R27-H1 enumeration protection is preserved.
		return c.JSON(fiber.Map{"success": true, "message": genericMsg})
	}

	code, err := services.GenerateVerificationCode()
	if err != nil {
		log.Printf("[ERROR] Failed to generate verification code: %v", err)
		return c.Status(500).JSON(fiber.Map{"error": "INTERNAL_ERROR", "message": "Failed to generate verification code, please try again"})
	}

	// Send FIRST, then save on success. Previously Save ran before SMTP send,
	// so a transient SMTP failure (500) would still trip the 60s CanResend
	// cooldown — the user got no email yet was told to wait. Saving only on
	// success also blocks a denial-of-service vector where an attacker trips
	// SMTP failures on a victim's email to lock them out of registration.
	if err := services.SendVerificationEmail(body.Email, code); err != nil {
		log.Printf("[SMTP] Failed to send verification email: %v", err)
		// R54-B1: release the cooldown slot set by AcquireResendSlot so the
		// user can retry immediately once SMTP recovers. Without this, a
		// transient SMTP failure locks the user out for 60s.
		h.verificationStore.ReleaseResendSlot(body.Email)
		return c.Status(500).JSON(fiber.Map{"error": "EMAIL_SEND_FAILED", "message": "Failed to send verification email, please try again"})
	}
	h.verificationStore.Save(body.Email, code)

	return c.JSON(fiber.Map{"success": true, "message": genericMsg})
}

func (h *AuthHandler) VerifyCode(c *fiber.Ctx) error {
	var body struct {
		Email string `json:"email"`
		Code  string `json:"code"`
	}
	if err := c.BodyParser(&body); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "INVALID_BODY", "message": "Invalid request body"})
	}

	body.Email = strings.TrimSpace(strings.ToLower(body.Email))
	if len(body.Email) > 254 || len(body.Code) > 10 {
		return c.Status(400).JSON(fiber.Map{"error": "VALIDATION_ERROR", "message": "Invalid input"})
	}
	if !services.IsValidEmail(body.Email) {
		return c.Status(400).JSON(fiber.Map{"error": "VALIDATION_ERROR", "message": "Invalid email format"})
	}
	if h.verificationStore.Verify(body.Email, body.Code) {
		return c.JSON(fiber.Map{"success": true, "message": "Email verified"})
	}
	return c.Status(400).JSON(fiber.Map{"error": "INVALID_CODE", "message": "Invalid or expired verification code"})
}

func (h *AuthHandler) VerifyLogin(c *fiber.Ctx) error {
	var body struct {
		Email string `json:"email"`
		Code  string `json:"code"`
	}
	if err := c.BodyParser(&body); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "INVALID_BODY", "message": "Invalid request body"})
	}

	body.Email = strings.TrimSpace(strings.ToLower(body.Email))
	if len(body.Email) > 254 || len(body.Code) > 10 {
		return c.Status(400).JSON(fiber.Map{"error": "VALIDATION_ERROR", "message": "Invalid input"})
	}
	if !services.IsValidEmail(body.Email) {
		return c.Status(400).JSON(fiber.Map{"error": "VALIDATION_ERROR", "message": "Invalid email format"})
	}
	if !h.verificationStore.Verify(body.Email, body.Code) {
		return c.Status(400).JSON(fiber.Map{"error": "INVALID_CODE", "message": "Invalid or expired verification code"})
	}

	token, err := models.GenerateToken(body.Email)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "TOKEN_GENERATION_FAILED", "message": "Failed to generate secure token"})
	}

	user, exists := h.userStore.GetByEmail(body.Email)
	created := false
	if !exists {
		// New user: use SaveIfAbsent to atomically create only if no
		// concurrent Register/VerifyLogin already created this email.
		// A plain Save here would overwrite a concurrent Register's
		// password with an empty password (account takeover).
		name := body.Email
		if idx := strings.Index(body.Email, "@"); idx > 0 {
			name = body.Email[:idx]
		}
		user = &models.User{
			ID:           uuid.New().String(),
			Email:        body.Email,
			Name:         name,
			MaxFreeUsage: 5,
			Plan:         "free",
			Token:        token,
			CreatedAt:    time.Now(),
		}
		if h.userStore.SaveIfAbsent(user) {
			created = true
		} else {
			// A concurrent request created this user between our
			// GetByEmail and SaveIfAbsent. Rotate the token atomically
			// on the existing user instead.
			updated, ok := h.userStore.UpdateUser(body.Email, func(u *models.User) {
				u.Token = token
			})
			if !ok {
				return c.Status(500).JSON(fiber.Map{"error": "INTERNAL_ERROR", "message": "Failed to update user session"})
			}
			user = updated
		}
	} else {
		// Existing user: rotate token atomically (consistent with Login).
		updated, ok := h.userStore.UpdateUser(body.Email, func(u *models.User) {
			u.Token = token
		})
		if !ok {
			return c.Status(500).JSON(fiber.Map{"error": "INTERNAL_ERROR", "message": "Failed to update user session"})
		}
		user = updated
	}
	if h.persistence != nil {
		// R37-B1: existing users use targeted UPDATE so a stale snapshot
		// doesn't clobber a concurrent persistUsage increment of
		// usage_count. New users still need SaveUser (INSERT).
		// R37-A1: use `created` instead of `exists` — SaveIfAbsent failure
		// means the user already exists, so UpdateUserToken is correct.
		if created {
			if err := h.persistence.SaveUser(user); err != nil {
				log.Printf("[ERROR] Failed to persist user: %v", err)
			}
		} else {
			if err := h.persistence.UpdateUserToken(user.Email, user.Token); err != nil {
				log.Printf("[ERROR] Failed to persist user token: %v", err)
			}
		}
	}

	// Use sanitizeUser for consistency with Login/Register/Me endpoints.
	// Previously this manually constructed a partial map missing
	// subscription_id and created_at, causing frontend display issues.
	response := sanitizeUser(user)
	response["token"] = token
	return c.JSON(fiber.Map{"success": true, "data": response})
}

func (h *AuthHandler) Register(c *fiber.Ctx) error {
	var body struct {
		Email    string `json:"email"`
		Password string `json:"password"`
		Name     string `json:"name"`
	}
	if err := c.BodyParser(&body); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "INVALID_BODY", "message": "Invalid request body"})
	}

	body.Email = strings.TrimSpace(strings.ToLower(body.Email))
	body.Name = strings.TrimSpace(body.Name)

	if !services.IsValidEmail(body.Email) {
		return c.Status(400).JSON(fiber.Map{"error": "VALIDATION_ERROR", "message": "Invalid email format"})
	}

	if len(body.Password) < 6 {
		return c.Status(400).JSON(fiber.Map{"error": "VALIDATION_ERROR", "message": "Password must be at least 6 characters"})
	}
	// bcrypt silently truncates passwords beyond 72 bytes; reject explicitly
	// to avoid two different passwords hashing to the same value.
	if len(body.Password) > 72 {
		return c.Status(400).JSON(fiber.Map{"error": "VALIDATION_ERROR", "message": "Password must be at most 72 characters"})
	}
	// Require at least one letter and one digit so trivially weak passwords
	// (e.g. "123456", "aaaaaa") are rejected at registration time.
	hasLetter := false
	hasDigit := false
	for _, r := range body.Password {
		// R54-B4: use unicode.IsLetter/IsDigit instead of ASCII-only check —
		// passwords with CJK/accented letters (e.g. "café123") were rejected.
		if unicode.IsLetter(r) {
			hasLetter = true
		}
		if unicode.IsDigit(r) {
			hasDigit = true
		}
	}
	if !hasLetter || !hasDigit {
		return c.Status(400).JSON(fiber.Map{"error": "VALIDATION_ERROR", "message": "Password must contain at least one letter and one number"})
	}
	// R54b-B4: removed redundant len(body.Email) > 254 check — IsValidEmail
	// (called above) already rejects emails >254 chars, so this branch was
	// dead code that could never trigger.

	if body.Name == "" {
		if idx := strings.Index(body.Email, "@"); idx > 0 {
			body.Name = body.Email[:idx]
		} else {
			body.Name = body.Email
		}
	}
	// Limit name length to prevent storage bloat and oversized API responses.
	if len(body.Name) > 200 {
		return c.Status(400).JSON(fiber.Map{"error": "VALIDATION_ERROR", "message": "Name is too long (max 200 characters)"})
	}

	// Fast-path check: if the email already exists, return 409 without
	// paying the bcrypt cost. This is an optimization, not a race guard —
	// the authoritative uniqueness check is SaveIfAbsent below.
	if _, exists := h.userStore.GetByEmail(body.Email); exists {
		return c.Status(409).JSON(fiber.Map{"error": "CONFLICT", "message": "Email already registered"})
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(body.Password), bcrypt.DefaultCost)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "INTERNAL_ERROR", "message": "Failed to hash password"})
	}

	token, err := models.GenerateToken(body.Email)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "TOKEN_GENERATION_FAILED", "message": "Failed to generate secure token"})
	}

	user := &models.User{
		ID:           uuid.New().String(),
		Email:        body.Email,
		Password:     string(hash),
		PasswordType: "bcrypt",
		Name:         body.Name,
		Token:        token,
		Plan:         "free",
		MaxFreeUsage: 5,
		CreatedAt:    time.Now(),
	}

	// Atomically insert only if no user with this email exists. This closes
	// the check-then-act race between the GetByEmail above and the insert:
	// two concurrent registrations for the same email would both pass the
	// GetByEmail check, but only the first SaveIfAbsent succeeds; the
	// second returns false and gets a 409. Without this, the second Save
	// would overwrite the first user's password (account takeover).
	if !h.userStore.SaveIfAbsent(user) {
		return c.Status(409).JSON(fiber.Map{"error": "CONFLICT", "message": "Email already registered"})
	}
	if h.persistence != nil {
		if err := h.persistence.SaveUser(user); err != nil {
			log.Printf("[ERROR] Failed to persist user: %v", err)
		}
	}

	return c.Status(201).JSON(fiber.Map{"success": true, "data": sanitizeUser(user)})
}

func (h *AuthHandler) Login(c *fiber.Ctx) error {
	var body struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := c.BodyParser(&body); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "INVALID_BODY", "message": "Invalid request body"})
	}

	body.Email = strings.TrimSpace(strings.ToLower(body.Email))
	// Validate email format/length early. Other auth endpoints (SendCode,
	// VerifyCode, VerifyLogin, Register) already do this; Login was the only
	// one missing it. Reject before bcrypt to avoid wasted CPU on junk input.
	// Returns the same INVALID_CREDENTIALS message (with timing equalization)
	// so the endpoint does not leak whether the email is registered.
	if !services.IsValidEmail(body.Email) || len(body.Email) > 254 {
		_ = bcrypt.CompareHashAndPassword(dummyBcryptHash, []byte(body.Password))
		return c.Status(401).JSON(fiber.Map{"error": "INVALID_CREDENTIALS", "message": "Invalid email or password"})
	}
	// Reject overly long passwords early — bcrypt truncates at 72 bytes, so
	// a 10MB password (global body limit) wastes CPU/memory on allocation
	// and copying before bcrypt even runs. Register already enforces 72.
	if len(body.Password) > 72 {
		_ = bcrypt.CompareHashAndPassword(dummyBcryptHash, []byte(body.Password))
		return c.Status(401).JSON(fiber.Map{"error": "INVALID_CREDENTIALS", "message": "Invalid email or password"})
	}
	user, ok := h.userStore.GetByEmail(body.Email)
	if !ok {
		// Timing equalization: perform a bcrypt comparison against a dummy
		// hash so this path takes roughly the same time as a wrong-password
		// attempt, preventing email enumeration via response latency.
		_ = bcrypt.CompareHashAndPassword(dummyBcryptHash, []byte(body.Password))
		return c.Status(401).JSON(fiber.Map{"error": "INVALID_CREDENTIALS", "message": "Invalid email or password"})
	}

	var passwordValid bool
	var upgradedHash []byte // non-nil if SHA256→bcrypt upgrade is needed

	if user.PasswordType == "bcrypt" {
		err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(body.Password))
		passwordValid = (err == nil)
	} else {
		hash := sha256.Sum256([]byte(body.Password))
		// Constant-time comparison to prevent timing attacks on the legacy
		// SHA256 password fallback path. bcrypt (above) is already constant-time.
		expected := hex.EncodeToString(hash[:])
		passwordValid = subtle.ConstantTimeCompare([]byte(user.Password), []byte(expected)) == 1
		if passwordValid {
			// Precompute bcrypt hash outside the store lock (CPU-intensive).
			if newHash, err := bcrypt.GenerateFromPassword([]byte(body.Password), bcrypt.DefaultCost); err == nil {
				upgradedHash = newHash
			} else {
				// bcrypt generation failure is extremely rare (typically OOM),
				// but if it happens the user's password stays as weak SHA256.
				// Log a warning so ops can investigate; the login itself
				// still succeeds since the password was verified correct.
				log.Printf("[WARN] bcrypt upgrade failed for %s: %v (password remains SHA256)", hashEmail(body.Email), err)
			}
		}
	}

	if !passwordValid {
		return c.Status(401).JSON(fiber.Map{"error": "INVALID_CREDENTIALS", "message": "Invalid email or password"})
	}

	// Use UpdateUser for atomic token rotation + password upgrade (consistent
	// with R2-1 pattern). Avoids the race where `user.Password = ...` writes
	// to the shared *User pointer outside the store's write lock.
	newToken, err := models.GenerateToken(body.Email)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "TOKEN_GENERATION_FAILED", "message": "Failed to generate secure token"})
	}
	updated, ok := h.userStore.UpdateUser(user.Email, func(u *models.User) {
		u.Token = newToken
		if upgradedHash != nil {
			u.Password = string(upgradedHash)
			u.PasswordType = "bcrypt"
		}
	})
	if !ok {
		return c.Status(500).JSON(fiber.Map{"error": "INTERNAL_ERROR", "message": "Failed to update user session"})
	}
	user = updated
	if h.persistence != nil {
		// R37-B1: targeted UPDATE of token + optional password upgrade,
		// avoiding SaveUser's INSERT OR REPLACE which clobbers usage_count.
		pwd := ""
		ptype := ""
		if upgradedHash != nil {
			pwd = string(upgradedHash)
			ptype = "bcrypt"
		}
		if err := h.persistence.UpdateUserTokenAndPassword(user.Email, user.Token, pwd, ptype); err != nil {
			log.Printf("[ERROR] Failed to persist login: %v", err)
		}
	}

	return c.JSON(fiber.Map{"success": true, "data": sanitizeUser(user)})
}

func (h *AuthHandler) Me(c *fiber.Ctx) error {
	user, ok := c.Locals("user").(*models.User)
	if !ok || user == nil {
		return c.Status(401).JSON(fiber.Map{"error": "UNAUTHORIZED", "message": "Authentication required"})
	}
	// R48-B2: use sanitizeUserPublic (excludes token) — the client already
	// has the token in localStorage (sent it in the Authorization header for
	// this request). Returning it again risks leakage via proxy/CDN/APM logs.
	return c.JSON(fiber.Map{"success": true, "data": sanitizeUserPublic(user)})
}

// Logout invalidates the user's token server-side so it can't be reused after
// logout. Without this, a token intercepted from browser history or logs remains
// valid until the user logs in again (which generates a new token).
func (h *AuthHandler) Logout(c *fiber.Ctx) error {
	user, ok := c.Locals("user").(*models.User)
	if !ok || user == nil {
		return c.Status(401).JSON(fiber.Map{"error": "UNAUTHORIZED", "message": "Authentication required"})
	}
	updated, ok := h.userStore.UpdateUser(user.Email, func(u *models.User) {
		u.Token = ""
	})
	if ok && h.persistence != nil {
		// R37-B1: targeted UPDATE of only token, avoiding SaveUser's
		// INSERT OR REPLACE which clobbers usage_count.
		if err := h.persistence.UpdateUserToken(user.Email, updated.Token); err != nil {
			log.Printf("[ERROR] Failed to persist logout: %v", err)
		}
	}
	return c.JSON(fiber.Map{"success": true})
}

func sanitizeUser(u *models.User) fiber.Map {
	return fiber.Map{
		"id":                 u.ID,
		"email":              u.Email,
		"name":               u.Name,
		"token":              u.Token,
		"plan":               u.Plan,
		"usage_count":        u.UsageCount,
		"max_free_usage":     u.MaxFreeUsage,
		"subscription_id":    u.SubscriptionID,
		"purchased_templates": u.PurchasedTemplates,
		"created_at":         u.CreatedAt,
	}
}

// sanitizeUserPublic returns user data without the token field. Used by /auth/me
// where the client already possesses the token (sent in Authorization header).
// R48-B2: prevents token leakage via proxy/CDN/APM response body logging.
func sanitizeUserPublic(u *models.User) fiber.Map {
	return fiber.Map{
		"id":                 u.ID,
		"email":              u.Email,
		"name":               u.Name,
		"plan":               u.Plan,
		"usage_count":        u.UsageCount,
		"max_free_usage":     u.MaxFreeUsage,
		"subscription_id":    u.SubscriptionID,
		"purchased_templates": u.PurchasedTemplates,
		"created_at":         u.CreatedAt,
	}
}
