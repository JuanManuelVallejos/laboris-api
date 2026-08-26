package usecase

import (
	"context"
	"errors"

	"github.com/laboris/laboris-api/internal/domain"
	"github.com/laboris/laboris-api/internal/geocoding"
	"github.com/laboris/laboris-api/internal/storage"
)

type MeUseCase struct {
	users         domain.UserRepository
	professionals domain.ProfessionalRepository
	storage       *storage.SupabaseClient
	geo           *geocoding.Client
}

func NewMeUseCase(users domain.UserRepository, professionals domain.ProfessionalRepository, storageClient *storage.SupabaseClient, geo *geocoding.Client) *MeUseCase {
	return &MeUseCase{users: users, professionals: professionals, storage: storageClient, geo: geo}
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
	signAttachments(uc.storage, prof.PortfolioPhotos)
	return prof, nil
}

func (uc *MeUseCase) UpdateMyProfessional(clerkID, trade, homeAddress, bio string, radiusKm int) (*domain.Professional, error) {
	user, err := uc.users.FindByClerkID(clerkID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, errors.New("user not found")
	}
	lat, lng, err := uc.geo.Geocode(context.Background(), homeAddress)
	if err != nil {
		return nil, err
	}
	return uc.professionals.UpdateByUserID(user.ID, trade, homeAddress, bio, lat, lng, radiusKm)
}

// UpdateMyAddress actualiza el domicilio del cliente (no del profesional —
// ver UpdateMyProfessional para eso), usado desde Perfil.
func (uc *MeUseCase) UpdateMyAddress(clerkID, address string) error {
	user, err := uc.users.FindByClerkID(clerkID)
	if err != nil {
		return err
	}
	if user == nil {
		return ErrUserNotOnboarded
	}
	lat, lng, err := uc.geo.Geocode(context.Background(), address)
	if err != nil {
		return err
	}
	return uc.users.UpdateHomeAddress(user.ID, address, lat, lng)
}
