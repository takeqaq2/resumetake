package handlers

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"resumetake/database"
	"resumetake/models"
	"resumetake/services"
)

type ResumeHandler struct {
	resumeStore *models.Store
	userStore   *models.UserStore
	db          *database.Database
}

func NewResumeHandler(resumeStore *models.Store, userStore *models.UserStore, db *database.Database) *ResumeHandler {
	return &ResumeHandler{
		resumeStore: resumeStore,
		userStore:   userStore,
		db:          db,
	}
}

// bridgeCtx creates a standard context.Context that is cancelled when the
// fasthttp connection closes (client disconnect). Fiber's c.UserContext()
// returns context.Background() which never cancels, so AI calls made with it
// would continue running — wasting upstream API cost and leaking goroutines —
// after the client has already disconnected. The returned CancelFunc must be
// deferred by the caller to release the watcher goroutine on normal exit.
func bridgeCtx(c *fiber.Ctx) (context.Context, context.CancelFunc) {
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		select {
		case <-c.Context().Done():
			cancel()
		case <-ctx.Done():
		}
	}()
	return ctx, cancel
}

// hashEmail returns a short hash prefix of the email for use in log messages.
// Full emails in logs create privacy/GDPR exposure — a 12-char SHA-256 prefix
// is sufficient for correlation across log lines without identifying the user.
func hashEmail(email string) string {
	h := sha256.Sum256([]byte(email))
	return hex.EncodeToString(h[:])[:12]
}

func (h *ResumeHandler) Create(c *fiber.Ctx) error {
	var body map[string]interface{}
	if err := c.BodyParser(&body); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "INVALID_BODY", "message": "Invalid request body"})
	}

	title, _ := body["title"].(string)
	title = strings.TrimSpace(title) // R51-B3: trim before storage
	if title == "" {
		return c.Status(400).JSON(fiber.Map{"error": "VALIDATION_ERROR", "message": "title is required"})
	}
	if len(title) > 200 {
		return c.Status(400).JSON(fiber.Map{"error": "VALIDATION_ERROR", "message": "title too long"})
	}

	id := uuid.New().String()
	now := time.Now().Format(time.RFC3339)

	content, ok := body["content"].(map[string]interface{})
	if !ok {
		content = map[string]interface{}{}
	}
	// Cap the marshaled content size — without this, an authenticated user
	// can store ~10MB (global body limit) of arbitrary JSON per resume,
	// exhausting SQLite storage and evicting other users' resumes from the
	// in-memory store (MaxResumes LRU). 100KB is generous for any real
	// resume structure.
	contentJSON, marshalErr := json.Marshal(content)
	if marshalErr != nil {
		// R54b-B3: previously the error was silently discarded (err == nil
		// guard), which meant a malformed content payload would skip the
		// size check and be persisted — hiding a real bug in the client or
		// a corrupt state. Log so ops can detect it.
		log.Printf("[ERROR] SaveResume json.Marshal failed: %v", marshalErr)
	} else if len(contentJSON) > 100*1024 {
		return c.Status(400).JSON(fiber.Map{"error": "CONTENT_TOO_LARGE", "message": "Resume content too large (max 100KB)"})
	}

	user, ok := c.Locals("user").(*models.User)
	if !ok || user == nil {
		return c.Status(401).JSON(fiber.Map{"error": "UNAUTHORIZED", "message": "Authentication required"})
	}

	// Per-user resume count limit — without this, a single authenticated
	// user can create unlimited resumes (each up to 100KB), exhausting
	// SQLite storage and evicting other users' resumes from the in-memory
	// LRU store (MaxResumes=5000). 50 is generous for real usage.
	// SaveResumeWithLimit does the count check + insert atomically in a
	// single transaction to prevent TOCTOU races (two concurrent Creates
	// both passing the count check before either inserts).
	const maxResumesPerUser = 50

	resume := &models.Resume{
		ID:        id,
		OwnerID:   user.ID,
		Title:     title,
		Content:   content,
		CreatedAt: now,
		UpdatedAt: now,
	}

	// Write to SQLite (authoritative storage) first. If this fails, the
	// resume would only exist in memory and be lost on restart — returning
	// 500 lets the user retry instead of silently losing data.
	if h.db != nil {
		if err := h.db.SaveResumeWithLimit(resume, maxResumesPerUser); err != nil {
			if errors.Is(err, database.ErrResumeLimitReached) {
				return c.Status(400).JSON(fiber.Map{"error": "RESUME_LIMIT_REACHED", "message": "You have reached the maximum number of resumes. Please delete an existing resume first."})
			}
			log.Printf("[ERROR] failed to persist resume %s: %v", resume.ID, err)
			return c.Status(500).JSON(fiber.Map{"error": "PERSIST_FAILED", "message": "Failed to save resume, please try again"})
		}
	}
	h.resumeStore.Save(resume)
	return c.Status(201).JSON(fiber.Map{"success": true, "data": resume})
}

