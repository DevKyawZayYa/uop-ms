package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Port     string
	MySQLDSN string
}

func Load() *Config {
	_ = godotenv.Load("../../.env")
	_ = godotenv.Load()

	return &Config{
		Port:     getEnv("ORDER_SERVICE_PORT", "8082"),
		MySQLDSN: mustEnv("ORDER_MYSQL_DSN"),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func mustEnv(key string) string {
	v := os.Getenv(key)
	if v == "" {
		log.Fatalf("missing required env: %s", key)
	}
	return v
}
