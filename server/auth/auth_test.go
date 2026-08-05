package auth_test

import (
	"path/filepath"
	"testing"
	"time"

	"server/auth"
	"server/database"
	"server/repository"
)

func setupTestDB(t *testing.T) *repository.Repository {
	t.Helper()
	dbPath := filepath.Join(t.TempDir(), "test.db")
	db, err := database.Open(dbPath)
	if err != nil {
		t.Fatalf("failed to open test db: %v", err)
	}
	t.Cleanup(func() { db.Close() })

	if err := database.MigrateDB(db); err != nil {
		t.Fatalf("failed to migrate test db: %v", err)
	}
	return repository.New(db)
}

func TestPasswordHashing(t *testing.T) {
	password := "super-secret-password-123"
	hash, err := auth.HashPassword(password)
	if err != nil {
		t.Fatalf("failed to hash password: %v", err)
	}

	ok, err := auth.VerifyPassword(password, hash)
	if err != nil || !ok {
		t.Errorf("password verification failed: ok=%v, err=%v", ok, err)
	}

	ok, _ = auth.VerifyPassword("wrong-password", hash)
	if ok {
		t.Error("expected verify to fail for wrong password")
	}
}

func TestJWTManager(t *testing.T) {
	mgr := auth.NewJWTManager("test-secret", 1*time.Hour)
	token, err := mgr.Issue("user-1", "alice")
	if err != nil {
		t.Fatalf("failed to issue token: %v", err)
	}

	claims, err := mgr.Validate(token)
	if err != nil {
		t.Fatalf("failed to validate token: %v", err)
	}
	if claims.UserID != "user-1" || claims.Username != "alice" {
		t.Errorf("unexpected claims: %+v", claims)
	}
}

func TestAuthServiceRegisterAndLogin(t *testing.T) {
	repos := setupTestDB(t)
	svc := auth.NewService(repos.Users, "test-secret", 1*time.Hour)

	regResp, err := svc.Register("alice", "password123")
	if err != nil {
		t.Fatalf("registration failed: %v", err)
	}
	if regResp.User.Username != "alice" {
		t.Errorf("expected username alice, got %s", regResp.User.Username)
	}

	// Duplicate registration should fail
	_, err = svc.Register("alice", "password123")
	if err == nil {
		t.Fatal("expected duplicate registration to fail")
	}

	// Login with correct password
	loginResp, err := svc.Login("alice", "password123")
	if err != nil {
		t.Fatalf("login failed: %v", err)
	}
	if loginResp.Token == "" {
		t.Error("expected non-empty token")
	}

	// Login with wrong password
	_, err = svc.Login("alice", "wrongpass")
	if err == nil {
		t.Fatal("expected login with wrong password to fail")
	}
}
