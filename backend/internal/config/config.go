// Package config loads runtime configuration from the environment.
package config

import (
	"os"
	"strings"
)

// Config holds all runtime settings. Values come from environment variables so the
// same binary works across docker/local without code changes.
type Config struct {
	Port                string
	DatabaseURL         string
	LogLevel            string
	CORSOrigins         []string
	PublicBaseURL       string
	STTProvider         string
	TranslationProvider string
	DefaultTargetLang   string
	OTelEnabled         bool
}

// Load reads configuration from the environment, applying sensible defaults.
func Load() Config {
	return Config{
		Port:                env("PORT", "8080"),
		DatabaseURL:         env("DATABASE_URL", "postgres://meetuser:meetpass@localhost:5432/meetings?sslmode=disable"),
		LogLevel:            env("LOG_LEVEL", "info"),
		CORSOrigins:         splitCSV(env("CORS_ORIGINS", "http://localhost")),
		PublicBaseURL:       env("PUBLIC_BASE_URL", "http://localhost"),
		STTProvider:         env("STT_PROVIDER", "mock"),
		TranslationProvider: env("TRANSLATION_PROVIDER", "mock"),
		DefaultTargetLang:   env("DEFAULT_TARGET_LANG", "ru"),
		OTelEnabled:         env("OTEL_ENABLED", "false") == "true",
	}
}

func env(key, def string) string {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		return v
	}
	return def
}

func splitCSV(v string) []string {
	parts := strings.Split(v, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if t := strings.TrimSpace(p); t != "" {
			out = append(out, t)
		}
	}
	return out
}
