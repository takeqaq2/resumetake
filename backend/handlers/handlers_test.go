package handlers

import (
	"bytes"
	"encoding/json"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"

	"resumetake/models"
	"resumetake/store"
)

func setupTestApp() *fiber.App {
	app := fiber.New()
	return app
}

func TestHealthEndpoint(t *testing.T) {
	app := setupTestApp()
	resumeStore := models.NewStore()
	userStore := models.NewUserStore()
	handler := NewResumeHandler(resumeStore, userStore, nil)

	app.Get("/api/health", handler.Health)

	req := httptest.NewRequest("GET", "/api/health", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("failed to test health endpoint: %v", err)
	}

	if resp.StatusCode != 200 {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}
}

func TestCreateResume(t *testing.T) {
	app := setupTestApp()
	resumeStore := models.NewStore()
	userStore := models.NewUserStore()
	handler := NewResumeHandler(resumeStore, userStore, nil)

	testUser := &models.User{ID: "u1", Email: "test@example.com", Name: "Test User"}
	app.Use(func(c *fiber.Ctx) error {
		c.Locals("user", testUser)
		return c.Next()
	})
	app.Post("/api/v1/resumes", handler.Create)

	body := map[string]interface{}{
		"title": "Test Resume",
		"content": map[string]interface{}{
			"personalInfo": map[string]interface{}{
				"name": "Test User",
			},
		},
	}
	bodyBytes, _ := json.Marshal(body)

	req := httptest.NewRequest("POST", "/api/v1/resumes", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("failed to test create resume: %v", err)
	}

	if resp.StatusCode != 201 {
		t.Errorf("expected status 201, got %d", resp.StatusCode)
	}

	// Verify the resume is owned by the authenticated user.
	var respBody map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&respBody)
	data := respBody["data"].(map[string]interface{})
	if data["owner_id"] != "u1" {
		t.Errorf("expected owner_id u1, got %v", data["owner_id"])
	}
}

func TestGetJobs(t *testing.T) {
	app := setupTestApp()
	handler := NewJobHandler()

	app.Get("/api/v1/jobs", handler.GetJobs)

	req := httptest.NewRequest("GET", "/api/v1/jobs", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("failed to test get jobs: %v", err)
	}

	if resp.StatusCode != 200 {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}
}

func TestGetPricing(t *testing.T) {
	app := setupTestApp()
	userStore := models.NewUserStore()
	persistence := store.NewUserPersistence(t.TempDir() + "/test_users.json")
	handler := NewPaymentHandler(userStore, persistence)

	app.Get("/api/v1/pricing", handler.GetPricing)

	req := httptest.NewRequest("GET", "/api/v1/pricing", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("failed to test get pricing: %v", err)
	}

	if resp.StatusCode != 200 {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}
}

func TestSendVerificationCode(t *testing.T) {
	app := setupTestApp()
	userStore := models.NewUserStore()
	verificationStore := models.NewVerificationStore()
	persistence := store.NewUserPersistence(t.TempDir() + "/test_users.json")
	handler := NewAuthHandler(userStore, verificationStore, persistence)

	app.Post("/api/v1/auth/send-code", handler.SendCode)

	body := map[string]interface{}{
		"email": "test@example.com",
	}
	bodyBytes, _ := json.Marshal(body)

	req := httptest.NewRequest("POST", "/api/v1/auth/send-code", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("failed to test send verification code: %v", err)
	}

	// SMTP is not configured in the test environment, so SendCode should
	// return 500 (EMAIL_SEND_FAILED) rather than silently succeeding.
	if resp.StatusCode != 500 {
		t.Errorf("expected status 500 (SMTP not configured), got %d", resp.StatusCode)
	}
}
