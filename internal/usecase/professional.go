package usecase

import "github.com/laboris/laboris-api/internal/domain"

type ProfessionalUseCase interface {
	GetAll() ([]domain.Professional, error)
	GetByID(id string) (*domain.Professional, error)
}

type professionalUseCase struct {
	repo domain.ProfessionalRepository
}

func NewProfessionalUseCase(repo domain.ProfessionalRepository) ProfessionalUseCase {
	return &professionalUseCase{repo: repo}
}

func (uc *professionalUseCase) GetAll() ([]domain.Professional, error) {
	return uc.repo.FindAll()
}

func (uc *professionalUseCase) GetByID(id string) (*domain.Professional, error) {
	return uc.repo.FindByID(id)
}