// persistUsage asynchronously writes the usage delta to SQLite. Without this,
// AI usage counts are lost on restart — free users could reuse their quota
// after each deploy. R52-B1: uses AdjustUserUsage (incremental UPDATE) instead
// of UpdateUserUsage (absolute write) to eliminate the stale-write race where
// concurrent goroutines could persist outdated UsageCount values.
// Failures are logged but do not affect the response (memory state is already correct).
func (h *ResumeHandler) persistUsage(email string, delta int) {
	if h.db == nil {
		return
	}
	go func() {
		defer func() {
			if r := recover(); r != nil {
				log.Printf("[PANIC] persistUsage recovered: %v", r)
			}
		}()
		if err := h.db.AdjustUserUsage(email, delta); err != nil {
			log.Printf("[ERROR] failed to persist usage for %s: %v", hashEmail(email), err)
		}
	}()
}

func (h *ResumeHandler) Get(c *fiber.Ctx) error {
	id := c.Params("id")
	user, ok := c.Locals("user").(*models.User)
	if !ok || user == nil {
		return c.Status(401).JSON(fiber.Map{"error": "UNAUTHORIZED", "message": "Authentication required"})
	}

	r, ok := h.resumeStore.Get(id)
	if !ok {
		// LRU cache miss — fall back to SQLite. Without this, resumes
		// evicted from the in-memory store (MaxResumes=5000) become
		// permanently inaccessible even though they still exist in the DB.
		if h.db != nil {
			dbResume, err := h.db.GetResume(id)
			if err != nil {
				log.Printf("[ERROR] sqlite.GetResume %s: %v", id, err)
				return c.Status(500).JSON(fiber.Map{"error": "DB_ERROR", "message": "Failed to retrieve resume"})
			}
			if dbResume == nil {
				return c.Status(404).JSON(fiber.Map{"error": "NOT_FOUND", "message": "Resume not found"})
			}
			r = dbResume
		} else {
			return c.Status(404).JSON(fiber.Map{"error": "NOT_FOUND", "message": "Resume not found"})
		}
	}
	if r.OwnerID != user.ID {
		return c.Status(403).JSON(fiber.Map{"error": "FORBIDDEN", "message": "You do not have access to this resume"})
	}
	// Re-warm the memory cache only AFTER the ownership check passes.
	// Previously this was before the check, allowing any authenticated
	// user to pollute the LRU cache by requesting arbitrary UUIDs,
	// evicting other users' cached resumes.
	if !ok {
		h.resumeStore.Save(r)
	}
	return c.JSON(fiber.Map{"success": true, "data": r})
}

func (h *ResumeHandler) Delete(c *fiber.Ctx) error {
	id := c.Params("id")
	user, ok := c.Locals("user").(*models.User)
	if !ok || user == nil {
		return c.Status(401).JSON(fiber.Map{"error": "UNAUTHORIZED", "message": "Authentication required"})
	}

	r, ok := h.resumeStore.Get(id)
	if !ok && h.db != nil {
		// LRU cache miss — check SQLite for ownership verification.
		dbResume, err := h.db.GetResume(id)
		if err != nil {
			log.Printf("[ERROR] sqlite.GetResume %s: %v", id, err)
			return c.Status(500).JSON(fiber.Map{"error": "DB_ERROR", "message": "Failed to retrieve resume"})
		}
		if dbResume == nil {
			return c.Status(404).JSON(fiber.Map{"error": "NOT_FOUND", "message": "Resume not found"})
		}
		r = dbResume
	} else if !ok {
		return c.Status(404).JSON(fiber.Map{"error": "NOT_FOUND", "message": "Resume not found"})
	}

	if r.OwnerID != user.ID {
		return c.Status(403).JSON(fiber.Map{"error": "FORBIDDEN", "message": "You do not have access to this resume"})
	}

	// Delete from DB first (authoritative store), then memory cache. The
	// prior order (memory → DB) caused "resurrection": if DB delete failed
	// and returned 500, the memory entry was already gone, but Get's
	// cache-miss path reloads from SQLite — so the user would see the
	// deleted resume reappear on their next request.
	if h.db != nil {
		if err := h.db.DeleteResume(id); err != nil {
			log.Printf("[ERROR] failed to delete resume %s from db: %v", id, err)
			return c.Status(500).JSON(fiber.Map{"error": "DELETE_FAILED", "message": "Failed to delete resume"})
		}
	}
	h.resumeStore.Delete(id)
	return c.JSON(fiber.Map{"success": true, "message": "Deleted"})
}

// List returns all resumes owned by the authenticated user. Previously users
// had to know a resume ID to access it — there was no way to list their own
// resumes. This endpoint returns metadata (id, title, timestamps) without
// the full content to keep the response lightweight.
func (h *ResumeHandler) List(c *fiber.Ctx) error {
	user, ok := c.Locals("user").(*models.User)
	if !ok || user == nil {
		return c.Status(401).JSON(fiber.Map{"error": "UNAUTHORIZED", "message": "Authentication required"})
	}

	var summaries []fiber.Map
	if h.db != nil {
		resumes, err := h.db.ListResumesByOwner(user.ID)
		if err != nil {
			log.Printf("[ERROR] failed to list resumes for owner %s: %v", hashEmail(user.Email), err)
			return c.Status(500).JSON(fiber.Map{"error": "DB_ERROR", "message": "Failed to list resumes"})
		}
		for _, r := range resumes {
			summaries = append(summaries, fiber.Map{
				"id":         r.ID,
				"title":      r.Title,
				"created_at": r.CreatedAt,
				"updated_at": r.UpdatedAt,
			})
		}
	} else {
		// Fallback: scan in-memory store (filtered by owner)
		all := h.resumeStore.GetAll()
		for _, r := range all {
			if r.OwnerID == user.ID {
				summaries = append(summaries, fiber.Map{
					"id":         r.ID,
					"title":      r.Title,
					"created_at": r.CreatedAt,
					"updated_at": r.UpdatedAt,
				})
			}
		}
	}
	if summaries == nil {
		summaries = []fiber.Map{}
	}
	return c.JSON(fiber.Map{"success": true, "data": summaries})
}

