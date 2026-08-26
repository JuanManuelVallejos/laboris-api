package usecase

import (
	"context"

	"github.com/laboris/laboris-api/internal/domain"
	"github.com/laboris/laboris-api/internal/geocoding"
)

type OnboardingInput struct {
	ClerkID     string
	Email       string
	FullName    string
	Role        string
	Trade       string
	HomeAddress string
	RadiusKm    int
	Bio         string
}

type OnboardingResult struct {
	UserID string
	Role   string
}

type OnboardingUseCase struct {
	users         domain.UserRepository
	professionals domain.ProfessionalRepository
	geo           *geocoding.Client
}

func NewOnboardingUseCase(u domain.UserRepository, p domain.ProfessionalRepository, geo *geocoding.Client) *OnboardingUseCase {
	return &OnboardingUseCase{users: u, professionals: p, geo: geo}
}

func (uc *OnboardingUseCase) Execute(in OnboardingInput) (*OnboardingResult, error) {
	existing, err := uc.users.FindByClerkID(in.ClerkID)
	if err != nil {
		return nil, err
	}

	var user *domain.User
	if existing != nil {
		user = existing
	} else {
		user, err = uc.users.Create(&domain.User{
			ClerkID:  in.ClerkID,
			Email:    in.Email,
			FullName: in.FullName,
		})
		if err != nil {
			return nil, err
		}
	}

	if err := uc.users.AddRole(user.ID, in.Role); err != nil {
		return nil, err
	}

	lat, lng, err := uc.geo.Geocode(context.Background(), in.HomeAddress)
	if err != nil {
		return nil, err
	}

	if in.Role == "professional" {
		radiusKm := in.RadiusKm
		_, err = uc.professionals.Create(&domain.Professional{
			UserID:      user.ID,
			Trade:       in.Trade,
			HomeAddress: in.HomeAddress,
			HomeLat:     &lat,
			HomeLng:     &lng,
			RadiusKm:    &radiusKm,
			Bio:         in.Bio,
		})
		if err != nil {
			return nil, err
		}
	} else {
		if err := uc.users.UpdateHomeAddress(user.ID, in.HomeAddress, lat, lng); err != nil {
			return nil, err
		}
	}

	return &OnboardingResult{UserID: user.ID, Role: in.Role}, nil
}
