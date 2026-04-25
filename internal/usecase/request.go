package usecase

import (
	"errors"
	"fmt"

	"github.com/laboris/laboris-api/internal/domain"
)

type RequestUseCase struct {
	requests      domain.RequestRepository
	users         domain.UserRepository
	professionals domain.ProfessionalRepository
	notifications *NotificationUseCase
}

func NewRequestUseCase(requests domain.RequestRepository, users domain.UserRepository, professionals domain.ProfessionalRepository) *RequestUseCase {
	return &RequestUseCase{requests: requests, users: users, professionals: professionals}
}

func (uc *RequestUseCase) SetNotifications(n *NotificationUseCase) {
	uc.notifications = n
}

func (uc *RequestUseCase) Create(clerkID, professionalID, description string) (*domain.Request, error) {
	user, err := uc.users.FindByClerkID(clerkID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, errors.New("user not found")
	}
	req, err := uc.requests.Create(&domain.Request{
		ClientID:       user.ID,
		ProfessionalID: professionalID,
		Description:    description,
	})
	if err != nil {
		return nil, err
	}

	if uc.notifications != nil && uc.professionals != nil {
		prof, lookupErr := uc.professionals.FindByID(professionalID)
		if lookupErr == nil && prof != nil {
			msg := fmt.Sprintf("Nueva solicitud de %s", user.FullName)
			_ = uc.notifications.CreateForUser(prof.UserID, "new_request", msg)
		}
	}

	return req, nil
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

func (uc *RequestUseCase) UpdateStatus(id, status, reason string) (*domain.Request, error) {
	if status != "accepted" && status != "rejected" {
		return nil, errors.New("invalid status")
	}
	if status == "rejected" && reason == "" {
		return nil, errors.New("rejection reason is required")
	}
	rq, err := uc.requests.UpdateStatus(id, status, reason)
	if err != nil {
		return nil, err
	}

	if uc.notifications != nil && uc.professionals != nil {
		prof, lookupErr := uc.professionals.FindByID(rq.ProfessionalID)
		if lookupErr == nil && prof != nil {
			var msg string
			if status == "accepted" {
				msg = fmt.Sprintf("%s aceptó tu solicitud", prof.Name)
			} else {
				msg = fmt.Sprintf("%s rechazó tu solicitud", prof.Name)
			}
			_ = uc.notifications.CreateForUser(rq.ClientID, "request_"+status, msg)
		}
	}

	return rq, nil
}
