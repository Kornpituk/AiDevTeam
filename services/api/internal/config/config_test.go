package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"
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

func TestDefaultTimeouts(t *testing.T) {
	// Unset any env vars that might interfere
	os.Unsetenv("STEP_TIMEOUT")
	os.Unsetenv("RUN_TIMEOUT")

	cfg := Load()

	if cfg.Tool.StepTimeout != 300*time.Second {
		t.Errorf("expected StepTimeout 300s, got %v", cfg.Tool.StepTimeout)
	}

	if cfg.Tool.RunTimeout != 1800*time.Second {
		t.Errorf("expected RunTimeout 1800s, got %v", cfg.Tool.RunTimeout)
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

func TestGetEnvInt(t *testing.T) {
	key := "TEST_ENV_INT"

	// Returns default when env unset
	if value := getEnvInt(key, 42); value != 42 {
		t.Errorf("expected default 42, got %d", value)
	}

	// Returns value when env set
	os.Setenv(key, "99")
	defer os.Unsetenv(key)

	if value := getEnvInt(key, 42); value != 99 {
		t.Errorf("expected 99, got %d", value)
	}
}

func TestGetEnvInt64(t *testing.T) {
	key := "TEST_ENV_INT64"

	// Returns default when env unset
	if value := getEnvInt64(key, 1048576); value != 1048576 {
		t.Errorf("expected default 1048576, got %d", value)
	}

	// Returns value when env set
	os.Setenv(key, "2097152")
	defer os.Unsetenv(key)

	if value := getEnvInt64(key, 1048576); value != 2097152 {
		t.Errorf("expected 2097152, got %d", value)
	}
}

func TestGetEnvDuration(t *testing.T) {
	key := "TEST_ENV_DURATION"

	// Returns default when env unset
	if value := getEnvDuration(key, 60*time.Second); value != 60*time.Second {
		t.Errorf("expected default 60s, got %v", value)
	}

	// Returns value when env set
	os.Setenv(key, "120s")
	defer os.Unsetenv(key)

	if value := getEnvDuration(key, 60*time.Second); value != 120*time.Second {
		t.Errorf("expected 120s, got %v", value)
	}

	// Handles invalid duration strings gracefully (returns default)
	os.Setenv(key, "not-a-duration")
	if value := getEnvDuration(key, 30*time.Second); value != 30*time.Second {
		t.Errorf("expected default 30s for invalid duration, got %v", value)
	}
}

func TestLoadEnvFile(t *testing.T) {
	dir := t.TempDir()
	envPath := filepath.Join(dir, ".env")

	content := `# This is a comment
DB_HOST=testhost
DB_PORT=5433
EMPTY_LINE_ABOVE=

DB_NAME="mydb"
DB_USER='admin'
`
	if err := os.WriteFile(envPath, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write .env file: %v", err)
	}

	// Clear any pre-existing env vars for test keys
	os.Unsetenv("DB_HOST")
	os.Unsetenv("DB_PORT")
	os.Unsetenv("DB_NAME")
	os.Unsetenv("DB_USER")

	loadEnvFile(envPath)

	tests := []struct {
		key      string
		expected string
	}{
		{"DB_HOST", "testhost"},
		{"DB_PORT", "5433"},
		{"DB_NAME", "mydb"},
		{"DB_USER", "admin"},
	}

	for _, tt := range tests {
		t.Run(tt.key, func(t *testing.T) {
			got := os.Getenv(tt.key)
			if got != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, got)
			}
		})
	}

	// Cleanup
	os.Unsetenv("DB_HOST")
	os.Unsetenv("DB_PORT")
	os.Unsetenv("DB_NAME")
	os.Unsetenv("DB_USER")
}

func TestLoad_LLMDefaults(t *testing.T) {
	// Unset relevant env vars
	os.Unsetenv("LLM_PROVIDER")
	os.Unsetenv("LLM_MODEL")
	os.Unsetenv("LLM_TIMEOUT")
	os.Unsetenv("LLM_API_KEY")
	os.Unsetenv("LLM_BASE_URL")

	cfg := Load()

	if cfg.LLM.Provider != "openai" {
		t.Errorf("expected Provider 'openai', got %q", cfg.LLM.Provider)
	}
	if cfg.LLM.Model != "gpt-4" {
		t.Errorf("expected Model 'gpt-4', got %q", cfg.LLM.Model)
	}
	if cfg.LLM.Timeout != 60*time.Second {
		t.Errorf("expected Timeout 60s, got %v", cfg.LLM.Timeout)
	}
	if cfg.LLM.APIKey != "" {
		t.Errorf("expected empty APIKey, got %q", cfg.LLM.APIKey)
	}
	if cfg.LLM.BaseURL != "" {
		t.Errorf("expected empty BaseURL, got %q", cfg.LLM.BaseURL)
	}
}

func TestLoad_ToolDefaults(t *testing.T) {
	// Unset relevant env vars
	os.Unsetenv("TOOL_READ_MAX_BYTES")
	os.Unsetenv("TOOL_SEARCH_MAX_RESULTS")
	os.Unsetenv("TOOL_MAX_ITERATIONS")
	os.Unsetenv("TOOL_CHOICE")
	os.Unsetenv("TOOL_WRITE_MAX_BYTES")
	os.Unsetenv("TOOL_BASH_TIMEOUT")

	cfg := Load()

	if cfg.Tool.ReadMaxBytes != 1048576 {
		t.Errorf("expected ReadMaxBytes 1048576, got %d", cfg.Tool.ReadMaxBytes)
	}
	if cfg.Tool.SearchMaxResults != 50 {
		t.Errorf("expected SearchMaxResults 50, got %d", cfg.Tool.SearchMaxResults)
	}
	if cfg.Tool.MaxToolIterations != 10 {
		t.Errorf("expected MaxToolIterations 10, got %d", cfg.Tool.MaxToolIterations)
	}
	if cfg.Tool.ToolChoice != "auto" {
		t.Errorf("expected ToolChoice 'auto', got %q", cfg.Tool.ToolChoice)
	}
	if cfg.Tool.WriteMaxBytes != 1048576 {
		t.Errorf("expected WriteMaxBytes 1048576, got %d", cfg.Tool.WriteMaxBytes)
	}
	if cfg.Tool.BashTimeout != 30 {
		t.Errorf("expected BashTimeout 30, got %d", cfg.Tool.BashTimeout)
	}
}
