package services

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
)

// Cap for reading PayPal HTTP bodies. PayPal API responses are small JSON
// payloads (tokens, orders, verification results); 1MB is generous and
// prevents a compromised endpoint from OOM-ing the process.
const maxPayPalResponseBytes = 1 << 20 // 1MB

func GetPayPalBaseURL() string {
	mode := os.Getenv("PAYPAL_MODE")
	if mode == "production" {
		return "https://api-m.paypal.com"
	}
	return "https://api-m.sandbox.paypal.com"
}

var (
	paypalTokenCache     string
	paypalTokenExpiry    time.Time
	paypalTokenCacheMu   sync.RWMutex // protects cache reads/writes
	paypalTokenRefreshMu sync.Mutex   // serializes refresh attempts (singleflight)
)

func GetPayPalAccessToken(ctx context.Context) (string, error) {
	clientID := os.Getenv("PAYPAL_CLIENT_ID")
	clientSecret := os.Getenv("PAYPAL_CLIENT_SECRET")
	if clientID == "" || clientSecret == "" {
		return "", fmt.Errorf("PayPal not configured")
	}

	// Fast path: return cached token under read lock (microseconds).
	paypalTokenCacheMu.RLock()
	if paypalTokenCache != "" && time.Now().Before(paypalTokenExpiry.Add(-5*time.Minute)) {
		t := paypalTokenCache
		paypalTokenCacheMu.RUnlock()
		return t, nil
	}
	paypalTokenCacheMu.RUnlock()

	// R31-6: singleflight — serialize token refresh so concurrent payment
	// requests share a single PayPal OAuth call instead of each firing its
	// own (thundering herd). Other goroutines block on refreshMu, then
	// re-check the cache (the first caller already refreshed it).
	paypalTokenRefreshMu.Lock()
	defer paypalTokenRefreshMu.Unlock()

	// Re-check cache after acquiring refreshMu — another goroutine may have
	// refreshed while we were waiting.
	paypalTokenCacheMu.RLock()
	if paypalTokenCache != "" && time.Now().Before(paypalTokenExpiry.Add(-5*time.Minute)) {
		t := paypalTokenCache
		paypalTokenCacheMu.RUnlock()
		return t, nil
	}
	paypalTokenCacheMu.RUnlock()

	creds := base64.StdEncoding.EncodeToString([]byte(clientID + ":" + clientSecret))
	apiURL := GetPayPalBaseURL() + "/v1/oauth2/token"

	req, err := http.NewRequestWithContext(ctx, "POST", apiURL, strings.NewReader("grant_type=client_credentials"))
	if err != nil {
		return "", fmt.Errorf("failed to create token request")
	}
	req.Header.Set("Authorization", "Basic "+creds)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to get PayPal token: %w", err)
	}
	defer resp.Body.Close()

	body, readErr := io.ReadAll(io.LimitReader(resp.Body, maxPayPalResponseBytes))
	if readErr != nil {
		return "", fmt.Errorf("failed to read PayPal token response (HTTP %d): %w", resp.StatusCode, readErr)
	}
	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return "", fmt.Errorf("invalid PayPal token response (HTTP %d): %w", resp.StatusCode, err)
	}

	if resp.StatusCode != 200 {
		return "", fmt.Errorf("PayPal token request failed: HTTP %d", resp.StatusCode)
	}

	token, ok := result["access_token"].(string)
	if !ok || token == "" {
		return "", fmt.Errorf("no access token in PayPal response (HTTP %d)", resp.StatusCode)
	}

	expiresIn := 3600
	if exp, ok := result["expires_in"].(float64); ok && exp > 0 {
		expiresIn = int(exp)
	}
	// R45-B4: floor the expiry at 10 minutes so the 5-minute refresh buffer
	// (line 45: paypalTokenExpiry.Add(-5*time.Minute)) never produces a past
	// time. Without this, an unusually short expires_in (<300s) would make
	// time.Now().Before(pastTime) always false, causing a token refresh on
	// every single PayPal API call.
	if expiresIn < 600 {
		expiresIn = 600
	}

	// Update cache under write lock.
	paypalTokenCacheMu.Lock()
	paypalTokenCache = token
	paypalTokenExpiry = time.Now().Add(time.Duration(expiresIn) * time.Second)
	paypalTokenCacheMu.Unlock()

	return token, nil
}

