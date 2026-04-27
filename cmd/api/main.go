package main

import (
	"context"
	"log"

	"github.com/clerk/clerk-sdk-go/v2"
	"github.com/jackc/pgx/v5/pgxpool"
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
	var mh *handler.MeHandler
	var rh *handler.RequestHandler
	var nh *handler.NotificationHandler
	var ah *handler.AdminHandler
	var jh *handler.JobHandler
	var pool *pgxpool.Pool

	if cfg.DatabaseURL != "" {
		if err := db.RunMigrations(cfg.DatabaseURL); err != nil {
			log.Fatalf("migrations failed: %v", err)
		}

		var err error
		pool, err = db.NewPool(context.Background(), cfg.DatabaseURL)
		if err != nil {
			log.Fatalf("db connection failed: %v", err)
		}
		defer pool.Close()

		profRepo := repopostgres.NewProfessionalRepository(pool)
		userRepo := repopostgres.NewUserRepository(pool)
		reqRepo := repopostgres.NewRequestRepository(pool)
		notifRepo := repopostgres.NewNotificationRepository(pool)
		jobRepo := repopostgres.NewJobRepository(pool)
		msgRepo := repopostgres.NewMessageRepository(pool)
		payRepo := repopostgres.NewPaymentRepository(pool)

		notifUC := usecase.NewNotificationUseCase(notifRepo, userRepo)

		reqUC := usecase.NewRequestUseCase(reqRepo, userRepo, profRepo)
		reqUC.SetNotifications(notifUC)
		reqUC.SetJobRepository(jobRepo)

		jobUC := usecase.NewJobUseCase(jobRepo, payRepo, userRepo, profRepo)
		jobUC.SetNotifications(notifUC)

		msgUC := usecase.NewMessageUseCase(msgRepo, reqRepo, userRepo, profRepo)

		ph = handler.NewProfessionalHandler(usecase.NewProfessionalUseCase(profRepo))
		oh = handler.NewOnboardingHandler(usecase.NewOnboardingUseCase(userRepo, profRepo))
		mh = handler.NewMeHandler(usecase.NewMeUseCase(userRepo, profRepo))
		rh = handler.NewRequestHandler(reqUC)
		nh = handler.NewNotificationHandler(notifUC)
		ah = handler.NewAdminHandler(usecase.NewAdminUseCase(userRepo, profRepo))
		jh = handler.NewJobHandler(jobUC, msgUC)

		log.Println("using PostgreSQL")
	} else {
		profRepo := repomemory.NewProfessionalRepository()
		ph = handler.NewProfessionalHandler(usecase.NewProfessionalUseCase(profRepo))
		oh = handler.NewOnboardingHandler(usecase.NewOnboardingUseCase(nil, nil))
		mh = handler.NewMeHandler(usecase.NewMeUseCase(nil, profRepo))
		rh = handler.NewRequestHandler(usecase.NewRequestUseCase(nil, nil, profRepo))

		log.Println("using in-memory repository (no DATABASE_URL set)")
	}

	r := handler.NewRouter(ph, oh, mh, rh, nh, ah, jh, pool)

	log.Printf("starting laboris-api on :%s", cfg.Port)
	if err := r.Run(":" + cfg.Port); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
