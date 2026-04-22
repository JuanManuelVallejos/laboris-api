package main

import (
	"context"
	"log"

	"github.com/clerk/clerk-sdk-go/v2"
	"github.com/laboris/laboris-api/config"
	"github.com/laboris/laboris-api/internal/db"
	"github.com/laboris/laboris-api/internal/handler"
	repomemory "github.com/laboris/laboris-api/internal/repository/memory"
	repopostgres "github.com/laboris/laboris-api/internal/repository/postgres"
	"github.com/laboris/laboris-api/internal/usecase"
)

func main() {
	cfg := config.Load()

	clerk.SetKey(cfg.ClerkSecretKey)

	var ph *handler.ProfessionalHandler
	var oh *handler.OnboardingHandler

	if cfg.DatabaseURL != "" {
		if err := db.RunMigrations(cfg.DatabaseURL); err != nil {
			log.Fatalf("migrations failed: %v", err)
		}

		pool, err := db.NewPool(context.Background(), cfg.DatabaseURL)
		if err != nil {
			log.Fatalf("db connection failed: %v", err)
		}
		defer pool.Close()

		profRepo := repopostgres.NewProfessionalRepository(pool)
		userRepo := repopostgres.NewUserRepository(pool)

		ph = handler.NewProfessionalHandler(usecase.NewProfessionalUseCase(profRepo))
		oh = handler.NewOnboardingHandler(usecase.NewOnboardingUseCase(userRepo, profRepo))

		log.Println("using PostgreSQL")
	} else {
		profRepo := repomemory.NewProfessionalRepository()
		ph = handler.NewProfessionalHandler(usecase.NewProfessionalUseCase(profRepo))
		oh = handler.NewOnboardingHandler(usecase.NewOnboardingUseCase(nil, nil))

		log.Println("using in-memory repository (no DATABASE_URL set)")
	}

	r := handler.NewRouter(ph, oh)

	log.Printf("starting laboris-api on :%s", cfg.Port)
	if err := r.Run(":" + cfg.Port); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
