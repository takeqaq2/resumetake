package middleware

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"

	"resumetake/models"
)

type StructuredLogger struct {
	logger *log.Logger
}

func NewStructuredLogger() *StructuredLogger {
	return &StructuredLogger{
		logger: log.New(os.Stdout, "", 0),
	}
}

type LogEntry struct {
	Timestamp  string  `json:"timestamp"`
	RequestID  string  `json:"request_id"`
	Method     string  `json:"method"`
	Path       string  `json:"path"`
	Status     int     `json:"status"`
	Latency    float64 `json:"latency_ms"`
	IP         string  `json:"ip"`
	UserAgent  string  `json:"user_agent"`
	UserID     string  `json:"user_id,omitempty"`
	Error      string  `json:"error,omitempty"`
}

// sanitizeLogField strips CR/LF from log fields to prevent log injection
// (CWE-117). Error messages from JSON parsers, database drivers, or
// third-party libraries can contain newlines; if they incorporate
// user-controlled input (e.g. a malformed resume title), an attacker
// can inject \n to forge fake log lines that hide malicious activity.
func sanitizeLogField(s string) string {
	s = strings.ReplaceAll(s, "\n", "\\n")
	s = strings.ReplaceAll(s, "\r", "\\r")
	return s
}

// R56b-B1: emit JSON instead of space-separated text. The prior format
// split fields on spaces, but User-Agent and error messages routinely
// contain spaces — making the log line unparseable by log aggregators
// (Loki, Datadog) that rely on field boundaries. JSON is self-delimiting
// and survives arbitrary content in any field. sanitizeLogField is still
// applied so CR/LF cannot inject fake log lines (CWE-117) — json.Marshal
// would escape them anyway, but defense-in-depth keeps the guard explicit.
func (s *StructuredLogger) Log(entry LogEntry) {
	entry.Path = sanitizeLogField(entry.Path)
	entry.IP = sanitizeLogField(entry.IP)
	entry.UserAgent = sanitizeLogField(entry.UserAgent)
	entry.Error = sanitizeLogField(entry.Error)
	b, err := json.Marshal(entry)
	if err != nil {
		// Marshal should never fail for this struct (all fields are
		// strings/numbers), but fall back to a safe text line so we
		// never drop an access-log entry due to a serialization bug.
		s.logger.Printf(`{"timestamp":%q,"error":"log marshal failed: %v"}`, entry.Timestamp, err)
		return
	}
	s.logger.Println(string(b))
}

// normalizedPath returns the route template (e.g. "/api/v1/resumes/:id")
// instead of the raw path (which contains dynamic UUIDs). This keeps log
// cardinality bounded — without it, every resume ID becomes a unique path
// in access logs, which both wastes space and leaks resource IDs to any
// log aggregator. Mirrors the same normalization in metrics.go.
func normalizedPath(c *fiber.Ctx) string {
	if r := c.Route(); r.Path != "" {
		return r.Path
	}
	return "UNMATCHED"
}

func StructuredLog() fiber.Handler {
	logger := NewStructuredLogger()

	return func(c *fiber.Ctx) error {
		start := time.Now()

		// R49-B3: defer the log write so panics in downstream handlers are
		// still recorded in the access log. Without defer, a panic would
		// skip the entire log block — the request would vanish from logs,
		// making it impossible to correlate Sentry alerts with access
		// patterns. The Sentry middleware (registered earlier) recovers
		// the panic and writes a 500, so c.Response().StatusCode() will
		// be 500 by the time this deferred function runs.
		var handlerErr error
		defer func() {
			latency := time.Since(start).Milliseconds()

			userID := ""
			if user, ok := c.Locals("user").(*models.User); ok && user != nil {
				userID = user.ID
			}

			rid, _ := c.Locals("requestid").(string)
			if rid == "" {
				rid = c.GetRespHeader("X-Request-ID")
			}

			// R39b-M1: force 500 on panic — Sentry's defer runs after this
			// one (LIFO), so c.Response().StatusCode() is still 200 here.
			status := c.Response().StatusCode()
			if r := recover(); r != nil {
				status = 500
				logger.Log(LogEntry{
					Timestamp: time.Now().Format(time.RFC3339),
					RequestID: rid,
					Method:    c.Method(),
					Path:      normalizedPath(c),
					Status:    status,
					Latency:   float64(latency),
					IP:        c.IP(),
					UserAgent: c.Get("User-Agent"),
					UserID:    userID,
					Error:     fmt.Sprintf("PANIC: %v", r),
				})
				panic(r) // re-panic so Sentry middleware can capture it
			}

			entry := LogEntry{
				Timestamp: time.Now().Format(time.RFC3339),
				RequestID: rid,
				Method:    c.Method(),
				Path:      normalizedPath(c),
				Status:    status,
				Latency:   float64(latency),
				IP:        c.IP(),
				UserAgent: c.Get("User-Agent"),
				UserID:    userID,
			}

			if handlerErr != nil {
				entry.Error = handlerErr.Error()
			}

			logger.Log(entry)
		}()

		handlerErr = c.Next()
		return handlerErr
	}
}
