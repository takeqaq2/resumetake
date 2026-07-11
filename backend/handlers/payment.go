package handlers

import (
	cryptorand "crypto/rand"
	"fmt"
	"log"
	"math"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"

	"resumetake/models"
	"resumetake/services"
)

type PaymentHandler struct {
	userStore   *models.UserStore
	persistence UserPersistence
	seenEvents  *models.SeenEventsStore
}

func NewPaymentHandler(userStore *models.UserStore, persistence UserPersistence) *PaymentHandler {
	return &PaymentHandler{
		userStore:   userStore,
		persistence: persistence,
		seenEvents:  models.NewSeenEventsStore(),
	}
}

// Close releases resources held by the handler (background goroutines).
func (h *PaymentHandler) Close() { h.seenEvents.Close() }

// planRank returns a comparable rank for plan names so we can reject paid
// downgrades (e.g. an Enterprise user buying Pro should not be downgraded).
// Annual plans rank higher than their monthly counterparts so that a monthly
// subscriber can upgrade to the annual tier (pro→pro_annual) without being
// blocked by the "already owned" guard.
func planRank(plan string) int {
	switch plan {
	case "enterprise_annual":
		return 4
	case "enterprise":
		return 3
	case "pro_annual":
		return 2
	case "pro":
		return 1
	default:
		return 0
	}
}

// normalizePlan maps annual plan IDs to their base plan for storage. We store
// "pro"/"enterprise" regardless of billing cycle; the billing cycle is
// reflected in the payment record, not the user's plan field.
func normalizePlan(plan string) string {
	switch plan {
	case "pro_annual":
		return "pro"
	case "enterprise_annual":
		return "enterprise"
	default:
		return plan
	}
}

// parseToCents converts a decimal string like "9.99" or "9.9900" into
// integer cents (999), avoiding float64 rounding errors in monetary
// comparisons. Returns an error if the string is not a valid decimal.
func parseToCents(s string) (int64, error) {
	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0, err
	}
	return int64(math.Round(f * 100)), nil
}

func (h *PaymentHandler) GetPricing(c *fiber.Ctx) error {
	tiers := []fiber.Map{
		{
			"id":    "free",
			"name":  "Free",
			"price": 0,
			"features": []string{
				"5 AI optimizations per month",
				"Basic resume templates",
				"1 language support",
				"Download as text",
			},
			"usage_limit": 5,
		},
		{
			"id":           "pro",
			"name":         "Pro",
			"price":        9.99,
			"annual_price": 79.99,
			"annual_save":  40,
			"features": []string{
				"Unlimited AI optimizations",
				"All premium templates",
				"10+ language support",
				"PDF export",
				"ATS score analysis",
				"Perspective analysis",
				"Priority support",
			},
			"usage_limit": -1,
		},
		{
			"id":           "enterprise",
			"name":         "Enterprise",
			"price":        49.99,
			"annual_price": 399.99,
			"annual_save":  200,
			"features": []string{
				"Everything in Pro",
				"Team collaboration",
				"Custom templates",
				"API access",
				"Batch resume processing",
				"Dedicated support",
				"SLA guarantee",
			},
			"usage_limit": -1,
		},
	}

	return c.JSON(fiber.Map{"success": true, "data": tiers})
}

func (h *PaymentHandler) GetTemplates(c *fiber.Ctx) error {
	lang := c.Query("lang", "en")
	templateData := map[string][]fiber.Map{
		"zh": {
			{"id": "professional", "name": "专业商务", "description": "适合传统行业和商务岗位"},
			{"id": "modern", "name": "现代简约", "description": "适合互联网和科技行业"},
			{"id": "creative", "name": "创意设计", "description": "适合设计和创意岗位"},
			{"id": "academic", "name": "学术科研", "description": "适合教育和研究岗位"},
			{"id": "executive", "name": "高管专用", "description": "适合高级管理岗位"},
			{"id": "minimal", "name": "极简风格", "description": "简洁大方，通用性强"},
		},
		"en": {
			{"id": "professional", "name": "Professional", "description": "For traditional and business roles"},
			{"id": "modern", "name": "Modern", "description": "For tech and startup roles"},
			{"id": "creative", "name": "Creative", "description": "For design and creative roles"},
			{"id": "academic", "name": "Academic", "description": "For education and research roles"},
			{"id": "executive", "name": "Executive", "description": "For senior management roles"},
			{"id": "minimal", "name": "Minimal", "description": "Clean and versatile"},
		},
	}

	data, ok := templateData[lang]
	if !ok {
		data = templateData["en"]
	}

	return c.JSON(fiber.Map{"success": true, "data": data})
}

