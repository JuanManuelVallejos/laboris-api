package config

import (
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Port           string
	ClerkSecretKey string
	DatabaseURL    string
}

func Load() *Config {
	_ = godotenv.Load()

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	return &Config{
		Port:           port,
		ClerkSecretKey: os.Getenv("CLERK_SECRET_KEY"),
		DatabaseURL:    os.Getenv("DATABASE_URL"),
	}
}
