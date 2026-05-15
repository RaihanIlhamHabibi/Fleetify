package config

import (
	"os"
	"strconv"
)

type Config struct {
	DBHost      string
	DBPort      string
	DBUser      string
	DBPassword  string
	DBName      string
	AppPort     string
	UploadDir   string
	WebhookURL  string
	RunSeeder   bool
}

func Load() Config {
	return Config{
		DBHost:     getEnv("DB_HOST", "localhost"),
		DBPort:     getEnv("DB_PORT", "3306"),
		DBUser:     getEnv("DB_USER", "fleetify"),
		DBPassword: getEnv("DB_PASSWORD", "fleetify_secret"),
		DBName:     getEnv("DB_NAME", "fleetify"),
		AppPort:    getEnv("APP_PORT", "8080"),
		UploadDir:  getEnv("UPLOAD_DIR", "./uploads"),
		WebhookURL: getEnv("WEBHOOK_URL", ""),
		RunSeeder:  getEnvBool("RUN_SEEDER", true),
	}
}

func getEnv(key, fallback string) string {
	if v, ok := os.LookupEnv(key); ok {
		return v
	}
	return fallback
}

func getEnvBool(key string, fallback bool) bool {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	b, err := strconv.ParseBool(v)
	if err != nil {
		return fallback
	}
	return b
}
