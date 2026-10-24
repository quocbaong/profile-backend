package config

import (
	"os"
	"time"
)

type Config struct {
	Port          string
	AppEnv        string
	DBDriver      string // "postgres" or "sqlite"
	DBDSN         string
	JWTSecret     string
	JWTExpire     time.Duration
	CORSOrigin    string
	AdminUsername string
	AdminPassword string
}

func LoadConfig() *Config {
	return &Config{
		Port:          getEnv("PORT", "8080"),
		AppEnv:        getEnv("APP_ENV", "development"),
		DBDriver:      getEnv("DB_DRIVER", "sqlite"), // default sqlite for zero-config local run
		DBDSN:         getEnv("DB_DSN", "portfolio.db"),
		JWTSecret:     getEnv("JWT_SECRET", "super-secret-jwt-key-change-in-production-2026"),
		JWTExpire:     time.Hour * 24 * 7, // 7 days
		CORSOrigin:    getEnv("CORS_ORIGIN", "http://localhost:3000,https://quocbao.click"),
		AdminUsername: getEnv("ADMIN_USERNAME", "admin"),
		AdminPassword: getEnv("ADMIN_PASSWORD", "Admin@123456"),
	}
}

func getEnv(key string, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}
