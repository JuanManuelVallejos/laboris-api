package usecase

import (
	"github.com/laboris/laboris-api/internal/domain"
	"github.com/laboris/laboris-api/internal/storage"
)

type ProfessionalUseCase interface {
	GetAll() ([]domain.Professional, error)
	GetByID(id string) (*domain.Professional, error)
}

type professionalUseCase struct {
	repo    domain.ProfessionalRepository
	storage *storage.SupabaseClient
}

func NewProfessionalUseCase(repo domain.ProfessionalRepository, storageClient *storage.SupabaseClient) ProfessionalUseCase {
	return &professionalUseCase{repo: repo, storage: storageClient}
}

func (uc *professionalUseCase) GetAll() ([]domain.Professional, error) {
	return uc.repo.FindAll()
}

func (uc *professionalUseCase) GetByID(id string) (*domain.Professional, error) {
	p, err := uc.repo.FindByID(id)
	if err != nil || p == nil {
		return p, err
	}
	signAttachments(uc.storage, p.PortfolioPhotos)
	return p, nil
}
