package config

import (
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	Port             string
	ClerkSecretKey   string
	DatabaseURL      string
	JobAutoCloseDays int
}

func Load() *Config {
	_ = godotenv.Load()

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	autoCloseDays := 3
	if v := os.Getenv("JOB_AUTO_CLOSE_DAYS"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			autoCloseDays = n
		}
	}

	return &Config{
		Port:             port,
		ClerkSecretKey:   os.Getenv("CLERK_SECRET_KEY"),
		DatabaseURL:      os.Getenv("DATABASE_URL"),
		JobAutoCloseDays: autoCloseDays,
	}
}
