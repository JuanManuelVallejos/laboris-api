package usecase

import (
	"errors"

	"github.com/laboris/laboris-api/internal/domain"
)

type NotificationUseCase struct {
	repo  domain.NotificationRepository
	users domain.UserRepository
}

func NewNotificationUseCase(repo domain.NotificationRepository, users domain.UserRepository) *NotificationUseCase {
	return &NotificationUseCase{repo: repo, users: users}
}

func (uc *NotificationUseCase) CreateForUser(userID, notifType, message string) error {
	_, err := uc.repo.Create(&domain.Notification{
		UserID:  userID,
		Type:    notifType,
		Message: message,
	})
	return err
}

func (uc *NotificationUseCase) ListForUser(clerkID string) ([]domain.Notification, error) {
	user, err := uc.users.FindByClerkID(clerkID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, errors.New("user not found")
	}
	return uc.repo.FindByUserID(user.ID)
}

func (uc *NotificationUseCase) CountUnread(clerkID string) (int, error) {
	user, err := uc.users.FindByClerkID(clerkID)
	if err != nil {
		return 0, err
	}
	if user == nil {
		return 0, errors.New("user not found")
	}
	return uc.repo.CountUnread(user.ID)
}

func (uc *NotificationUseCase) MarkAllRead(clerkID string) error {
	user, err := uc.users.FindByClerkID(clerkID)
	if err != nil {
		return err
	}
	if user == nil {
		return errors.New("user not found")
	}
	return uc.repo.MarkAllRead(user.ID)
}