func (h *PaymentHandler) CreatePayPalOrder(c *fiber.Ctx) error {
	paypalClientID := strings.TrimSpace(os.Getenv("PAYPAL_CLIENT_ID"))
	if paypalClientID == "" {
		return c.Status(503).JSON(fiber.Map{"error": "PAYMENT_NOT_CONFIGURED", "message": "Payment system not configured"})
	}

	var body struct {
		Plan string `json:"plan"`
		Lang string `json:"lang"`
	}
	if err := c.BodyParser(&body); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "INVALID_BODY", "message": "Invalid request body"})
	}

	// Whitelist body.Lang before interpolating it into return_url/cancel_url.
	// Without this, a client could pass Lang="../../admin" or other path
	// segments that would be injected into the PayPal redirect URLs.
	if !services.IsValidLang(body.Lang) {
		body.Lang = "en"
	}

	user, ok := c.Locals("user").(*models.User)
	if !ok || user == nil {
		return c.Status(401).JSON(fiber.Map{"error": "UNAUTHORIZED", "message": "Authentication required"})
	}

	priceAmounts := map[string]string{
		"pro":               "9.99",
		"pro_annual":        "79.99",
		"enterprise":        "49.99",
		"enterprise_annual": "399.99",
	}

	priceNames := map[string]string{
		"pro":               "ResumeTake Pro - Monthly",
		"pro_annual":        "ResumeTake Pro - Annual",
		"enterprise":        "ResumeTake Enterprise - Monthly",
		"enterprise_annual": "ResumeTake Enterprise - Annual",
	}

	amount, ok := priceAmounts[body.Plan]
	if !ok {
		return c.Status(400).JSON(fiber.Map{"error": "INVALID_PLAN", "message": "Invalid plan"})
	}

	// Reject duplicate purchases — a user who already has this plan (or a
	// higher one) shouldn't be able to pay again for the same tier. Without
	// this, CapturePayPalOrder's downgrade guard silently no-ops the plan
	// update while the user's payment still goes through, leading to refund
	// disputes.
	if planRank(body.Plan) <= planRank(user.Plan) {
		return c.Status(409).JSON(fiber.Map{"error": "ALREADY_OWNED", "message": "You already have this plan or a higher one"})
	}

	accessToken, err := services.GetPayPalAccessToken(c.UserContext())
	if err != nil {
		return c.Status(502).JSON(fiber.Map{"error": "PAYPAL_ERROR", "message": "Failed to connect to PayPal"})
	}

	uid := user.ID
	// R31-7: use UnixNano + a random suffix so two orders created in the same
	// second don't collide on orderID. The previous Unix()-based ID caused
	// duplicate reference_id values for concurrent purchases, breaking
	// ownership verification in CapturePayPalOrder.
	var rnd [4]byte
	if _, err := cryptorand.Read(rnd[:]); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "INTERNAL_ERROR", "message": "Failed to generate order ID"})
	}
	orderID := fmt.Sprintf("rt_%s_%d_%x", uid, time.Now().UnixNano(), rnd[:])
	baseURL := os.Getenv("PUBLIC_BASE_URL")
	if baseURL == "" {
		baseURL = "https://resume.takee.top"
	}
	returnURL := fmt.Sprintf("%s/%s/pricing?payment=success&order_id=%s", baseURL, body.Lang, orderID)
	cancelURL := fmt.Sprintf("%s/%s/pricing?payment=cancelled", baseURL, body.Lang)

	paypalOrderID, err := services.CreatePayPalOrder(c.UserContext(), accessToken, amount, "USD", priceNames[body.Plan], orderID, returnURL, cancelURL)
	if err != nil {
		log.Printf("[ERROR] CreatePayPalOrder failed: %v", err)
		return c.Status(502).JSON(fiber.Map{"error": "PAYPAL_ERROR", "message": "Failed to create PayPal order"})
	}

	return c.JSON(fiber.Map{
		"success":   true,
		"order_id":  paypalOrderID,
		"client_id": paypalClientID,
		"plan":      body.Plan,
		"amount":    amount,
		"currency":  "USD",
	})
}

