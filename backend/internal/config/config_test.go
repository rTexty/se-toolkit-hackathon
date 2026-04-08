package config

import (
	"os"
	"testing"
)

func TestLoadDefaults(t *testing.T) {
	os.Unsetenv("PORT")
	os.Unsetenv("DATABASE_URL")
	os.Unsetenv("JWT_SECRET")

	cfg := Load()

	if cfg.Port != "8080" {
		t.Errorf("Port = %s, want 8080", cfg.Port)
	}
	if cfg.JWTSecret != "super-secret-key" {
		t.Errorf("JWTSecret = %s, want super-secret-key", cfg.JWTSecret)
	}
}

func TestLoadFromEnv(t *testing.T) {
	os.Setenv("PORT", "9090")
	os.Setenv("JWT_SECRET", "custom-secret")
	defer func() {
		os.Unsetenv("PORT")
		os.Unsetenv("JWT_SECRET")
	}()

	cfg := Load()

	if cfg.Port != "9090" {
		t.Errorf("Port = %s, want 9090", cfg.Port)
	}
	if cfg.JWTSecret != "custom-secret" {
		t.Errorf("JWTSecret = %s, want custom-secret", cfg.JWTSecret)
	}
}

func TestGetEnv(t *testing.T) {
	os.Setenv("TEST_VAR", "hello")
	defer os.Unsetenv("TEST_VAR")

	if v := getEnv("TEST_VAR", "default"); v != "hello" {
		t.Errorf("getEnv = %s, want hello", v)
	}
	if v := getEnv("NON_EXISTENT", "fallback"); v != "fallback" {
		t.Errorf("getEnv = %s, want fallback", v)
	}
}
