package config

import (
	"fmt"
	"os"
)

func Validate(cfg *Config) error {
	if cfg.JWTSecret == "" {
		return fmt.Errorf("JWT_SECRET environment variable is required")
	}

	if len(cfg.JWTSecret) < 32 {
		return fmt.Errorf("JWT_SECRET must be at least 32 characters for security")
	}

	if cfg.DBHost == "" {
		return fmt.Errorf("DB_HOST environment variable is required")
	}

	return nil
}

func MustValidate(cfg *Config) {
	if err := Validate(cfg); err != nil {
		fmt.Fprintf(os.Stderr, "Configuration validation failed: %v\n", err)
		os.Exit(1)
	}
}
