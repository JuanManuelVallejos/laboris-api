package usecase

import (
	"errors"

	"github.com/laboris/laboris-api/internal/domain"
	"github.com/laboris/laboris-api/internal/realtime"
)

type NotificationUseCase struct {
	repo  domain.NotificationRepository
	users domain.UserRepository
	hub   *realtime.Hub[*domain.Notification]
}

func NewNotificationUseCase(repo domain.NotificationRepository, users domain.UserRepository) *NotificationUseCase {
	return &NotificationUseCase{repo: repo, users: users}
}

func (uc *NotificationUseCase) SetHub(h *realtime.Hub[*domain.Notification]) {
	uc.hub = h
}

func (uc *NotificationUseCase) CreateForUser(userID, notifType, message, entityID string) error {
	n, err := uc.repo.Create(&domain.Notification{
		UserID:   userID,
		Type:     notifType,
		Message:  message,
		EntityID: entityID,
	})
	if err != nil {
		return err
	}
	if uc.hub != nil {
		uc.hub.Publish(userID, n)
	}
	return nil
}

func (uc *NotificationUseCase) Subscribe(clerkID string) (<-chan *domain.Notification, func(), error) {
	user, err := uc.users.FindByClerkID(clerkID)
	if err != nil || user == nil {
		return nil, nil, errors.New("user not found")
	}
	if uc.hub == nil {
		return nil, nil, errors.New("realtime not configured")
	}
	ch, unsub := uc.hub.Subscribe(user.ID)
	return ch, unsub, nil
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
