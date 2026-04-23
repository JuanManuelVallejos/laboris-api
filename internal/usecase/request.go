package usecase

import (
	"errors"

	"github.com/laboris/laboris-api/internal/domain"
)

type RequestUseCase struct {
	requests      domain.RequestRepository
	users         domain.UserRepository
	professionals domain.ProfessionalRepository
}

func NewRequestUseCase(requests domain.RequestRepository, users domain.UserRepository, professionals domain.ProfessionalRepository) *RequestUseCase {
	return &RequestUseCase{requests: requests, users: users, professionals: professionals}
}

func (uc *RequestUseCase) Create(clerkID, professionalID, description string) (*domain.Request, error) {
	user, err := uc.users.FindByClerkID(clerkID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, errors.New("user not found")
	}
	return uc.requests.Create(&domain.Request{
		ClientID:       user.ID,
		ProfessionalID: professionalID,
		Description:    description,
	})
}

func (uc *RequestUseCase) ListReceivedByProfessional(clerkID string) ([]domain.Request, error) {
	user, err := uc.users.FindByClerkID(clerkID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, errors.New("user not found")
	}
	prof, err := uc.professionals.FindByUserID(user.ID)
	if err != nil {
		return nil, err
	}
	if prof == nil {
		return nil, errors.New("professional profile not found")
	}
	return uc.requests.FindByProfessionalID(prof.ID)
}

func (uc *RequestUseCase) ListSentByClient(clerkID string) ([]domain.Request, error) {
	user, err := uc.users.FindByClerkID(clerkID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, errors.New("user not found")
	}
	return uc.requests.FindByClientID(user.ID)
}

func (uc *RequestUseCase) UpdateStatus(id, status string) (*domain.Request, error) {
	if status != "accepted" && status != "rejected" {
		return nil, errors.New("invalid status")
	}
	return uc.requests.UpdateStatus(id, status)
}
