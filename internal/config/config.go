package config

import (
	"os"
	"strings"
	"time"
)

type Config struct {
	Port                string
	AppEnv              string
	DBDriver            string // "postgres" or "sqlite"
	DBDSN               string
	JWTSecret           string
	JWTExpire           time.Duration
	CORSOrigin          string
	AdminUsername       string
	AdminPassword       string
	SupabaseURL         string
	SupabaseKey         string
	SupabaseBucket      string
}

// loadDotEnv: Tự động nạp file .env cục bộ nếu tồn tại (không phụ thuộc thư viện ngoài)
func loadDotEnv() {
	paths := []string{".env", "../.env", "profile-backend/.env"}
	for _, p := range paths {
		data, err := os.ReadFile(p)
		if err == nil {
			lines := strings.Split(string(data), "\n")
			for _, line := range lines {
				line = strings.TrimSpace(line)
				if line == "" || strings.HasPrefix(line, "#") {
					continue
				}
				parts := strings.SplitN(line, "=", 2)
				if len(parts) == 2 {
					key := strings.TrimSpace(parts[0])
					val := strings.TrimSpace(parts[1])
					val = strings.Trim(val, `"'`)
					// Chỉ set nếu môi trường chưa có biến này
					if os.Getenv(key) == "" {
						_ = os.Setenv(key, val)
					}
				}
			}
			break
		}
	}
}

func LoadConfig() *Config {
	loadDotEnv()

	return &Config{
		Port:                getEnv("PORT", "8085"),
		AppEnv:              getEnv("APP_ENV", "development"),
		DBDriver:            getEnv("DB_DRIVER", "sqlite"),
		DBDSN:               getEnv("DB_DSN", "portfolio.db"),
		JWTSecret:           getEnv("JWT_SECRET", "super-secret-jwt-key-change-in-production-2026"),
		JWTExpire:           time.Hour * 24 * 7,
		CORSOrigin:          getEnv("CORS_ORIGIN", "http://localhost:3000,https://quocbao.click"),
		AdminUsername:       getEnv("ADMIN_USERNAME", ""),
		AdminPassword:       getEnv("ADMIN_PASSWORD", ""),
		SupabaseURL:         getEnv("SUPABASE_URL", ""),
		SupabaseKey:         getEnv("SUPABASE_KEY", ""),
		SupabaseBucket:      getEnv("SUPABASE_BUCKET", "portfolio"),
	}
}

func getEnv(key string, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}