func (h *PaymentHandler) CapturePayPalOrder(c *fiber.Ctx) error {
	var body struct {
		OrderID string `json:"order_id"`
		Plan    string `json:"plan"`
	}
	if err := c.BodyParser(&body); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "INVALID_BODY", "message": "Invalid request body"})
	}

	if body.OrderID == "" {
		return c.Status(400).JSON(fiber.Map{"error": "MISSING_ORDER_ID", "message": "Order ID is required"})
	}

	// Validate OrderID format to prevent path manipulation in PayPal API URL
	// construction. PayPal order IDs are alphanumeric (e.g. "O-XXXXXXXXX" or
	// 17-char EC-XXXX...). Reject anything containing path/query separators.
	// R54b-B1: allow lowercase letters — CapturePayPalOrder's own validation
	// accepts [a-zA-Z0-9-_], and PayPal may return mixed-case IDs. The
	// uppercase-only check here could reject valid orders from capture.
	for _, ch := range body.OrderID {
		if !((ch >= 'A' && ch <= 'Z') || (ch >= 'a' && ch <= 'z') || (ch >= '0' && ch <= '9') || ch == '-' || ch == '_') {
			return c.Status(400).JSON(fiber.Map{"error": "INVALID_ORDER_ID", "message": "Invalid order ID format"})
		}
	}

	// Verify the authenticated user BEFORE calling any PayPal API. If the
	// AuthMiddleware is misconfigured or removed, we must not capture a
	// payment that cannot be attributed to a user (would leave the user
	// having paid but not upgraded).
	user, ok := c.Locals("user").(*models.User)
	if !ok || user == nil {
		return c.Status(401).JSON(fiber.Map{"error": "UNAUTHORIZED", "message": "Authentication required"})
	}

	// Validate the requested plan against the server-side price map; ignore any
	// client-supplied plan that doesn't match a known tier.
	expectedAmount, ok := map[string]string{
		"pro":               "9.99",
		"pro_annual":        "79.99",
		"enterprise":        "49.99",
		"enterprise_annual": "399.99",
	}[body.Plan]
	if !ok {
		return c.Status(400).JSON(fiber.Map{"error": "INVALID_PLAN", "message": "Invalid plan"})
	}

	accessToken, err := services.GetPayPalAccessToken(c.UserContext())
	if err != nil {
		return c.Status(502).JSON(fiber.Map{"error": "PAYPAL_ERROR", "message": "Failed to connect to PayPal"})
	}

	result, err := services.CapturePayPalOrder(c.UserContext(), accessToken, body.OrderID)
	if err != nil {
		log.Printf("[ERROR] CapturePayPalOrder failed: %v", err)
		return c.Status(502).JSON(fiber.Map{"error": "PAYPAL_ERROR", "message": "Failed to capture PayPal order"})
	}

	status, _ := result["status"].(string)
	if status != "COMPLETED" {
		return c.Status(400).JSON(fiber.Map{"error": "PAYMENT_INCOMPLETE", "message": "Payment not completed"})
	}

	// Verify the captured amount matches the expected price for the plan to
	// prevent clients from upgrading after paying for a cheaper tier.
	// Fail-closed: if we cannot extract the captured amount from the PayPal
	// response (malformed/unexpected structure), reject rather than skip the
	// check — the signature alone does not guarantee the amount is correct.
	capturedAmount, capturedCurrency := extractCapturedAmount(result)
	if capturedAmount == "" {
		log.Printf("[ERROR] CapturePayPalOrder: could not extract captured amount from PayPal response for order %s", body.OrderID)
		return c.Status(400).JSON(fiber.Map{"error": "AMOUNT_UNVERIFIABLE", "message": "Could not verify payment amount"})
	}
	if capturedCurrency != "" && capturedCurrency != "USD" {
		log.Printf("[ERROR] CapturePayPalOrder: currency mismatch for order %s: got %s, expected USD", body.OrderID, capturedCurrency)
		return c.Status(400).JSON(fiber.Map{"error": "CURRENCY_MISMATCH", "message": "Currency mismatch"})
	}
	// Compare amounts in integer cents to avoid float64 rounding issues
	// with monetary values. PayPal may return "9.990" or "9.9900" for
	// "9.99" — parsing to cents normalizes these correctly.
	capturedCents, err1 := parseToCents(capturedAmount)
	expectedCents, err2 := parseToCents(expectedAmount)
	if err1 != nil || err2 != nil || capturedCents != expectedCents {
		return c.Status(400).JSON(fiber.Map{
			"error":   "AMOUNT_MISMATCH",
			"message": "Captured amount does not match plan price",
		})
	}

	// Verify the captured order belongs to the current user. The reference_id
	// was set to "rt_{uid}_{timestamp}" during CreatePayPalOrder (R27-M6:
	// full UUID, not truncated); reject if it doesn't match the current user's
	// ID (prevents using another user's captured order to upgrade your own
	// account).
	// Fail-closed: if reference_id is missing/unextractable, we cannot prove
	// ownership, so reject.
	refID := extractReferenceID(result)
	if refID == "" {
		log.Printf("[ERROR] CapturePayPalOrder: could not extract reference_id from PayPal response for order %s", body.OrderID)
		return c.Status(403).JSON(fiber.Map{"error": "ORDER_OWNERSHIP_UNVERIFIABLE", "message": "Could not verify order ownership"})
	}
	uidPrefix := user.ID
	// R53-B4: exact uid match instead of prefix — prevents prefix collision
	// if user ID format ever changes from fixed-length UUIDs. Splits
	// "rt_{uid}_{rest}" by "_" and compares the uid segment exactly.
	refParts := strings.SplitN(refID, "_", 3) // ["rt", uid, rest]
	if len(refParts) < 2 || refParts[1] != uidPrefix {
		return c.Status(403).JSON(fiber.Map{"error": "ORDER_OWNERSHIP_MISMATCH", "message": "This order does not belong to your account"})
	}

	plan := normalizePlan(body.Plan)
	orderID := body.OrderID
	// R33-H1: extract the PayPal capture ID so refund webhooks can look up
	// the user. For REFUND events, parent_id is the capture ID (not the
	// order ID stored in SubscriptionID), so without storing CaptureID the
	// refund webhook can never find the user to downgrade them.
	captureID := extractCaptureID(result)
	// R55b-B3: log when captureID extraction fails — a malformed PayPal
	// response leaves CaptureID empty, which prevents refund webhooks from
	// finding the user (GetByCaptureID won't match ""). The capture still
	// succeeds (amount verified, plan upgraded), so we don't fail-closed,
	// but ops needs visibility to manually record the capture ID.
	if captureID == "" {
		log.Printf("[WARN] CapturePayPalOrder: extractCaptureID returned empty for order %s — refund webhook lookup will fail", orderID)
	}
	// Atomically apply the plan upgrade under the user store write lock to
	// avoid racing with concurrent Optimize usage increments on the same user.
	// Guard against paid downgrade: only apply the new plan if its rank is
	// >= the current plan's rank (e.g. pro→enterprise ok, enterprise→pro
	// rejected). This mirrors the webhook's "never downgrade" guard.
	updatedUser, ok := h.userStore.UpdateUser(user.Email, func(u *models.User) {
		// Always store SubscriptionID and CaptureID so the webhook can
		// find the user for self-healing even if the downgrade guard
		// rejects the plan change. Without this, a concurrent upgrade
		// causes the webhook to loop forever on retries.
		u.SubscriptionID = orderID
		u.CaptureID = captureID
		if planRank(plan) >= planRank(u.Plan) {
			u.Plan = plan
			u.MaxFreeUsage = -1
		}
	})
	if !ok {
		return c.Status(404).JSON(fiber.Map{"error": "USER_NOT_FOUND", "message": "User no longer exists"})
	}
	if h.persistence != nil {
		// R37-B1: targeted UPDATE of plan fields only, avoiding SaveUser's
		// INSERT OR REPLACE which clobbers a concurrent persistUsage
		// increment of usage_count.
		if err := h.persistence.UpdateUserPlan(updatedUser.Email, updatedUser.Plan, orderID, captureID, updatedUser.MaxFreeUsage); err != nil {
			log.Printf("[ERROR] Failed to persist user plan upgrade: %v", err)
		}
	}

	// R40-L1: audit log for successful payment capture — critical for
	// financial reconciliation and dispute resolution. Without this there
	// is no server-side record of who paid for what and when.
	log.Printf("[PAYMENT] User %s upgraded to %s, order=%s, capture=%s, amount=%s %s",
		hashEmail(user.Email), plan, orderID, captureID, capturedAmount, capturedCurrency)

	// R30-L8: return the normalized plan (e.g. "pro" not "pro_annual") so
	// the frontend's view of the user's plan matches the stored value.
	return c.JSON(fiber.Map{"success": true, "plan": normalizePlan(body.Plan)})
}

