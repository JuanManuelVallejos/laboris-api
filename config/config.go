package config

import (
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	Port                   string
	ClerkSecretKey         string
	DatabaseURL            string
	JobAutoCloseDays       int
	SupabaseURL            string
	SupabaseServiceRoleKey string
	SupabaseStorageBucket  string
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

	storageBucket := os.Getenv("SUPABASE_STORAGE_BUCKET")
	if storageBucket == "" {
		storageBucket = "attachments"
	}

	return &Config{
		Port:                   port,
		ClerkSecretKey:         os.Getenv("CLERK_SECRET_KEY"),
		DatabaseURL:            os.Getenv("DATABASE_URL"),
		JobAutoCloseDays:       autoCloseDays,
		SupabaseURL:            os.Getenv("SUPABASE_URL"),
		SupabaseServiceRoleKey: os.Getenv("SUPABASE_SERVICE_ROLE_KEY"),
		SupabaseStorageBucket:  storageBucket,
	}
}
