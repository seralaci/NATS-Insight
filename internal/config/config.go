package config

import (
	"log/slog"
	"os"
)

// Config holds the application configuration.
type Config struct {
	ListenAddr   string
	NATSUrl      string
	MonitorUrl   string
	DataDir      string
	NATSUsername string
	NATSPassword string
	NATSToken    string
	LogLevel     slog.Level
}

// Load reads configuration from environment variables with sensible defaults.
func Load() *Config {
	return &Config{
		ListenAddr:   getEnv("LISTEN_ADDR", ":8080"),
		NATSUrl:      getEnv("NATS_URL", "nats://localhost:4222"),
		MonitorUrl:   getEnv("NATS_MONITOR_URL", "http://localhost:8222"),
		DataDir:      getEnv("DATA_DIR", ""),
		NATSUsername: getEnv("NATS_USERNAME", ""),
		NATSPassword: getEnv("NATS_PASSWORD", ""),
		NATSToken:    getEnv("NATS_TOKEN", ""),
		LogLevel:     parseLogLevel(getEnv("LOG_LEVEL", "info")),
	}
}

func getEnv(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}

func parseLogLevel(s string) slog.Level {
	switch s {
	case "debug":
		return slog.LevelDebug
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}
