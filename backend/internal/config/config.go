package config

import (
	"log"
	"os"
)

// Config captures runtime settings needed by the API server.
type Config struct {
	HTTPAddr           string
	FrontendURL        string
	AuthCallbackURL    string
	GoogleClientID     string
	GoogleClientSecret string
	SessionSecret      string
}

// Load reads configuration from environment variables. Defaults are provided
// for local development so the server can start without extensive setup.
func Load() Config {
	cfg := Config{
		HTTPAddr:           getEnv("HTTP_ADDR", ":3000"),
		FrontendURL:        getEnv("FRONTEND_URL", "http://localhost:5173"),
		AuthCallbackURL:    getEnv("AUTH_CALLBACK_URL", "http://localhost:3000/auth/google/callback"),
		GoogleClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
		GoogleClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
		SessionSecret:      getEnv("SESSION_SECRET", "dev-session-secret"),
	}

	if cfg.GoogleClientID == "" || cfg.GoogleClientSecret == "" {
		log.Println("warning: GOOGLE_CLIENT_ID or GOOGLE_CLIENT_SECRET unset; auth will fail against real providers")
	}
	if cfg.SessionSecret == "dev-session-secret" {
		log.Println("warning: SESSION_SECRET not set, using insecure default")
	}

	return cfg
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
