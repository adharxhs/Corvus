package api_test

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"server/api"
	"server/auth"
	"server/database"
	"server/models"
	"server/repository"
	"server/services"
)

type dummyWSHandler struct{}

func (d *dummyWSHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func setupServer(t *testing.T) (http.Handler, *auth.Service) {
	t.Helper()
	dbPath := filepath.Join(t.TempDir(), "test.db")
	db, err := database.Open(dbPath)
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	t.Cleanup(func() { db.Close() })

	if err := database.MigrateDB(db); err != nil {
		t.Fatalf("migrate: %v", err)
	}

	repos := repository.New(db)
	svcs := services.New(repos, 24*time.Hour)
	authSvc := auth.NewService(repos.Users, "test-jwt-secret", 1*time.Hour)
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))

	router := api.NewRouter(logger, authSvc, svcs, &dummyWSHandler{}, "*")
	return router, authSvc
}

func TestHealthEndpoint(t *testing.T) {
	handler, _ := setupServer(t)
	req := httptest.NewRequest("GET", "/health", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200 OK, got %d", rec.Code)
	}
}

func TestRegisterAndLoginFlow(t *testing.T) {
	handler, _ := setupServer(t)

	// Register
	regBody, _ := json.Marshal(models.RegisterRequest{Username: "bob", Password: "password123"})
	req := httptest.NewRequest("POST", "/register", bytes.NewReader(regBody))
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201 Created, got %d: %s", rec.Code, rec.Body.String())
	}

	// Login
	loginBody, _ := json.Marshal(models.LoginRequest{Username: "bob", Password: "password123"})
	req = httptest.NewRequest("POST", "/login", bytes.NewReader(loginBody))
	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d: %s", rec.Code, rec.Body.String())
	}

	var loginResp models.LoginResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &loginResp); err != nil {
		t.Fatalf("decode login response: %v", err)
	}
	if loginResp.Token == "" {
		t.Error("expected non-empty token")
	}

	// Access protected route without token
	req = httptest.NewRequest("GET", "/chat-requests", nil)
	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected 401 Unauthorized, got %d", rec.Code)
	}

	// Access protected route with token
	req = httptest.NewRequest("GET", "/chat-requests", nil)
	req.Header.Set("Authorization", "Bearer "+loginResp.Token)
	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200 OK, got %d", rec.Code)
	}
}
