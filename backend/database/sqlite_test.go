package database

import (
	"os"
	"testing"
	"time"

	"resumetake/models"
)

func setupTestDB(t *testing.T) *Database {
	t.Helper()
	dbPath := t.TempDir() + "/test.db"
	db, err := NewDatabase(dbPath)
	if err != nil {
		t.Fatalf("failed to create test database: %v", err)
	}
	return db
}

func TestCreateUser(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	user := &models.User{
		ID:           "test-id",
		Email:        "test@example.com",
		Password:     "hashed_password",
		PasswordType: "bcrypt",
		Name:         "Test User",
		Token:        "test-token",
		UsageCount:   0,
		MaxFreeUsage: 5,
		Plan:         "free",
		CreatedAt:    time.Now(),
	}

	err := db.SaveUser(user)
	if err != nil {
		t.Fatalf("failed to save user: %v", err)
	}

	saved, err := db.GetUser("test@example.com")
	if err != nil {
		t.Fatalf("failed to get user: %v", err)
	}
	if saved == nil {
		t.Fatal("user not found")
	}
	if saved.Email != "test@example.com" {
		t.Errorf("expected email test@example.com, got %s", saved.Email)
	}
	if saved.PasswordType != "bcrypt" {
		t.Errorf("expected password type bcrypt, got %s", saved.PasswordType)
	}
}

func TestGetUserByToken(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	user := &models.User{
		ID:           "test-id",
		Email:        "test@example.com",
		Password:     "hashed_password",
		PasswordType: "bcrypt",
		Name:         "Test User",
		Token:        "my-token",
		Plan:         "free",
		CreatedAt:    time.Now(),
	}

	db.SaveUser(user)

	saved, err := db.GetUserByToken("my-token")
	if err != nil {
		t.Fatalf("failed to get user by token: %v", err)
	}
	if saved == nil {
		t.Fatal("user not found by token")
	}
	if saved.Email != "test@example.com" {
		t.Errorf("expected email test@example.com, got %s", saved.Email)
	}
}

func TestDeleteUser(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	user := &models.User{
		ID:        "test-id",
		Email:     "test@example.com",
		Password:  "hashed_password",
		Name:      "Test User",
		Token:     "test-token",
		CreatedAt: time.Now(),
	}

	db.SaveUser(user)

	err := db.DeleteUser("test@example.com")
	if err != nil {
		t.Fatalf("failed to delete user: %v", err)
	}

	saved, err := db.GetUser("test@example.com")
	if err != nil {
		t.Fatalf("failed to get user: %v", err)
	}
	if saved != nil {
		t.Error("user should be deleted")
	}
}

func TestCreateResume(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	resume := &models.Resume{
		ID:    "resume-id",
		Title: "Test Resume",
		Content: map[string]interface{}{
			"summary": "Test summary",
		},
		CreatedAt: time.Now().Format(time.RFC3339),
		UpdatedAt: time.Now().Format(time.RFC3339),
	}

	err := db.SaveResume(resume)
	if err != nil {
		t.Fatalf("failed to save resume: %v", err)
	}

	saved, err := db.GetResume("resume-id")
	if err != nil {
		t.Fatalf("failed to get resume: %v", err)
	}
	if saved == nil {
		t.Fatal("resume not found")
	}
	if saved.Title != "Test Resume" {
		t.Errorf("expected title Test Resume, got %s", saved.Title)
	}
}

func TestMigrateFromJSON(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	users := map[string]*models.User{
		"test@example.com": {
			ID:           "test-id",
			Email:        "test@example.com",
			Password:     "hashed_password",
			PasswordType: "bcrypt",
			Name:         "Test User",
			Token:        "test-token",
			Plan:         "free",
			CreatedAt:    time.Now(),
		},
	}

	err := db.MigrateFromJSON(users)
	if err != nil {
		t.Fatalf("failed to migrate: %v", err)
	}

	saved, err := db.GetUser("test@example.com")
	if err != nil {
		t.Fatalf("failed to get user: %v", err)
	}
	if saved == nil {
		t.Fatal("user not found after migration")
	}
}

func TestGetAllUsers(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	for i := 0; i < 3; i++ {
		user := &models.User{
			ID:        "id-" + string(rune('0'+i)),
			Email:     "user" + string(rune('0'+i)) + "@example.com",
			Password:  "password",
			Name:      "User",
			CreatedAt: time.Now(),
		}
		db.SaveUser(user)
	}

	users, err := db.GetAllUsers()
	if err != nil {
		t.Fatalf("failed to get all users: %v", err)
	}
	if len(users) != 3 {
		t.Errorf("expected 3 users, got %d", len(users))
	}
}

func TestClose(t *testing.T) {
	dbPath := t.TempDir() + "/test.db"
	db, err := NewDatabase(dbPath)
	if err != nil {
		t.Fatalf("failed to create database: %v", err)
	}

	err = db.Close()
	if err != nil {
		t.Fatalf("failed to close database: %v", err)
	}
}

func TestNonExistentUser(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	user, err := db.GetUser("nonexistent@example.com")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if user != nil {
		t.Error("expected nil for non-existent user")
	}
}

func TestMain(m *testing.M) {
	os.Exit(m.Run())
}