func CreatePayPalOrder(ctx context.Context, token, amount, currency, description, orderID, returnURL, cancelURL string) (string, error) {
	apiURL := GetPayPalBaseURL() + "/v2/checkout/orders"

	orderPayload := map[string]interface{}{
		"intent": "CAPTURE",
		"purchase_units": []map[string]interface{}{
			{
				"reference_id": orderID,
				"description":  description,
				"amount": map[string]interface{}{
					"currency_code": currency,
					"value":         amount,
				},
			},
		},
		"application_context": map[string]interface{}{
			"return_url":   returnURL,
			"cancel_url":   cancelURL,
			"brand_name":   "ResumeTake",
			"landing_page": "BILLING",
			"user_action":  "PAY_NOW",
		},
	}

	jsonData, err := json.Marshal(orderPayload)
	if err != nil {
		return "", fmt.Errorf("failed to marshal order")
	}

	req, err := http.NewRequestWithContext(ctx, "POST", apiURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to create order request")
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Prefer", "return=representation")

	resp, err := httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to create PayPal order: %w", err)
	}
	defer resp.Body.Close()

	body, readErr := io.ReadAll(io.LimitReader(resp.Body, maxPayPalResponseBytes))
	if readErr != nil {
		return "", fmt.Errorf("failed to read PayPal order response: %w", readErr)
	}
	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return "", fmt.Errorf("invalid PayPal order response: %w", err)
	}

	if resp.StatusCode != 201 {
		errMsg := "Unknown error"
		if e, ok := result["message"].(string); ok {
			errMsg = e
		}
		return "", fmt.Errorf("PayPal order creation failed: %s", errMsg)
	}

	orderIDResp, ok := result["id"].(string)
	if !ok || orderIDResp == "" {
		return "", fmt.Errorf("no order ID in PayPal response")
	}

	return orderIDResp, nil
}

func CapturePayPalOrder(ctx context.Context, token, orderID string) (map[string]interface{}, error) {
	// R54-B2: defense-in-depth — validate orderID format before URL
	// interpolation, matching GetPayPalCapture's validation. Prevents
	// path/API endpoint manipulation if a future caller skips validation.
	for _, ch := range orderID {
		if !((ch >= 'A' && ch <= 'Z') || (ch >= 'a' && ch <= 'z') || (ch >= '0' && ch <= '9') || ch == '-' || ch == '_') {
			return nil, fmt.Errorf("invalid order ID format")
		}
	}
	apiURL := GetPayPalBaseURL() + "/v2/checkout/orders/" + orderID + "/capture"

	req, err := http.NewRequestWithContext(ctx, "POST", apiURL, strings.NewReader("{}"))
	if err != nil {
		return nil, fmt.Errorf("failed to create capture request")
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Prefer", "return=representation")

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to capture PayPal order: %w", err)
	}
	defer resp.Body.Close()

	body, readErr := io.ReadAll(io.LimitReader(resp.Body, maxPayPalResponseBytes))
	if readErr != nil {
		return nil, fmt.Errorf("failed to read PayPal capture response: %w", readErr)
	}
	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("invalid PayPal capture response: %w", err)
	}

	if resp.StatusCode != 200 && resp.StatusCode != 201 {
		errMsg := "Unknown error"
		if e, ok := result["message"].(string); ok {
			errMsg = e
		}
		return nil, fmt.Errorf("PayPal capture failed: %s", errMsg)
	}

	return result, nil
}

