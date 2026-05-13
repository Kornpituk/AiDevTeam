package config

import (
	"bufio"
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	Port     string
	Database DatabaseConfig
	LLM      LLMConfig
	Tool     ToolConfig
}

type ToolConfig struct {
	WorkspaceRoot     string
	ReadMaxBytes      int64
	SearchMaxResults  int
	MaxToolIterations int
	RequireApproval   string
	ToolChoice        string
	WriteMaxBytes     int64
	BashTimeout       int
	BashBlocked       string
}

type DatabaseConfig struct {
	Host            string
	Port            string
	User            string
	Password        string
	Name            string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
	ConnMaxIdleTime time.Duration
}

type LLMConfig struct {
	Provider string
	APIKey   string
	Model    string
	BaseURL  string
	Timeout  time.Duration
}

// loadEnvFile reads a .env file and sets environment variables.
// This avoids needing an external dependency like godotenv.
func loadEnvFile(path string) {
	f, err := os.Open(path)
	if err != nil {
		return
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Parse KEY=VALUE
		eq := strings.Index(line, "=")
		if eq < 0 {
			continue
		}

		key := strings.TrimSpace(line[:eq])
		value := strings.TrimSpace(line[eq+1:])

		// Remove surrounding quotes if present
		if len(value) >= 2 && ((value[0] == '"' && value[len(value)-1] == '"') || (value[0] == '\'' && value[len(value)-1] == '\'')) {
			value = value[1 : len(value)-1]
		}

		// Only set if not already set (OS env vars take precedence)
		if os.Getenv(key) == "" {
			os.Setenv(key, value)
		}
	}
}

func Load() *Config {
	// Try to load .env from common locations
	// 1. Current working directory
	loadEnvFile(".env")
	// 2. Project root (relative to services/api/)
	loadEnvFile("../../.env")

	return &Config{
		Port: getEnv("PORT", "8080"),
		Database: DatabaseConfig{
			Host:            getEnv("DB_HOST", "localhost"),
			Port:            getEnv("DB_PORT", "5432"),
			User:            getEnv("DB_USER", "postgres"),
			Password:        getEnv("DB_PASSWORD", "postgres"),
			Name:            getEnv("DB_NAME", "aidevt"),
			MaxOpenConns:    getEnvInt("DB_MAX_OPEN_CONNS", 25),
			MaxIdleConns:    getEnvInt("DB_MAX_IDLE_CONNS", 5),
			ConnMaxLifetime: getEnvDuration("DB_CONN_MAX_LIFETIME", 5*time.Minute),
			ConnMaxIdleTime: getEnvDuration("DB_CONN_MAX_IDLE_TIME", 1*time.Minute),
		},
		LLM: LLMConfig{
			Provider: getEnv("LLM_PROVIDER", "openai"),
			APIKey:   getEnv("LLM_API_KEY", ""),
			Model:    getEnv("LLM_MODEL", "gpt-4"),
			BaseURL:  getEnv("LLM_BASE_URL", ""),
			Timeout:  getEnvDuration("LLM_TIMEOUT", 60*time.Second),
		},
		Tool: ToolConfig{
			WorkspaceRoot:     getEnv("WORKSPACE_ROOT", ""),
			ReadMaxBytes:      getEnvInt64("TOOL_READ_MAX_BYTES", 1048576),
			SearchMaxResults:  getEnvInt("TOOL_SEARCH_MAX_RESULTS", 50),
			MaxToolIterations: getEnvInt("TOOL_MAX_ITERATIONS", 10),
			RequireApproval:   getEnv("TOOL_REQUIRE_APPROVAL", ""),
			ToolChoice:        getEnv("TOOL_CHOICE", "auto"),
			WriteMaxBytes:     getEnvInt64("TOOL_WRITE_MAX_BYTES", 1048576), // 1MB
			BashTimeout:       getEnvInt("TOOL_BASH_TIMEOUT", 30),
			BashBlocked:       getEnv("TOOL_BASH_BLOCKED_COMMANDS", ""),
		},
	}
}

func getEnvInt64(key string, defaultValue int64) int64 {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.ParseInt(value, 10, 64); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func getEnvDuration(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if durationValue, err := time.ParseDuration(value); err == nil {
			return durationValue
		}
	}
	return defaultValue
}
