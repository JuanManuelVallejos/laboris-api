package usecase

import (
	"errors"

	"github.com/laboris/laboris-api/internal/domain"
)

type MeUseCase struct {
	users         domain.UserRepository
	professionals domain.ProfessionalRepository
}

func NewMeUseCase(users domain.UserRepository, professionals domain.ProfessionalRepository) *MeUseCase {
	return &MeUseCase{users: users, professionals: professionals}
}

func (uc *MeUseCase) GetMyProfessional(clerkID string) (*domain.Professional, error) {
	user, err := uc.users.FindByClerkID(clerkID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, errors.New("user not found")
	}
	return uc.professionals.FindByUserID(user.ID)
}