// VerifyPayPalWebhook verifies the signature of an incoming PayPal webhook
// using PayPal's verify-webhook-signature API. The raw body and the
// PAYPAL-TRANSMISSION-* / PAYPAL-CERT-URL / PAYPAL-AUTH-ALGO headers must be
// forwarded from the incoming request. Returns true if the signature is valid.
func VerifyPayPalWebhook(ctx context.Context, rawBody []byte, headers map[string]string) (bool, error) {
	webhookID := os.Getenv("PAYPAL_WEBHOOK_ID")
	if webhookID == "" {
		return false, fmt.Errorf("PAYPAL_WEBHOOK_ID not configured")
	}

	transmissionID := headers["PAYPAL-TRANSMISSION-ID"]
	transmissionTime := headers["PAYPAL-TRANSMISSION-TIME"]
	transmissionSig := headers["PAYPAL-TRANSMISSION-SIG"]
	certURL := headers["PAYPAL-CERT-URL"]
	authAlgo := headers["PAYPAL-AUTH-ALGO"]

	if transmissionID == "" || transmissionSig == "" || certURL == "" || authAlgo == "" {
		return false, fmt.Errorf("missing PayPal webhook signature headers")
	}

	// Defense-in-depth: validate cert_url hostname to prevent SSRF. An
	// attacker could supply a forged cert_url pointing to their own server
	// with a self-signed cert. PayPal's verify endpoint fetches this URL to
	// obtain the certificate used for signature verification.
	if !strings.HasPrefix(certURL, "https://") ||
		(!strings.HasPrefix(certURL, "https://www.paypal.com/") &&
			!strings.HasPrefix(certURL, "https://api.paypal.com/") &&
			!strings.HasPrefix(certURL, "https://sfapi7.paypal.com/")) {
		return false, fmt.Errorf("untrusted cert_url host")
	}

	token, err := GetPayPalAccessToken(ctx)
	if err != nil {
		return false, fmt.Errorf("failed to get PayPal access token: %w", err)
	}

	payload := map[string]interface{}{
		"transmission_id":   transmissionID,
		"transmission_time": transmissionTime,
		"cert_url":          certURL,
		"auth_algo":         authAlgo,
		"transmission_sig":  transmissionSig,
		"webhook_id":        webhookID,
		"webhook_event":     json.RawMessage(rawBody),
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return false, fmt.Errorf("failed to marshal verification payload")
	}

	apiURL := GetPayPalBaseURL() + "/v1/verify/webhook-signature"
	req, err := http.NewRequestWithContext(ctx, "POST", apiURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return false, fmt.Errorf("failed to create verification request")
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := httpClient.Do(req)
	if err != nil {
		return false, fmt.Errorf("failed to call PayPal verification: %w", err)
	}
	defer resp.Body.Close()

	body, readErr := io.ReadAll(io.LimitReader(resp.Body, maxPayPalResponseBytes))
	if readErr != nil {
		return false, fmt.Errorf("failed to read PayPal verification response (HTTP %d): %w", resp.StatusCode, readErr)
	}
	var result struct {
		VerificationStatus string `json:"verification_status"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return false, fmt.Errorf("invalid PayPal verification response (HTTP %d): %w", resp.StatusCode, err)
	}

	if resp.StatusCode != 200 {
		return false, fmt.Errorf("PayPal verification API returned HTTP %d", resp.StatusCode)
	}

	return result.VerificationStatus == "SUCCESS", nil
}

// GetPayPalCapture retrieves the details of a captured payment, including
// the original amount and refund status. Used by the REFUND webhook handler
// to determine whether a refund is full or partial — only full refunds
// (status=="REFUNDED") should trigger a plan downgrade.
//
// R46-B1: previously returned only amount/currency, and the caller compared
// single-refund amount vs capture amount. This was exploitable via near-full
// refunds (e.g. $9.98 vs $9.99) or split refunds ($5 + $4.99) — each refund
// individually < capture amount, so the plan was never downgraded despite
// the user receiving a full refund. The capture `status` field is the
// authoritative signal: PayPal sets it to "PARTIALLY_REFUNDED" after any
// partial refund and "REFUNDED" only when the full amount has been refunded.
func GetPayPalCapture(ctx context.Context, token, captureID string) (amount, currency, status string, err error) {
	// Defense-in-depth: validate captureID format before URL interpolation.
	// PayPal capture IDs are uppercase alphanumeric with hyphens (e.g.
	// "92G12345AB678901X"). Reject anything containing path separators or
	// query characters to prevent API endpoint manipulation.
	// R54b-B6: allow lowercase to align with CapturePayPalOrder's OrderID
	// validation — PayPal IDs may be mixed-case across endpoints.
	for _, ch := range captureID {
		if !((ch >= 'A' && ch <= 'Z') || (ch >= 'a' && ch <= 'z') || (ch >= '0' && ch <= '9') || ch == '-' || ch == '_') {
			return "", "", "", fmt.Errorf("invalid capture ID format")
		}
	}
	apiURL := GetPayPalBaseURL() + "/v2/payments/captures/" + captureID
	req, err := http.NewRequestWithContext(ctx, "GET", apiURL, nil)
	if err != nil {
		return "", "", "", fmt.Errorf("failed to create capture request")
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := httpClient.Do(req)
	if err != nil {
		return "", "", "", fmt.Errorf("failed to fetch capture: %w", err)
	}
	defer resp.Body.Close()

	body, readErr := io.ReadAll(io.LimitReader(resp.Body, maxPayPalResponseBytes))
	if readErr != nil {
		return "", "", "", fmt.Errorf("failed to read PayPal capture response: %w", readErr)
	}
	var result struct {
		Amount struct {
			Value        string `json:"value"`
			CurrencyCode string `json:"currency_code"`
		} `json:"amount"`
		Status string `json:"status"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return "", "", "", fmt.Errorf("invalid PayPal capture response: %w", err)
	}
	if resp.StatusCode != 200 {
		return "", "", "", fmt.Errorf("PayPal capture lookup failed: %d", resp.StatusCode)
	}
	return result.Amount.Value, result.Amount.CurrencyCode, result.Status, nil
}
