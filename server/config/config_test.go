package config_test

import (
	"os"
	"testing"
	"time"

	"server/config"
)

func TestConfigLoad(t *testing.T) {
	os.Setenv("JWT_SECRET", "test-secret-key")
	defer os.Unsetenv("JWT_SECRET")

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("expected successful load, got %v", err)
	}

	if cfg.HTTPPort != "8080" {
		t.Errorf("expected port 8080, got %s", cfg.HTTPPort)
	}
	if cfg.DatabasePath != "corvus.db" {
		t.Errorf("expected db corvus.db, got %s", cfg.DatabasePath)
	}
	if cfg.JWTSecret != "test-secret-key" {
		t.Errorf("expected secret test-secret-key, got %s", cfg.JWTSecret)
	}
	if cfg.JWTExpiration != 24*time.Hour {
		t.Errorf("expected 24h expiration, got %v", cfg.JWTExpiration)
	}
}

func TestConfigMissingSecret(t *testing.T) {
	os.Unsetenv("JWT_SECRET")

	_, err := config.Load()
	if err == nil {
		t.Fatal("expected error when JWT_SECRET is missing, got nil")
	}
}
