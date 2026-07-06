package config

import (
	"fmt"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

// AppConfig holds global application settings.
type AppConfig struct {
	Env             string
	Port            string
	UploadDir       string
	MaxUploadSizeMB int64
	BaseURL         string
}

// JWTConfig holds JWT settings.
type JWTConfig struct {
	Secret      string
	ExpireHours int
}

// Config is the root config container.
type Config struct {
	App AppConfig
	DB  DBConfig
	JWT JWTConfig
}

var cfg *Config

// Load reads .env and initialises the global Config singleton.
func Load() *Config {
	_ = godotenv.Load() // non-fatal if missing (env vars may already be set)

	cfg = &Config{
		App: AppConfig{
			Env:             getEnv("APP_ENV", "development"),
			Port:            getEnv("APP_PORT", "8080"),
			UploadDir:       getEnv("UPLOAD_DIR", "./uploads"),
			MaxUploadSizeMB: int64(getEnvInt("MAX_UPLOAD_SIZE_MB", 10)),
			BaseURL:         getEnv("BASE_URL", "http://localhost:8080"),
		},
		DB:  loadDB(),
		JWT: loadJWT(),
	}
	return cfg
}

// Get returns the already-loaded Config singleton.
// Panics if Load() has not been called first.
func Get() *Config {
	if cfg == nil {
		panic("config.Load() must be called before config.Get()")
	}
	return cfg
}

// ── helpers ──────────────────────────────────────────────────────────────────

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func mustGetEnv(key string) string {
	v := os.Getenv(key)
	if v == "" {
		panic(fmt.Sprintf("required env var %q is not set", key))
	}
	return v
}

func getEnvInt(key string, fallback int) int {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return fallback
}
