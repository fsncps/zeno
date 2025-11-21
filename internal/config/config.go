package config

import (
	"os"
)

type Config struct {
	DBUser string
	DBPass string
	DBHost string
	DBPort string
	DBName string
}

func Load() Config {
	return Config{
		DBUser: getenv("ZENODB_USER", "zeno_user"),
		DBPass: getenv("ZENODB_PASS", "secret"),
		DBHost: getenv("ZENODB_HOST", "127.0.0.1"),
		DBPort: getenv("ZENODB_PORT", "3306"),
		DBName: getenv("ZENODB_NAME", "zeno"),
	}
}

func getenv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
