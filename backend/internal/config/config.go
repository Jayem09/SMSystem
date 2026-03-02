package config

import (
	"log"
	"os"
	"path/filepath"

	"github.com/joho/godotenv"
)

// Config holds all configuration for the application.
type Config struct {
	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string
	ServerPort string
	JWTSecret  string
	JWTExpiry  string
}

// Load reads configuration from .env file and environment variables.
func Load() *Config {
	// Try loading .env from:
	// 1. Current working directory
	// 2. Parent directory
	// 3. Executable directory (robust for bundled apps like Tauri)
	paths := []string{".env", "../.env"}

	if execPath, err := os.Executable(); err == nil {
		paths = append(paths, filepath.Join(filepath.Dir(execPath), ".env"))
		paths = append(paths, filepath.Join(filepath.Dir(execPath), "../Resources/.env"))
	}

	for _, path := range paths {
		if err := godotenv.Load(path); err == nil {
			log.Printf("Loaded configuration from %s", path)
			break
		}
	}

	return &Config{
		DBHost:     getEnv("DB_HOST", "localhost"),
		DBPort:     getEnv("DB_PORT", "3306"),
		DBUser:     getEnv("DB_USER", "smsystem"),
		DBPassword: getEnv("DB_PASSWORD", "smsystem_secret"),
		DBName:     getEnv("DB_NAME", "smsystem_db"),
		ServerPort: getEnv("SERVER_PORT", "8080"),
		JWTSecret:  getEnv("JWT_SECRET", "default-secret-change-me"),
		JWTExpiry:  getEnv("JWT_EXPIRY", "24h"),
	}
}

func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}
