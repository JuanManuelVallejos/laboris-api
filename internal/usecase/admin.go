package usecase

import (
	"errors"

	"github.com/laboris/laboris-api/internal/domain"
)

type AdminUseCase struct {
	users         domain.UserRepository
	professionals domain.ProfessionalRepository
}

func NewAdminUseCase(users domain.UserRepository, professionals domain.ProfessionalRepository) *AdminUseCase {
	return &AdminUseCase{users: users, professionals: professionals}
}

func (uc *AdminUseCase) ListUsers(page, limit int) ([]domain.UserWithRoles, int64, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}
	return uc.users.FindAllPaginated(page, limit)
}

func (uc *AdminUseCase) ListProfessionals(page, limit int) ([]domain.Professional, int64, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}
	return uc.professionals.FindAllPaginated(page, limit)
}

func (uc *AdminUseCase) VerifyProfessional(id string, verified bool) error {
	return uc.professionals.SetVerified(id, verified)
}

func (uc *AdminUseCase) SetProfessionalStatus(id string, status string) error {
	if status != "active" && status != "suspended" {
		return errors.New("invalid status")
	}
	return uc.professionals.SetStatus(id, status)
}

func (uc *AdminUseCase) DeleteProfessional(id string) error {
	return uc.professionals.Delete(id)
}
