package usecase

import (
	"errors"

	"github.com/laboris/laboris-api/internal/domain"
	"github.com/laboris/laboris-api/internal/storage"
)

type MeUseCase struct {
	users         domain.UserRepository
	professionals domain.ProfessionalRepository
	storage       *storage.SupabaseClient
}

func NewMeUseCase(users domain.UserRepository, professionals domain.ProfessionalRepository, storageClient *storage.SupabaseClient) *MeUseCase {
	return &MeUseCase{users: users, professionals: professionals, storage: storageClient}
}

func (uc *MeUseCase) GetMyProfessional(clerkID string) (*domain.Professional, error) {
	user, err := uc.users.FindByClerkID(clerkID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, errors.New("user not found")
	}
	prof, err := uc.professionals.FindByUserID(user.ID)
	if err != nil || prof == nil {
		return prof, err
	}
	signPortfolioPhotos(uc.storage, prof)
	return prof, nil
}

func (uc *MeUseCase) UpdateMyProfessional(clerkID, trade, zone, bio string) (*domain.Professional, error) {
	user, err := uc.users.FindByClerkID(clerkID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, errors.New("user not found")
	}
	return uc.professionals.UpdateByUserID(user.ID, trade, zone, bio)
}