// extractCapturedAmount reads the gross amount value from a PayPal capture response.
func extractCapturedAmount(result map[string]interface{}) (value, currency string) {
	units, _ := result["purchase_units"].([]interface{})
	if len(units) == 0 {
		return "", ""
	}
	unit, _ := units[0].(map[string]interface{})
	if unit == nil {
		return "", ""
	}
	payments, _ := unit["payments"].(map[string]interface{})
	if payments == nil {
		return "", ""
	}
	captures, _ := payments["captures"].([]interface{})
	if len(captures) == 0 {
		return "", ""
	}
	capture, _ := captures[0].(map[string]interface{})
	if capture == nil {
		return "", ""
	}
	amount, _ := capture["amount"].(map[string]interface{})
	if amount == nil {
		return "", ""
	}
	value, _ = amount["value"].(string)
	currency, _ = amount["currency_code"].(string)
	return value, currency
}

// extractReferenceID reads the reference_id from a PayPal capture response.
// This is the custom ID we set during CreatePayPalOrder ("rt_{uid}_{ts}",
// full UUID — R27-M6 removed the [:8] truncation), used to verify order
// ownership.
func extractReferenceID(result map[string]interface{}) string {
	units, _ := result["purchase_units"].([]interface{})
	if len(units) == 0 {
		return ""
	}
	unit, _ := units[0].(map[string]interface{})
	if unit == nil {
		return ""
	}
	refID, _ := unit["reference_id"].(string)
	return refID
}

