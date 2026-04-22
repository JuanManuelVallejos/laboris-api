package usecase

import "github.com/laboris/laboris-api/internal/domain"

type OnboardingInput struct {
	ClerkID  string
	Email    string
	FullName string
	Role     string
	Trade    string
	Zone     string
	Bio      string
}

type OnboardingResult struct {
	UserID string
	Role   string
}

type OnboardingUseCase struct {
	users         domain.UserRepository
	professionals domain.ProfessionalRepository
}

func NewOnboardingUseCase(u domain.UserRepository, p domain.ProfessionalRepository) *OnboardingUseCase {
	return &OnboardingUseCase{users: u, professionals: p}
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

	if in.Role == "professional" {
		_, err = uc.professionals.Create(&domain.Professional{
			UserID: user.ID,
			Trade:  in.Trade,
			Zone:   in.Zone,
			Bio:    in.Bio,
		})
		if err != nil {
			return nil, err
		}
	}

	return &OnboardingResult{UserID: user.ID, Role: in.Role}, nil
}
