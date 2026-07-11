package middleware

import (
	"fmt"
	"log"
	"os"
	"regexp"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/getsentry/sentry-go"
)

// R50-B5: regex patterns for PII that may leak into Sentry via third-party
// error messages (SQLite driver, AI provider SDKs, etc.). BeforeSend
// scrubs these before the event is sent to Sentry.
var (
	scrubEmailRe   = regexp.MustCompile(`[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}`)
	scrubBearerRe  = regexp.MustCompile(`(?i)bearer\s+[a-zA-Z0-9\-_.]+`)
	scrubTokenRe   = regexp.MustCompile(`(?i)(token|password|secret|api[_-]?key)["'\s:=]+[a-zA-Z0-9\-_.]{8,}`)
)

func scrubPII(s string) string {
	s = scrubEmailRe.ReplaceAllString(s, "[email]")
	s = scrubBearerRe.ReplaceAllString(s, "Bearer [redacted]")
	s = scrubTokenRe.ReplaceAllString(s, "${1}=[redacted]")
	return s
}

func scrubSentryEvent(event *sentry.Event, hint *sentry.EventHint) *sentry.Event {
	if event == nil {
		return nil
	}
	if event.Message != "" {
		event.Message = scrubPII(event.Message)
	}
	for i, ex := range event.Exception {
		if ex.Value != "" {
			event.Exception[i].Value = scrubPII(ex.Value)
		}
	}
	return event
}

func InitSentry() {
	dsn := os.Getenv("SENTRY_DSN")
	if dsn == "" {
		log.Println("[Sentry] SENTRY_DSN not set, Sentry disabled")
		return
	}

	err := sentry.Init(sentry.ClientOptions{
		Dsn:              dsn,
		TracesSampleRate: 0.1,
		EnableTracing:    true,
		// R50-B5: filter PII (email, bearer tokens, passwords) from events
		// before they reach Sentry. Third-party libs (SQLite driver, AI
		// provider SDKs) may include raw user input in error messages.
		BeforeSend: scrubSentryEvent,
	})
	if err != nil {
		log.Printf("[Sentry] Failed to initialize: %v", err)
		return
	}

	log.Println("[Sentry] Initialized successfully")
}

// FlushSentry flushes any buffered events to Sentry before the process exits.
// Must be called during graceful shutdown to avoid losing captured errors.
func FlushSentry() {
	if os.Getenv("SENTRY_DSN") != "" {
		sentry.Flush(2 * time.Second)
	}
}

func Sentry() fiber.Handler {
	return func(c *fiber.Ctx) error {
		hub := sentry.GetHubFromContext(c.UserContext())
		if hub == nil {
			hub = sentry.CurrentHub().Clone()
			// Attach the cloned hub to the request context so downstream
			// handlers that call sentry.GetHubFromContext(ctx) receive the
			// per-request hub with the configured scope (method/path/ip).
			// Without this, GetHubFromContext returns nil outside this
			// middleware, falling back to the global hub — losing
			// request-scoped tags.
			c.SetUserContext(sentry.SetHubOnContext(c.UserContext(), hub))
		}

		hub.ConfigureScope(func(scope *sentry.Scope) {
			// R49-B2: method + normalized path as tags (bounded cardinality),
			// but raw path / IP / User-Agent as Context (high cardinality —
			// would explode Sentry's tag index and degrade query perf).
			// normalizedPath mirrors the logic in structured_logger.go /
			// metrics.go — uses the route template (e.g. /api/v1/resumes/:id)
			// instead of the raw path with dynamic UUIDs.
			scope.SetTag("method", c.Method())
			scope.SetTag("path", normalizedPath(c))
			// R45-B2: attach request_id so Sentry events can be correlated
			// with structured access logs (StructuredLog logs request_id).
			// requestid middleware runs before Sentry, so c.Locals is set.
			if rid, ok := c.Locals("requestid").(string); ok && rid != "" {
				scope.SetTag("request_id", rid)
			}
			scope.SetContext("request", sentry.Context{
				"method":      c.Method(),
				"path":        c.Path(),
				"ip":          c.IP(),
				"user_agent":  c.Get("User-Agent"),
			})
		})

		defer func() {
			if r := recover(); r != nil {
				err, ok := r.(error)
				if !ok {
					err = fmt.Errorf("%v", r)
				}
				hub.CaptureException(err)
				// If the response has already started streaming (e.g. SSE
				// OptimizeStream calls SetBodyStream before iterating AI
				// chunks), writing a JSON error body would corrupt the
				// stream — the client would receive mixed SSE data + JSON.
				// In that case, just let the connection close; the panic
				// is already captured in Sentry.
				if c.Response().IsBodyStream() || len(c.Response().Body()) > 0 {
					return
				}
				_ = c.Status(500).JSON(fiber.Map{"error": "INTERNAL_ERROR", "message": "Internal server error"})
			}
		}()

		return c.Next()
	}
}
