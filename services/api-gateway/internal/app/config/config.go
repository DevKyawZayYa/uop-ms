package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Port              string
	ProductServiceURL string
	OrderServiceURL   string
}

func Load() *Config {
	_ = godotenv.Load("../../.env")
	_ = godotenv.Load()

	cfg := &Config{
		Port:              getEnv("GATEWAY_PORT", "8080"),
		ProductServiceURL: getEnv("PRODUCT_SERVICE_URL", "http://localhost:8081"),
		OrderServiceURL:   getEnv("ORDER_SERVICE_URL", "http://localhost:8082"),
	}

	if cfg.ProductServiceURL == "" || cfg.OrderServiceURL == "" {
		log.Fatal("missing downstream service URLs")
	}

	return cfg
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
