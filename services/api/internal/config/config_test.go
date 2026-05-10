package config

import (
	"os"
	"testing"
)

func TestLoad_Defaults(t *testing.T) {
	cfg := Load()

	if cfg.Port != "8080" {
		t.Errorf("expected Port 8080, got %s", cfg.Port)
	}

	if cfg.Database.Host != "localhost" {
		t.Errorf("expected DB Host localhost, got %s", cfg.Database.Host)
	}

	if cfg.Database.Port != "5432" {
		t.Errorf("expected DB Port 5432, got %s", cfg.Database.Port)
	}

	if cfg.Database.User != "postgres" {
		t.Errorf("expected DB User postgres, got %s", cfg.Database.User)
	}

	if cfg.Database.Password != "postgres" {
		t.Errorf("expected DB Password postgres, got %s", cfg.Database.Password)
	}

	if cfg.Database.Name != "aidevt" {
		t.Errorf("expected DB Name aidevt, got %s", cfg.Database.Name)
	}
}

func TestLoad_FromEnv(t *testing.T) {
	os.Setenv("PORT", "9090")
	os.Setenv("DB_HOST", "test-host")
	os.Setenv("DB_PORT", "5433")
	os.Setenv("DB_USER", "test-user")
	os.Setenv("DB_PASSWORD", "test-pass")
	os.Setenv("DB_NAME", "test-db")

	defer func() {
		os.Unsetenv("PORT")
		os.Unsetenv("DB_HOST")
		os.Unsetenv("DB_PORT")
		os.Unsetenv("DB_USER")
		os.Unsetenv("DB_PASSWORD")
		os.Unsetenv("DB_NAME")
	}()

	cfg := Load()

	if cfg.Port != "9090" {
		t.Errorf("expected Port 9090, got %s", cfg.Port)
	}

	if cfg.Database.Host != "test-host" {
		t.Errorf("expected DB Host test-host, got %s", cfg.Database.Host)
	}

	if cfg.Database.Port != "5433" {
		t.Errorf("expected DB Port 5433, got %s", cfg.Database.Port)
	}

	if cfg.Database.User != "test-user" {
		t.Errorf("expected DB User test-user, got %s", cfg.Database.User)
	}

	if cfg.Database.Password != "test-pass" {
		t.Errorf("expected DB Password test-pass, got %s", cfg.Database.Password)
	}

	if cfg.Database.Name != "test-db" {
		t.Errorf("expected DB Name test-db, got %s", cfg.Database.Name)
	}
}

func TestGetEnv(t *testing.T) {
	key := "TEST_ENV_VAR"

	if value := getEnv(key, "default"); value != "default" {
		t.Errorf("expected default, got %s", value)
	}

	os.Setenv(key, "custom-value")
	defer os.Unsetenv(key)

	if value := getEnv(key, "default"); value != "custom-value" {
		t.Errorf("expected custom-value, got %s", value)
	}
}