func (h *ResumeHandler) Upload(c *fiber.Ctx) error {
	file, err := c.FormFile("file")
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "NO_FILE", "message": "No file uploaded"})
	}

	if file.Size <= 0 {
		return c.Status(400).JSON(fiber.Map{"error": "EMPTY_FILE", "message": "File is empty"})
	}

	filename := filepath.Base(file.Filename)
	if filename != file.Filename || strings.Contains(filename, "\x00") {
		return c.Status(400).JSON(fiber.Map{"error": "INVALID_FILENAME", "message": "Invalid filename"})
	}

	if strings.ContainsAny(filename, `/\`) {
		return c.Status(400).JSON(fiber.Map{"error": "INVALID_FILENAME", "message": "Invalid filename"})
	}

	ext := strings.ToLower(filepath.Ext(filename))
	if !models.AllowedUploadExt[ext] {
		return c.Status(400).JSON(fiber.Map{"error": "INVALID_TYPE", "message": "Only .txt, .md and .pdf files are supported"})
	}

	isPDF := ext == ".pdf"
	if isPDF && file.Size > models.MaxPDFBytes {
		return c.Status(400).JSON(fiber.Map{"error": "FILE_TOO_LARGE", "message": "PDF too large (max 2MB)"})
	}
	if !isPDF && file.Size > models.MaxUploadBytes {
		return c.Status(400).JSON(fiber.Map{"error": "FILE_TOO_LARGE", "message": "File too large (max 1MB)"})
	}

	contentType := strings.ToLower(file.Header.Get("Content-Type"))
	if isPDF {
		if contentType != "" && contentType != "application/pdf" {
			return c.Status(400).JSON(fiber.Map{"error": "INVALID_MIME", "message": "Invalid PDF file"})
		}
	} else {
		allowedMime := strings.HasPrefix(contentType, "text/plain") || contentType == "text/markdown" || contentType == "application/octet-stream"
		if contentType != "" && !allowedMime {
			return c.Status(400).JSON(fiber.Map{"error": "INVALID_MIME", "message": "Only plain text and markdown files are supported"})
		}
	}

	f, err := file.Open()
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "READ_ERROR", "message": "Failed to read file"})
	}
	defer f.Close()

	maxRead := int64(models.MaxUploadBytes + 1)
	if isPDF {
		maxRead = int64(models.MaxPDFBytes + 1)
	}

	rawData, err := io.ReadAll(io.LimitReader(f, maxRead))
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "READ_ERROR", "message": "Failed to read file content"})
	}

	if int64(len(rawData)) > maxRead-1 {
		if isPDF {
			return c.Status(400).JSON(fiber.Map{"error": "FILE_TOO_LARGE", "message": "PDF too large (max 2MB)"})
		}
		return c.Status(400).JSON(fiber.Map{"error": "FILE_TOO_LARGE", "message": "File too large (max 1MB)"})
	}

	if isPDF {
		if !services.ValidatePDFHeader(rawData) {
			return c.Status(400).JSON(fiber.Map{"error": "INVALID_PDF", "message": "Not a valid PDF file"})
		}
		text, pageCount, pdfErr := services.ExtractPDFTextWithOCR(rawData)
		if pdfErr != nil {
			log.Printf("[ERROR] PDF parse failed: %v", pdfErr)
			return c.Status(400).JSON(fiber.Map{"error": "PDF_PARSE_ERROR", "message": "Failed to parse PDF file"})
		}
		if text == "" {
			return c.Status(400).JSON(fiber.Map{"error": "EMPTY_CONTENT", "message": "PDF contains no extractable text"})
		}
		// R57-B5: PDF text extraction (especially OCR) can produce invalid
		// UTF-8 bytes from corrupted fonts or OCR misreads. Without this
		// check, invalid bytes propagate to JSON marshaling (causing
		// encoding errors) and AI providers (causing API failures).
		// strings.ToValidUTF8 replaces invalid sequences with U+FFFD.
		if !utf8.ValidString(text) {
			text = strings.ToValidUTF8(text, "\uFFFD")
		}
		if len(text) > models.MaxPDFTextChars {
			text = services.TruncateUTF8(text, models.MaxPDFTextChars)
		}
		return c.JSON(fiber.Map{
			"success": true,
			"data": map[string]interface{}{
				"filename":  filename,
				"size":      file.Size,
				"text":      text,
				"pageCount": pageCount,
			},
		})
	}

	if bytes.Contains(rawData, []byte{0}) || !utf8.Valid(rawData) {
		return c.Status(400).JSON(fiber.Map{"error": "UNSAFE_CONTENT", "message": "Only valid UTF-8 text files are supported"})
	}

	text := string(rawData)
	text = strings.ReplaceAll(text, "\r\n", "\n")
	text = strings.TrimSpace(text)

	if text == "" {
		return c.Status(400).JSON(fiber.Map{"error": "EMPTY_CONTENT", "message": "Resume text is empty"})
	}

	if len(text) > 15000 {
		text = services.TruncateUTF8(text, 15000)
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data": map[string]interface{}{
			"filename": filename,
			"size":     file.Size,
			"text":     text,
		},
	})
}

func (h *ResumeHandler) Optimize(c *fiber.Ctx) error {
	user, ok := c.Locals("user").(*models.User)
	if !ok || user == nil {
		return c.Status(401).JSON(fiber.Map{"error": "UNAUTHORIZED", "message": "Authentication required"})
	}

	var body map[string]interface{}
	if err := c.BodyParser(&body); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "INVALID_BODY", "message": "Invalid request body"})
	}

	lang, _ := body["lang"].(string)
	if lang == "" {
		lang = "en"
	}
	if !services.IsValidLang(lang) {
		lang = "en"
	}

	targetJob, _ := body["target_job"].(string)
	if targetJob == "" {
		targetJob, _ = body["target_position"].(string)
	}
	jobDesc, _ := body["job_description"].(string)

	if len(targetJob) > 500 {
		targetJob = services.TruncateUTF8(targetJob, 500)
	}
	if len(jobDesc) > 2000 {
		jobDesc = services.TruncateUTF8(jobDesc, 2000)
	}

	var resumeContent string
	if rc, ok := body["resume_text"].(string); ok && rc != "" {
		resumeContent = rc
	} else if rc, ok := body["resume"].(string); ok && rc != "" {
		resumeContent = rc
	} else if rc, ok := body["resume_content"].(string); ok && rc != "" {
		// R30-H5: handle string resume_content directly — json.Marshal on a
		// string produces "\"text\"" (with literal quotes), double-encoding
		// the content and degrading AI output quality.
		resumeContent = rc
	} else if rc, mErr := json.Marshal(body["resume_content"]); mErr != nil {
		log.Printf("[ERROR] json.Marshal resume_content failed: %v", mErr)
	} else if rc != nil && string(rc) != "null" {
		resumeContent = string(rc)
	}

	if len(resumeContent) > models.MaxResumeChars {
		resumeContent = services.TruncateUTF8(resumeContent, models.MaxResumeChars)
	}
	if resumeContent == "" {
		return c.Status(400).JSON(fiber.Map{"error": "NO_CONTENT", "message": "Resume content is required"})
	}

	modules, _ := body["modules"].([]interface{})
	if len(modules) == 0 {
		modules = []interface{}{"ats", "star", "quant", "summary", "format"}
	}
	// R51-B6: cap modules slice — only 5 known module names are matched by
	// BuildModuleHints, but a malicious client could send thousands of
	// entries to trigger large map allocation. Matches GenerateResume's
	// len > 50 guard on messages.
	if len(modules) > 20 {
		modules = modules[:20]
	}

	moduleHints := services.BuildModuleHints(modules)
	systemPrompt := services.GetPrompt(lang)

	if cached, ok := services.GetCachedResult(resumeContent, jobDesc, targetJob, moduleHints, lang); ok {
		return c.JSON(fiber.Map{"success": true, "data": cached, "cached": true, "usage_count": user.UsageCount, "max_free_usage": user.MaxFreeUsage})
	}

	dedupeKey := services.GenerateDedupeKey(user.ID, resumeContent, jobDesc, targetJob, moduleHints, lang)
	acquired, dedupToken, dedupErr := services.TryAcquireRequest(dedupeKey)
	if dedupErr != nil {
		return c.Status(503).JSON(fiber.Map{"error": "DEDUP_UNAVAILABLE", "message": "Temporary server failure, please retry"})
	}
	if !acquired {
		return c.Status(429).JSON(fiber.Map{"error": "REQUEST_IN_PROGRESS", "message": "Same request already in progress"})
	}
	defer services.ReleaseRequest(dedupeKey, dedupToken)

	// Atomically check quota AND increment in one step to prevent TOCTOU:
	// concurrent requests could each pass a separate `UsageCount >= Max`
	// check before any of them increments. If the AI call fails later, we
	// decrement to refund the reservation. Placed after cache/dedupe so
	// cached results and dedupe conflicts are not charged.
	newUsage, allowed := h.userStore.CheckAndIncrementUsage(user.Email)
	if !allowed {
		return c.Status(403).JSON(fiber.Map{"error": "LIMIT_EXCEEDED", "message": "Free usage limit exceeded. Please upgrade."})
	}
	h.persistUsage(user.Email, 1)

	ctx, cancel := bridgeCtx(c)
	defer cancel()
	result, lastErr := services.CallAI(ctx, resumeContent, targetJob, jobDesc, lang, moduleHints, systemPrompt)
	if lastErr != nil {
		h.userStore.DecrementUsage(user.Email) // refund reservation on failure
		h.persistUsage(user.Email, -1)
		log.Printf("[ERROR] Optimize AI call failed: %v", lastErr)
		return c.Status(503).JSON(fiber.Map{
			"error":   "AI_UNAVAILABLE",
			"message": "Service temporarily unavailable, please try again later",
		})
	}

	// Only cache when the AI result actually parses as valid resume JSON.
	if rawJSON, err := json.Marshal(result); err == nil {
		if score := services.ValidateAIResponse(string(rawJSON)); score.IsValid {
			services.SetCachedResult(resumeContent, jobDesc, targetJob, moduleHints, lang, result)
		}
	}

	return c.JSON(fiber.Map{"success": true, "data": result, "cached": false, "usage_count": newUsage, "max_free_usage": user.MaxFreeUsage})
}

func (h *ResumeHandler) Perspective(c *fiber.Ctx) error {
	user, ok := c.Locals("user").(*models.User)
	if !ok || user == nil {
		return c.Status(401).JSON(fiber.Map{"error": "UNAUTHORIZED", "message": "Authentication required"})
	}

	var body map[string]interface{}
	if err := c.BodyParser(&body); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "INVALID_BODY", "message": "Invalid request body"})
	}

	lang, _ := body["lang"].(string)
	if lang == "" {
		lang = "en"
	}
	if !services.IsValidLang(lang) {
		lang = "en"
	}

	var resumeContent string
	if rc, ok := body["resume_text"].(string); ok && rc != "" {
		resumeContent = rc
	}
	if len(resumeContent) > models.MaxResumeChars {
		resumeContent = services.TruncateUTF8(resumeContent, models.MaxResumeChars)
	}
	if resumeContent == "" {
		return c.Status(400).JSON(fiber.Map{"error": "NO_CONTENT", "message": "Resume content is required"})
	}

	targetJob, _ := body["target_job"].(string)
	jobDesc, _ := body["job_description"].(string)

	if len(targetJob) > 500 {
		targetJob = services.TruncateUTF8(targetJob, 500)
	}
	if len(jobDesc) > 2000 {
		jobDesc = services.TruncateUTF8(jobDesc, 2000)
	}

	prompt := services.GetPerspectivePrompt(lang)

	dedupeKey := services.GenerateDedupeKey(user.ID, resumeContent, jobDesc, targetJob, "", lang)
	acquired, dedupToken, dedupErr := services.TryAcquireRequest(dedupeKey)
	if dedupErr != nil {
		return c.Status(503).JSON(fiber.Map{"error": "DEDUP_UNAVAILABLE", "message": "Temporary server failure, please retry"})
	}
	if !acquired {
		return c.Status(429).JSON(fiber.Map{"error": "REQUEST_IN_PROGRESS", "message": "Same request already in progress"})
	}
	defer services.ReleaseRequest(dedupeKey, dedupToken)

	// Atomically check quota AND increment to prevent TOCTOU race on free
	// usage limit. Refund via DecrementUsage if the AI call fails.
	newUsage, allowed := h.userStore.CheckAndIncrementUsage(user.Email)
	if !allowed {
		return c.Status(403).JSON(fiber.Map{"error": "LIMIT_EXCEEDED", "message": "Free usage limit exceeded. Please upgrade."})
	}
	h.persistUsage(user.Email, 1)

	ctx, cancel := bridgeCtx(c)
	defer cancel()
	result, err := services.CallAIFromProviders(ctx, resumeContent, targetJob, jobDesc, lang, prompt)
	if err != nil {
		h.userStore.DecrementUsage(user.Email) // refund reservation on failure
		h.persistUsage(user.Email, -1)
		log.Printf("[ERROR] Perspective AI call failed: %v", err)
		return c.Status(503).JSON(fiber.Map{"error": "AI_UNAVAILABLE", "message": "AI service temporarily unavailable"})
	}

	return c.JSON(fiber.Map{"success": true, "data": result, "usage_count": newUsage, "max_free_usage": user.MaxFreeUsage})
}

func (h *ResumeHandler) GenerateResume(c *fiber.Ctx) error {
	// Allow operators to disable the AI chat feature (e.g. to control cost
	// or take it offline) by setting ENABLE_GENERATE_RESUME=false. Default
	// is enabled to preserve current production behavior. AdminHealth
	// reports the same flag so monitoring reflects the actual state.
	if os.Getenv("ENABLE_GENERATE_RESUME") == "false" {
		return c.Status(503).JSON(fiber.Map{"error": "FEATURE_DISABLED", "message": "This feature is currently disabled"})
	}
	user, ok := c.Locals("user").(*models.User)
	if !ok || user == nil {
		return c.Status(401).JSON(fiber.Map{"error": "UNAUTHORIZED", "message": "Authentication required"})
	}

	var body struct {
		Messages []models.GroqMessage `json:"messages"`
		Lang     string               `json:"lang"`
	}
	if err := c.BodyParser(&body); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "INVALID_BODY", "message": "Invalid request body"})
	}

	if len(body.Messages) == 0 {
		return c.Status(400).JSON(fiber.Map{"error": "NO_MESSAGES", "message": "Messages array is required"})
	}
	if len(body.Messages) > 50 {
		return c.Status(400).JSON(fiber.Map{"error": "TOO_MANY_MESSAGES", "message": "Too many messages (max 50)"})
	}
	// Cap each message length to prevent token-bombing the AI provider.
	for i := range body.Messages {
		if len(body.Messages[i].Content) > 5000 {
			body.Messages[i].Content = services.TruncateUTF8(body.Messages[i].Content, 5000)
		}
		// Sanitize role: only "user" and "assistant" allowed from client.
		// Prevents prompt-injection via a client-supplied "system" role.
		if body.Messages[i].Role != "user" && body.Messages[i].Role != "assistant" {
			body.Messages[i].Role = "user"
		}
	}

	if body.Lang == "" {
		body.Lang = "en"
	}
	if !services.IsValidLang(body.Lang) {
		body.Lang = "en"
	}

	prompt := services.GetGenerateResumePrompt(body.Lang)
	messages := []models.GroqMessage{{Role: "system", Content: prompt}}
	messages = append(messages, body.Messages...)

	// Dedup: prevent concurrent duplicate requests from wasting AI quota.
	lastMsg := body.Messages[len(body.Messages)-1].Content
	dedupeKey := services.GenerateDedupeKey(user.ID, lastMsg, "", "", "", body.Lang)
	acquired, dedupToken, dedupErr := services.TryAcquireRequest(dedupeKey)
	if dedupErr != nil {
		return c.Status(503).JSON(fiber.Map{"error": "DEDUP_UNAVAILABLE", "message": "Temporary server failure, please retry"})
	}
	if !acquired {
		return c.Status(429).JSON(fiber.Map{"error": "REQUEST_IN_PROGRESS", "message": "Same request already in progress"})
	}
	defer services.ReleaseRequest(dedupeKey, dedupToken)

	// Atomically check quota AND increment to prevent TOCTOU race.
	newUsage, allowed := h.userStore.CheckAndIncrementUsage(user.Email)
	if !allowed {
		return c.Status(403).JSON(fiber.Map{"error": "LIMIT_EXCEEDED", "message": "Free usage limit exceeded. Please upgrade."})
	}
	h.persistUsage(user.Email, 1)

	ctx, cancel := bridgeCtx(c)
	defer cancel()
	result, err := services.CallAIWithMessages(ctx, messages)
	if err != nil {
		h.userStore.DecrementUsage(user.Email) // refund reservation on failure
		h.persistUsage(user.Email, -1)
		log.Printf("[ERROR] GenerateResume AI call failed: %v", err)
		return c.Status(503).JSON(fiber.Map{"error": "AI_UNAVAILABLE", "message": "AI service temporarily unavailable"})
	}

	return c.JSON(fiber.Map{"success": true, "data": result, "usage_count": newUsage, "max_free_usage": user.MaxFreeUsage})
}

func GetMemUsage() uint64 {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	return m.Alloc / 1024 / 1024
}

func (h *ResumeHandler) Health(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{"status": "ok"})
}

func (h *ResumeHandler) AdminHealth(c *fiber.Ctx) error {
	providers := services.GetAIProviders()
	providerNames := make([]string, len(providers))
	providerStatus := make(map[string]bool)

	for i, p := range providers {
		providerNames[i] = p.Name
		providerStatus[p.Name] = p.APIKey != ""
	}

	// R53b-B3: actually ping the DB instead of just checking the pointer is
	// non-nil. A nil check reports "healthy" even after Close() is called or
	// the connection pool is exhausted, hiding outages from monitoring.
	var dbHealthy bool
	var dbError string
	if h.db == nil {
		dbError = "database not initialized"
	} else if err := h.db.Ping(); err != nil {
		dbError = err.Error()
	} else {
		dbHealthy = true
	}

	return c.JSON(fiber.Map{
		"status":                  "healthy",
		"timestamp":               time.Now().Format(time.RFC3339),
		"requests":                h.resumeStore.Count(),
		"version":                 "2.1.0",
		"ai":                      providerNames,
		"ai_providers":            providerStatus,
		"memory":                  fmt.Sprintf("%d MB", GetMemUsage()),
		"users":                   h.userStore.Count(),
		"cache":                   services.GetCacheStats(),
		"database":                dbHealthy,
		"database_error":          dbError,
		"generate_resume_enabled": os.Getenv("ENABLE_GENERATE_RESUME") != "false",
	})
}

func (h *ResumeHandler) OptimizeStream(c *fiber.Ctx) error {
	user, ok := c.Locals("user").(*models.User)
	if !ok || user == nil {
		return c.Status(401).JSON(fiber.Map{"error": "UNAUTHORIZED", "message": "Authentication required"})
	}

	var body map[string]interface{}
	if err := c.BodyParser(&body); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "INVALID_BODY", "message": "Invalid request body"})
	}

	lang, _ := body["lang"].(string)
	if lang == "" {
		lang = "en"
	}
	if !services.IsValidLang(lang) {
		lang = "en"
	}

	targetJob, _ := body["target_job"].(string)
	if targetJob == "" {
		targetJob, _ = body["target_position"].(string)
	}
	jobDesc, _ := body["job_description"].(string)

	// Truncate targetJob/jobDesc to the same limits as Optimize to prevent
	// oversized prompts from inflating AI cost/latency and to keep parity
	// with the non-streaming path (Optimize truncates at 500/2000).
	if len(targetJob) > 500 {
		targetJob = services.TruncateUTF8(targetJob, 500)
	}
	if len(jobDesc) > 2000 {
		jobDesc = services.TruncateUTF8(jobDesc, 2000)
	}

	var resumeContent string
	if rc, ok := body["resume_text"].(string); ok && rc != "" {
		resumeContent = rc
	} else if rc, ok := body["resume"].(string); ok && rc != "" {
		resumeContent = rc
	} else if rc, ok := body["resume_content"].(string); ok && rc != "" {
		// R30-H5: handle string resume_content directly — json.Marshal on a
		// string produces "\"text\"" (with literal quotes), double-encoding
		// the content and degrading AI output quality.
		resumeContent = rc
	} else if rc, mErr := json.Marshal(body["resume_content"]); mErr != nil {
		log.Printf("[ERROR] json.Marshal resume_content failed: %v", mErr)
	} else if rc != nil && string(rc) != "null" {
		resumeContent = string(rc)
	}

	if len(resumeContent) > models.MaxResumeChars {
		resumeContent = services.TruncateUTF8(resumeContent, models.MaxResumeChars)
	}
	if resumeContent == "" {
		return c.Status(400).JSON(fiber.Map{"error": "NO_CONTENT", "message": "Resume content is required"})
	}

	// Build moduleHints to match Optimize's cache key. Without this, the
	// cache check below passes "" as moduleHints and never matches entries
	// written by Optimize (which uses real moduleHints), making the cache
	// check dead code.
	modules, _ := body["modules"].([]interface{})
	if len(modules) == 0 {
		modules = []interface{}{"ats", "star", "quant", "summary", "format"}
	}
	// R51-B6: cap modules slice — only 5 known module names are matched by
	// BuildModuleHints, but a malicious client could send thousands of
	// entries to trigger large map allocation. Matches GenerateResume's
	// len > 50 guard on messages.
	if len(modules) > 20 {
		modules = modules[:20]
	}
	moduleHints := services.BuildModuleHints(modules)

	// Cache check: if Optimize (non-stream) already produced a result for
	// this input, return it directly instead of starting a new SSE stream.
	// Without this, switching between /optimize and /optimize-stream (or
	// repeating /optimize-stream) always consumes an AI call + quota.
	if cached, ok := services.GetCachedResult(resumeContent, jobDesc, targetJob, moduleHints, lang); ok {
		return c.JSON(fiber.Map{"success": true, "data": cached, "cached": true, "usage_count": user.UsageCount, "max_free_usage": user.MaxFreeUsage})
	}

	// Dedupe: prevent concurrent identical stream requests from each
	// consuming an upstream AI call (M4 — Optimize had this, OptimizeStream
	// was missing it). R27-L9: include moduleHints so requests with different
	// module selections are not incorrectly deduped.
	dedupeKey := services.GenerateDedupeKey(user.ID, resumeContent, jobDesc, targetJob, moduleHints, lang)
	acquired, dedupToken, dedupErr := services.TryAcquireRequest(dedupeKey)
	if dedupErr != nil {
		return c.Status(503).JSON(fiber.Map{"error": "DEDUP_UNAVAILABLE", "message": "Temporary server failure, please retry"})
	}
	if !acquired {
		return c.Status(429).JSON(fiber.Map{"error": "REQUEST_IN_PROGRESS", "message": "Same request already in progress"})
	}
	defer services.ReleaseRequest(dedupeKey, dedupToken)

	// Atomically check quota AND increment to prevent TOCTOU race.
	// Refund via DecrementUsage if the stream/fallback fails.
	newUsage, allowed := h.userStore.CheckAndIncrementUsage(user.Email)
	if !allowed {
		return c.Status(403).JSON(fiber.Map{"error": "LIMIT_EXCEEDED", "message": "Free usage limit exceeded. Please upgrade."})
	}
	h.persistUsage(user.Email, 1)

	systemPrompt := services.GetOptimizePrompt(lang)

	// R57b-B1: append injection defense for the streaming path (CallAI and
	// CallAIFromProviders do this internally, but OptimizeStream constructs
	// messages directly).
	systemPrompt += services.InjectionDefenseSuffix

	userMsg := services.BuildUserMsg(lang, targetJob, jobDesc, resumeContent)
	// R51-B1: append moduleHints to match Optimize (non-stream) path.
	// Without this, streaming AI calls ignore the user's module selection
	// (ATS/STAR/etc.), and the result is cached under a key that includes
	// moduleHints — polluting the cache for subsequent non-stream calls.
	if moduleHints != "" {
		userMsg += "\n\nOptimization focus:\n" + moduleHints
	}

	messages := []models.GroqMessage{
		{Role: "system", Content: systemPrompt},
		{Role: "user", Content: userMsg},
	}

	provider, err := services.GetFirstProvider()
	if err != nil {
		// Refund the pre-incremented quota — no AI call was made.
		h.userStore.DecrementUsage(user.Email)
		h.persistUsage(user.Email, -1)
		return c.Status(503).JSON(fiber.Map{"error": "NO_PROVIDER", "message": "No AI provider available"})
	}

	// Bridge fasthttp's connection-level context to a standard context so
	// CallAIStreamWithContext's internal goroutines (<-parent.Done()) actually
	// fire when the client disconnects. Fiber's c.UserContext() returns
	// context.Background() unless explicitly set, which never cancels and
	// leaks 2 goroutines + 1 HTTP conn to the AI provider per aborted SSE.
	// R38b-L3: reuse bridgeCtx instead of duplicating its logic here.
	// The inline version was functionally identical but required syncing
	// changes in two places if bridgeCtx ever needs updating.
	streamCtx, streamCancel := bridgeCtx(c)
	defer streamCancel()
	stream, errCh, err := services.CallAIStreamWithContext(streamCtx, provider, messages)
	if err != nil {
		log.Printf("[Stream] Stream failed, falling back to sync: %v", err)
		result, syncErr := services.CallAI(streamCtx, resumeContent, targetJob, jobDesc, lang, moduleHints, systemPrompt)
		if syncErr != nil {
			h.userStore.DecrementUsage(user.Email) // refund reservation
			h.persistUsage(user.Email, -1)
			log.Printf("[ERROR] OptimizeStream sync fallback failed: %v", syncErr)
			return c.Status(503).JSON(fiber.Map{"error": "AI_UNAVAILABLE", "message": "AI service temporarily unavailable"})
		}
		// Sync fallback succeeded — usage already pre-charged above.
		// Cache the result so subsequent identical requests (stream or
		// non-stream) hit the cache instead of consuming another AI call.
		// Only cache if the AI output is valid (matches Optimize's behavior).
		if rawJSON, err := json.Marshal(result); err == nil {
			if score := services.ValidateAIResponse(string(rawJSON)); score.IsValid {
				services.SetCachedResult(resumeContent, jobDesc, targetJob, moduleHints, lang, result)
			}
		}
		return c.JSON(fiber.Map{"success": true, "data": result, "fallback": true, "usage_count": newUsage, "max_free_usage": user.MaxFreeUsage})
	}

	c.Set("Content-Type", "text/event-stream")
	c.Set("Cache-Control", "no-cache")
	c.Set("Connection", "keep-alive")
	c.Set("X-Accel-Buffering", "no")

	c.Context().SetBodyStream(nil, -1)

	streamDone := false
	ctx := c.Context()
	var streamResult strings.Builder
	for chunk := range stream {
		if ctx.Err() != nil {
			break
		}
		// R45-B1: SSE spec requires each line of a data field to be prefixed
		// with "data: ". AI chunks may contain \n (e.g. formatted JSON, multi-
		// line strings). Without escaping, content after \n is treated as a
		// new event field and silently dropped by the client parser, producing
		// truncated/corrupted JSON.
		// R48-B3: also strip bare \r — SSE spec treats \r, \r\n, and \n as
		// line terminators. After replacing \n, a bare \r (e.g. from Windows
		// CRLF where \r survives after \n is escaped) would still be
		// interpreted as a field delimiter, fragmenting the event frame.
		escaped := strings.ReplaceAll(chunk, "\r", "")
		escaped = strings.ReplaceAll(escaped, "\n", "\ndata: ")
		if _, err := c.Write([]byte("data: " + escaped + "\n\n")); err != nil {
			break
		}
		streamDone = true
		streamResult.WriteString(chunk)
	}

	// Drain the error channel to learn whether the producer exited cleanly
	// (saw upstream [DONE]) or was cut short by a read error. A nil error
	// means the upstream stream completed normally; a non-nil error means
	// the response was truncated and we must NOT send [DONE] to the client
	// (otherwise the client treats a truncated stream as complete).
	streamErr := <-errCh
	cleanClose := streamErr == nil && ctx.Err() == nil

	if cleanClose {
		if _, err := c.Write([]byte("data: [DONE]\n\n")); err != nil {
			log.Printf("[Stream] failed to write [DONE]: %v", err)
		}
		// Cache the streamed result so subsequent identical requests
		// hit the cache instead of consuming another AI call. Only cache
		// if the AI output is valid (matches Optimize's behavior).
		fullResult := streamResult.String()
		if score := services.ValidateAIResponse(fullResult); score.IsValid {
			// R50-H1: cache as parsed map (not raw string) to match Optimize's
			// cache type. Without this, a /optimize-stream request caches a
			// string, and a subsequent /optimize request hitting the same cache
			// returns "data": "<raw JSON string>" instead of an object, causing
			// the frontend to access data.optimized_content on a string → crash.
			if parsed := services.ParseJSONContent(fullResult); parsed != nil {
				services.SetCachedResult(resumeContent, jobDesc, targetJob, moduleHints, lang, parsed)
			}
		}
		// R40-L2: log successful stream completion for monitoring
		// success rate and latency distribution of AI streaming calls.
		log.Printf("[Stream] completed successfully for user %s", hashEmail(user.Email))
	} else if streamErr != nil {
		log.Printf("[Stream] producer ended with error: %v", streamErr)
		if _, err := c.Write([]byte("data: {\"error\":\"stream_interrupted\"}\n\n")); err != nil {
			log.Printf("[Stream] failed to write error event: %v", err)
		}
	}

	// Usage was pre-charged above. Refund ONLY if no content was delivered
	// to the client (streamDone=false). The prior condition required both
	// streamDone AND cleanClose, which meant a client could read all AI
	// chunks (streamDone=true) then deliberately disconnect before [DONE]
	// to force cleanClose=false and get a refund — bypassing the free
	// quota limit. Once content has been delivered, the AI call's cost
	// is incurred and the quota charge should stand.
	if !streamDone {
		h.userStore.DecrementUsage(user.Email)
		h.persistUsage(user.Email, -1)
	}

	return nil
}