// extractCaptureID reads the capture ID from a PayPal capture response.
// R33-H1: the capture ID is different from the order ID — PayPal refund
// webhooks use the capture ID as parent_id, so we must store it separately
// to look up users on refund events.
func extractCaptureID(result map[string]interface{}) string {
	units, _ := result["purchase_units"].([]interface{})
	if len(units) == 0 {
		return ""
	}
	unit, _ := units[0].(map[string]interface{})
	if unit == nil {
		return ""
	}
	payments, _ := unit["payments"].(map[string]interface{})
	if payments == nil {
		return ""
	}
	captures, _ := payments["captures"].([]interface{})
	if len(captures) == 0 {
		return ""
	}
	capture, _ := captures[0].(map[string]interface{})
	if capture == nil {
		return ""
	}
	id, _ := capture["id"].(string)
	return id
}

func (h *PaymentHandler) PayPalWebhook(c *fiber.Ctx) error {
	rawBody := c.Body()

	// Verify the webhook signature before trusting the payload. If
	// PAYPAL_WEBHOOK_ID is not configured, reject the webhook rather than
	// processing an unauthenticated event that could grant paid plans for free.
	verified, err := services.VerifyPayPalWebhook(c.UserContext(), rawBody, map[string]string{
		"PAYPAL-TRANSMISSION-ID":   c.Get("PAYPAL-TRANSMISSION-ID"),
		"PAYPAL-TRANSMISSION-TIME": c.Get("PAYPAL-TRANSMISSION-TIME"),
		"PAYPAL-TRANSMISSION-SIG":  c.Get("PAYPAL-TRANSMISSION-SIG"),
		"PAYPAL-CERT-URL":          c.Get("PAYPAL-CERT-URL"),
		"PAYPAL-AUTH-ALGO":         c.Get("PAYPAL-AUTH-ALGO"),
	})
	if err != nil {
		log.Printf("[PayPal Webhook] verification error: %v", err)
		return c.Status(503).JSON(fiber.Map{"error": "WEBHOOK_NOT_CONFIGURED", "message": "Webhook verification is not configured"})
	}
	if !verified {
		return c.Status(401).JSON(fiber.Map{"error": "INVALID_SIGNATURE", "message": "Webhook signature verification failed"})
	}

	var event struct {
		ID       string `json:"id"`
		Event    string `json:"event_type"`
		Resource struct {
			ParentID string `json:"parent_id"`
			ID       string `json:"id"`
			Status   string `json:"status"`
			Amount   *struct {
				Value        string `json:"value"`
				CurrencyCode string `json:"currency_code"`
			} `json:"amount"`
		} `json:"resource"`
	}

	if err := c.BodyParser(&event); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "INVALID_EVENT", "message": "Invalid webhook event"})
	}

	// Idempotency: PayPal delivers webhooks at-least-once and retries on
	// non-2xx responses. Use MarkSeen as an atomic check-and-set: it returns
	// true only for the first sighting of event.ID, false if already seen.
	// This eliminates the TOCTOU race that existed with the prior IsSeen+MarkSeen
	// split (two concurrent deliveries of the same event both passed IsSeen=false
	// before either reached MarkSeen, causing duplicate side effects).
	if !h.seenEvents.MarkSeen(event.ID) {
		log.Printf("[PayPal Webhook] Duplicate event %s ignored (idempotency)", event.ID)
		return c.JSON(fiber.Map{"success": true, "deduplicated": true})
	}
	// R55b-B2: safety net — if processing panics after MarkSeen but before
	// an explicit UnmarkSeen, the event would stay "seen" forever and PayPal's
	// retries would be silently deduplicated. This defer unmarks on panic so
	// PayPal can retry. Re-panic lets the recover middleware handle the 500.
	defer func() {
		if r := recover(); r != nil {
			h.seenEvents.UnmarkSeen(event.ID)
			log.Printf("[PayPal Webhook] Panic processing event %s, unmarked for retry: %v", event.ID, r)
			panic(r)
		}
	}()

	if event.Event == "PAYMENT.CAPTURE.COMPLETED" && event.Resource.Status == "COMPLETED" {
		orderID := event.Resource.ParentID
		if orderID != "" {
			// O(1) lookup via subscription ID reverse index instead of
			// scanning all users on every webhook event.
			if u, ok := h.userStore.GetBySubscriptionID(orderID); ok {
				originalPlan := u.Plan
				updated, ok := h.userStore.UpdateUser(u.Email, func(usr *models.User) {
					// Only upgrade free users — never downgrade an existing
					// paid plan (pro/enterprise) to avoid race with CapturePayPalOrder
					// which sets the correct plan synchronously.
					// R50-B2: This webhook fires before CapturePayPalOrder for
					// enterprise purchases, so the user briefly has "pro" instead
					// of "enterprise". This is eventually consistent — CapturePayPalOrder
					// runs right after (client-side capture) and upgrades "pro" →
					// "enterprise" via planRank comparison. PayPal's webhook payload
					// does not include reference_id/custom_id, so we cannot determine
					// the purchased plan here. Acceptable: the window is seconds,
					// and pro already unlocks all features.
					if usr.Plan == "free" || usr.Plan == "" {
						usr.Plan = "pro"
						usr.MaxFreeUsage = -1
					}
				})
				if !ok {
					// UpdateUser failed (user deleted between GetBySubscriptionID
					// and UpdateUser). Unmark the event so PayPal's retry will
					// be reprocessed instead of silently deduplicated.
					h.seenEvents.UnmarkSeen(event.ID)
					log.Printf("[PayPal Webhook] UpdateUser failed for %s, returning 500 for retry", hashEmail(u.Email))
					return c.Status(500).JSON(fiber.Map{"error": "UPDATE_FAILED", "message": "Failed to update user, will retry"})
				}
				// R38b-M4: skip persistence when the callback was a no-op (user
				// was already pro/enterprise). Without this, every duplicate
				// webhook event triggers an unnecessary UPDATE users SET plan=...
				// write, increasing SQLite write-lock contention.
				if updated.Plan != originalPlan && h.persistence != nil {
					// R37-B1: targeted UPDATE of plan fields only.
					if err := h.persistence.UpdateUserPlan(updated.Email, updated.Plan, updated.SubscriptionID, updated.CaptureID, updated.MaxFreeUsage); err != nil {
						// Persistence failed — unmark so PayPal retries and the
						// user's plan eventually gets persisted.
						h.seenEvents.UnmarkSeen(event.ID)
						log.Printf("[ERROR] Failed to persist webhook plan upgrade: %v", err)
						return c.Status(500).JSON(fiber.Map{"error": "PERSIST_FAILED", "message": "Failed to persist, will retry"})
					}
				}
			} else {
				// User not found — capture may not have run yet. Unmark so
				// PayPal retries; this allows reprocessing once capture completes.
				h.seenEvents.UnmarkSeen(event.ID)
				log.Printf("[PayPal Webhook] No user found for orderID %s, returning 500 for retry", orderID)
				return c.Status(500).JSON(fiber.Map{"error": "USER_NOT_FOUND", "message": "Capture not completed yet"})
			}
		} else {
			// R30-M4: COMPLETED event without parent_id cannot be processed.
			// Log and return 500 so PayPal retries — without this, the event
			// was silently marked as processed and the user never got upgraded.
			log.Printf("[PayPal Webhook] COMPLETED event %s missing parent_id", event.ID)
			h.seenEvents.UnmarkSeen(event.ID)
			return c.Status(500).JSON(fiber.Map{"error": "MISSING_PARENT_ID", "message": "Cannot process without parent_id"})
		}
	} else if event.Event == "PAYMENT.CAPTURE.REFUNDED" || event.Event == "PAYMENT.SALE.REFUNDED" {
		// Refund event — downgrade the user to free so they can't keep
		// using paid features after a refund. Without this, a user could
		// buy pro, refund via PayPal, and retain unlimited access.
		// R33-H1: for REFUND events, parent_id is the capture ID, not the
		// order ID. We now store CaptureID during capture, so try that
		// first. Fall back to GetBySubscriptionID for older captures that
		// predate the CaptureID field (backward compatibility).
		// R47-B1: do NOT fall back to event.Resource.ID when parent_id is
		// empty — for REFUND events, Resource.ID is the *refund* ID, not
		// the capture ID. Using it would (a) never match a user via
		// GetByCaptureID/GetBySubscriptionID, and (b) even if it did,
		// GetPayPalCapture would fail because the refund ID is not a
		// capture ID. Return 500 so PayPal retries — the payload may be
		// incomplete on first delivery.
		refundParentID := event.Resource.ParentID
		if refundParentID == "" {
			h.seenEvents.UnmarkSeen(event.ID)
			log.Printf("[PayPal Webhook] Refund %s: missing parent_id, cannot identify capture — retrying", event.ID)
			return c.Status(500).JSON(fiber.Map{"error": "MISSING_PARENT_ID", "message": "Refund parent_id missing, will retry"})
		}
		// Only process completed refunds — PENDING/DENIED refunds should
		// not trigger a downgrade. PayPal sends separate webhooks for
		// each status change, so marking PENDING as seen is correct.
		if event.Resource.Status != "" && event.Resource.Status != "COMPLETED" {
			log.Printf("[PayPal Webhook] Refund %s has status %s, not COMPLETED — skipping downgrade", event.ID, event.Resource.Status)
			return c.JSON(fiber.Map{"success": true, "skipped": "refund not completed"})
		}
		u, ok := h.userStore.GetByCaptureID(refundParentID)
		if !ok {
			u, ok = h.userStore.GetBySubscriptionID(refundParentID)
		}
		if ok {
			// R46-B1: use PayPal capture status instead of amount comparison
			// to determine full vs partial refund. The old amount comparison
			// was exploitable via near-full refunds ($9.98 vs $9.99) or split
			// refunds ($5 + $4.99) — each refund individually < capture amount,
			// so the plan was never downgraded despite full money returned.
			// PayPal's capture status is authoritative: "REFUNDED" = full,
			// "PARTIALLY_REFUNDED" = partial. Only downgrade on full refund.
			// R47-B2: removed the amount-existence check — the amount is no
			// longer used for the refund decision (captureStatus is the sole
			// authority). The check caused unnecessary 500 retries when the
			// amount field was missing but the capture status was available.
			token, err := services.GetPayPalAccessToken(c.UserContext())
			if err != nil {
				h.seenEvents.UnmarkSeen(event.ID)
				log.Printf("[PayPal Webhook] Refund %s: failed to get PayPal token for verification: %v", event.ID, err)
				return c.Status(500).JSON(fiber.Map{"error": "TOKEN_FAILED", "message": "Failed to verify refund, will retry"})
			}
			_, _, captureStatus, err := services.GetPayPalCapture(c.UserContext(), token, refundParentID)
			if err != nil {
				h.seenEvents.UnmarkSeen(event.ID)
				log.Printf("[PayPal Webhook] Refund %s: failed to fetch capture %s for status verification: %v", event.ID, refundParentID, err)
				return c.Status(500).JSON(fiber.Map{"error": "CAPTURE_LOOKUP_FAILED", "message": "Failed to verify refund status, will retry"})
			}
			if captureStatus != "REFUNDED" {
				log.Printf("[PayPal Webhook] Refund %s: capture status=%s — keeping user %s plan", event.ID, captureStatus, hashEmail(u.Email))
				return c.JSON(fiber.Map{"success": true, "skipped": "partial refund"})
			}
			// R46-B5: only downgrade if the user's current CaptureID matches
		// the refunded capture. If the user already re-purchased with a
		// different capture, downgrading would revoke the new purchase.
		// R55-B1: use a bool flag instead of comparing updated.Plan to
		// u.Plan (external snapshot from GetByCaptureID). Under concurrent
		// repurchase, u.Plan may differ from the pre-update value inside
		// UpdateUser, causing a false positive that persists "free" over
		// the new purchase.
		var downgraded bool
		_, ok = h.userStore.UpdateUser(u.Email, func(usr *models.User) {
			if usr.CaptureID != "" && usr.CaptureID != refundParentID {
				return
			}
			usr.Plan = "free"
			usr.MaxFreeUsage = 5
			usr.SubscriptionID = ""
			usr.CaptureID = ""
			downgraded = true
		})
		if !ok {
			h.seenEvents.UnmarkSeen(event.ID)
			log.Printf("[PayPal Webhook] UpdateUser failed for refund %s, returning 500", hashEmail(u.Email))
			return c.Status(500).JSON(fiber.Map{"error": "UPDATE_FAILED", "message": "Failed to update user, will retry"})
		}
		if downgraded && h.persistence != nil {
			if err := h.persistence.UpdateUserPlan(u.Email, "free", "", "", 5); err != nil {
				h.seenEvents.UnmarkSeen(event.ID)
				log.Printf("[ERROR] Failed to persist webhook refund downgrade: %v", err)
				return c.Status(500).JSON(fiber.Map{"error": "PERSIST_FAILED", "message": "Failed to persist, will retry"})
			}
		}
			log.Printf("[PayPal Webhook] User %s downgraded to free after full refund %s", hashEmail(u.Email), refundParentID)
		} else {
			// User not found by capture ID or subscription ID — may
			// already be free or order was from a different system.
			// Keep marked as seen to stop retries (no action to retry).
			log.Printf("[PayPal Webhook] Refund %s: no user found for captureID/orderID %s", event.ID, refundParentID)
		}
	} else {
		// Log unhandled event types for monitoring — PayPal may introduce new
		// event types that require business logic (disputes, subscriptions, etc.).
		log.Printf("[PayPal Webhook] Unhandled event type: %s, id: %s", event.Event, event.ID)
	}

	return c.JSON(fiber.Map{"success": true})
}
