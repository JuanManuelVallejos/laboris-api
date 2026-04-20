package main

import (
	"log"

	"github.com/laboris/laboris-api/config"
	"github.com/laboris/laboris-api/internal/handler"
	"github.com/laboris/laboris-api/internal/repository/memory"
	"github.com/laboris/laboris-api/internal/usecase"
)

func main() {
	cfg := config.Load()

	repo := memory.NewProfessionalRepository()
	uc := usecase.NewProfessionalUseCase(repo)
	ph := handler.NewProfessionalHandler(uc)

	r := handler.NewRouter(ph)

	log.Printf("starting laboris-api on :%s", cfg.Port)
	if err := r.Run(":" + cfg.Port); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
