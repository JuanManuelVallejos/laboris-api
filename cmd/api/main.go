package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/clerk/clerk-sdk-go/v2"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/laboris/laboris-api/config"
	"github.com/laboris/laboris-api/internal/db"
	"github.com/laboris/laboris-api/internal/domain"
	"github.com/laboris/laboris-api/internal/handler"
	"github.com/laboris/laboris-api/internal/realtime"
	repomemory "github.com/laboris/laboris-api/internal/repository/memory"
	repopostgres "github.com/laboris/laboris-api/internal/repository/postgres"
	"github.com/laboris/laboris-api/internal/storage"
	"github.com/laboris/laboris-api/internal/usecase"
)

func main() {
	cfg := config.Load()

	clerk.SetKey(cfg.ClerkSecretKey)

	storageClient := storage.NewSupabaseClient(cfg.SupabaseURL, cfg.SupabaseServiceRoleKey, cfg.SupabaseStorageBucket)
	sseConnLimiter := realtime.NewConnLimiter(6) // tope de streams SSE concurrentes por usuario

	var ph *handler.ProfessionalHandler
	var oh *handler.OnboardingHandler
	var mh *handler.MeHandler
	var rh *handler.RequestHandler
	var nh *handler.NotificationHandler
	var ah *handler.AdminHandler
	var jh *handler.JobHandler
	var atth *handler.AttachmentHandler
	var wh *handler.ClerkWebhookHandler
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
		reworkRepo := repopostgres.NewReworkRecordRepository(pool)
		attachmentRepo := repopostgres.NewAttachmentRepository(pool)

		messageHub := realtime.NewHub[*domain.Message]()
		notificationHub := realtime.NewHub[*domain.Notification]()
		jobHub := realtime.NewHub[*domain.Job]()

		notifUC := usecase.NewNotificationUseCase(notifRepo, userRepo)
		notifUC.SetHub(notificationHub)

		reqUC := usecase.NewRequestUseCase(reqRepo, userRepo, profRepo)
		reqUC.SetNotifications(notifUC)
		reqUC.SetJobRepository(jobRepo)
		reqUC.SetStorage(storageClient)
		reqUC.SetAutoCloseDays(cfg.JobAutoCloseDays)

		jobUC := usecase.NewJobUseCase(jobRepo, payRepo, userRepo, profRepo, reworkRepo)
		jobUC.SetNotifications(notifUC)
		jobUC.SetHub(jobHub)
		jobUC.SetAutoCloseDays(cfg.JobAutoCloseDays)

		msgUC := usecase.NewMessageUseCase(msgRepo, reqRepo, userRepo, profRepo)
		msgUC.SetNotifications(notifUC)
		msgUC.SetHub(messageHub)

		attachmentUC := usecase.NewAttachmentUseCase(attachmentRepo, userRepo, profRepo, reqRepo, storageClient)

		ph = handler.NewProfessionalHandler(usecase.NewProfessionalUseCase(profRepo, storageClient))
		oh = handler.NewOnboardingHandler(usecase.NewOnboardingUseCase(userRepo, profRepo))
		mh = handler.NewMeHandler(usecase.NewMeUseCase(userRepo, profRepo, storageClient))
		rh = handler.NewRequestHandler(reqUC)
		nh = handler.NewNotificationHandler(notifUC, sseConnLimiter)
		ah = handler.NewAdminHandler(usecase.NewAdminUseCase(userRepo, profRepo))
		jh = handler.NewJobHandler(jobUC, msgUC, sseConnLimiter)
		atth = handler.NewAttachmentHandler(attachmentUC)
		if cfg.ClerkWebhookSecret != "" {
			wh = handler.NewClerkWebhookHandler(usecase.NewUserSyncUseCase(userRepo), cfg.ClerkWebhookSecret)
		}

		log.Println("using PostgreSQL")
	} else {
		profRepo := repomemory.NewProfessionalRepository()
		ph = handler.NewProfessionalHandler(usecase.NewProfessionalUseCase(profRepo, storageClient))
		oh = handler.NewOnboardingHandler(usecase.NewOnboardingUseCase(nil, nil))
		mh = handler.NewMeHandler(usecase.NewMeUseCase(nil, profRepo, storageClient))
		rh = handler.NewRequestHandler(usecase.NewRequestUseCase(nil, nil, profRepo))

		log.Println("using in-memory repository (no DATABASE_URL set)")
	}

	r := handler.NewRouter(ph, oh, mh, rh, nh, ah, jh, atth, wh, pool)

	// WriteTimeout deliberately left unset: it's the only timeout that would
	// cut off long-lived SSE streams (job/message/notification updates),
	// which need to stay open indefinitely by design. ReadHeaderTimeout and
	// ReadTimeout guard against slow/malicious request writers (slowloris)
	// without affecting how long the server may take to respond.
	srv := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           r,
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       30 * time.Second,
		IdleTimeout:       120 * time.Second,
	}

	log.Printf("starting laboris-api on :%s", cfg.Port)
	if err := srv.ListenAndServe(); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
