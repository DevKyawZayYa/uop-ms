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

	AuthMode          string // dev | cognito
	CognitoRegion     string
	CognitoUserPoolID string
	CognitoClientID   string
}

func Load() *Config {
	_ = godotenv.Load("../../.env")
	_ = godotenv.Load()

	cfg := &Config{
		Port:              getEnv("GATEWAY_PORT", "8080"),
		ProductServiceURL: getEnv("PRODUCT_SERVICE_URL", "http://localhost:8081"),
		OrderServiceURL:   getEnv("ORDER_SERVICE_URL", "http://localhost:8082"),

		AuthMode:          getEnv("AUTH_MODE", "dev"),
		CognitoRegion:     getEnv("COGNITO_REGION", ""),
		CognitoUserPoolID: getEnv("COGNITO_USER_POOL_ID", ""),
		CognitoClientID:   getEnv("COGNITO_APP_CLIENT_ID", ""),
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
