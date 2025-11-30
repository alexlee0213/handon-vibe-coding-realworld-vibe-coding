package config

import (
	"errors"
	"log/slog"
	"os"
	"time"

	"github.com/joho/godotenv"
)

// Default insecure JWT secret - must be changed in production
const defaultJWTSecret = "your-secret-key-change-in-production"

// ErrInsecureJWTSecret is returned when the default JWT secret is used in production
var ErrInsecureJWTSecret = errors.New("JWT_SECRET must be set to a secure value in production")

type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	JWT      JWTConfig
	CORS     CORSConfig
}

type ServerConfig struct {
	Port string
	Env  string
}

type DatabaseConfig struct {
	URL string
}

type JWTConfig struct {
	Secret string
	Expiry time.Duration
}

type CORSConfig struct {
	AllowedOrigins []string
}

func Load() (*Config, error) {
	// Load .env file if it exists (ignore error if file doesn't exist)
	// This allows environment variables to be set via .env file in development
	// while still supporting direct environment variables in production
	// Try multiple paths: current directory, parent directory (for running from backend/)
	envPaths := []string{".env", "../.env"}
	envLoaded := false
	for _, path := range envPaths {
		if err := godotenv.Load(path); err == nil {
			envLoaded = true
			slog.Debug("loaded .env file", "path", path)
			break
		}
	}
	if !envLoaded {
		slog.Debug("no .env file found, using system environment variables")
	}

	env := getEnv("SERVER_ENV", "development")
	jwtSecret := getEnv("JWT_SECRET", defaultJWTSecret)

	// Validate JWT secret in production
	if env == "production" && jwtSecret == defaultJWTSecret {
		return nil, ErrInsecureJWTSecret
	}

	// Warn if using default secret in development
	if jwtSecret == defaultJWTSecret {
		slog.Warn("using default JWT secret - not suitable for production")
	}

	// Parse CORS allowed origins from environment
	allowedOrigins := parseOrigins(getEnv("CORS_ALLOWED_ORIGINS", ""))

	cfg := &Config{
		Server: ServerConfig{
			Port: getEnv("SERVER_PORT", "8080"),
			Env:  env,
		},
		Database: DatabaseConfig{
			URL: getEnv("DATABASE_URL", "sqlite3://./data/conduit.db"),
		},
		JWT: JWTConfig{
			Secret: jwtSecret,
			Expiry: parseDuration(getEnv("JWT_EXPIRY", "72h")),
		},
		CORS: CORSConfig{
			AllowedOrigins: allowedOrigins,
		},
	}

	return cfg, nil
}

func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

func parseDuration(s string) time.Duration {
	d, err := time.ParseDuration(s)
	if err != nil {
		return 72 * time.Hour
	}
	return d
}

// parseOrigins parses comma-separated CORS origins
func parseOrigins(s string) []string {
	if s == "" {
		return []string{"*"} // Default to all origins for development
	}
	var origins []string
	for _, origin := range splitAndTrim(s, ",") {
		if origin != "" {
			origins = append(origins, origin)
		}
	}
	if len(origins) == 0 {
		return []string{"*"}
	}
	return origins
}

// splitAndTrim splits a string and trims whitespace from each part
func splitAndTrim(s, sep string) []string {
	parts := make([]string, 0)
	for _, part := range split(s, sep) {
		trimmed := trim(part)
		if trimmed != "" {
			parts = append(parts, trimmed)
		}
	}
	return parts
}

// split is a simple string split function
func split(s, sep string) []string {
	if s == "" {
		return nil
	}
	result := make([]string, 0)
	start := 0
	for i := 0; i < len(s); i++ {
		if len(s)-i >= len(sep) && s[i:i+len(sep)] == sep {
			result = append(result, s[start:i])
			start = i + len(sep)
		}
	}
	result = append(result, s[start:])
	return result
}

// trim removes leading and trailing whitespace
func trim(s string) string {
	start := 0
	end := len(s)
	for start < end && (s[start] == ' ' || s[start] == '\t' || s[start] == '\n' || s[start] == '\r') {
		start++
	}
	for end > start && (s[end-1] == ' ' || s[end-1] == '\t' || s[end-1] == '\n' || s[end-1] == '\r') {
		end--
	}
	return s[start:end]
}

func (c *Config) IsDevelopment() bool {
	return c.Server.Env == "development"
}

func (c *Config) IsProduction() bool {
	return c.Server.Env == "production"
}
