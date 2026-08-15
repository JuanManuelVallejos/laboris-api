package usecase

import (
	"errors"
	"fmt"

	"github.com/laboris/laboris-api/internal/domain"
	"github.com/laboris/laboris-api/internal/realtime"
)

type MessageUseCase struct {
	messages      domain.MessageRepository
	requests      domain.RequestRepository
	users         domain.UserRepository
	professionals domain.ProfessionalRepository
	notifications *NotificationUseCase
	hub           *realtime.Hub[*domain.Message]
}

func NewMessageUseCase(
	messages domain.MessageRepository,
	requests domain.RequestRepository,
	users domain.UserRepository,
	professionals domain.ProfessionalRepository,
) *MessageUseCase {
	return &MessageUseCase{messages: messages, requests: requests, users: users, professionals: professionals}
}

func (uc *MessageUseCase) SetNotifications(n *NotificationUseCase) {
	uc.notifications = n
}

func (uc *MessageUseCase) SetHub(h *realtime.Hub[*domain.Message]) {
	uc.hub = h
}

func (uc *MessageUseCase) Send(clerkID, requestID, content string) (*domain.Message, error) {
	if content == "" {
		return nil, errors.New("message content cannot be empty")
	}
	user, err := uc.users.FindByClerkID(clerkID)
	if err != nil || user == nil {
		return nil, errors.New("user not found")
	}
	req, err := uc.requests.FindByID(requestID)
	if err != nil || req == nil {
		return nil, errors.New("request not found")
	}
	if req.Status == "closed" || req.Status == "rejected" {
		return nil, errors.New("cannot send messages in a closed or rejected request")
	}

	prof, _ := uc.professionals.FindByUserID(user.ID)
	isClient := user.ID == req.ClientID
	isProfessional := prof != nil && prof.ID == req.ProfessionalID
	if !isClient && !isProfessional {
		return nil, errors.New("forbidden: only the client or professional can send messages")
	}

	msg, err := uc.messages.Create(&domain.Message{
		RequestID:  requestID,
		SenderID:   user.ID,
		SenderName: user.FullName,
		Content:    content,
	})
	if err != nil {
		return nil, err
	}

	if uc.hub != nil {
		uc.hub.Publish(requestID, msg)
	}

	if uc.notifications != nil {
		recipientUserID := req.ClientID
		if isClient {
			if prof == nil {
				prof, _ = uc.professionals.FindByID(req.ProfessionalID)
			}
			recipientUserID = ""
			if prof != nil {
				recipientUserID = prof.UserID
			}
		}
		if recipientUserID != "" && recipientUserID != user.ID {
			_ = uc.notifications.CreateForUser(recipientUserID, "job_new_message",
				fmt.Sprintf("%s te envió un mensaje", user.FullName), req.JobID)
		}
	}

	return msg, nil
}

func (uc *MessageUseCase) Subscribe(clerkID, requestID string) (<-chan *domain.Message, func(), error) {
	user, err := uc.users.FindByClerkID(clerkID)
	if err != nil || user == nil {
		return nil, nil, errors.New("user not found")
	}
	req, err := uc.requests.FindByID(requestID)
	if err != nil || req == nil {
		return nil, nil, errors.New("request not found")
	}
	prof, _ := uc.professionals.FindByUserID(user.ID)
	isClient := user.ID == req.ClientID
	isProfessional := prof != nil && prof.ID == req.ProfessionalID
	if !isClient && !isProfessional {
		return nil, nil, errors.New("forbidden")
	}
	if uc.hub == nil {
		return nil, nil, errors.New("realtime not configured")
	}
	ch, unsub := uc.hub.Subscribe(requestID)
	return ch, unsub, nil
}

func (uc *MessageUseCase) ListByRequest(clerkID, requestID string) ([]domain.Message, error) {
	user, err := uc.users.FindByClerkID(clerkID)
	if err != nil || user == nil {
		return nil, errors.New("user not found")
	}
	req, err := uc.requests.FindByID(requestID)
	if err != nil || req == nil {
		return nil, errors.New("request not found")
	}

	prof, _ := uc.professionals.FindByUserID(user.ID)
	isClient := user.ID == req.ClientID
	isProfessional := prof != nil && prof.ID == req.ProfessionalID
	if !isClient && !isProfessional {
		return nil, errors.New("forbidden")
	}

	msgs, err := uc.messages.FindByRequestID(requestID)
	if err != nil {
		return nil, err
	}

	// mark as read for the reader
	if isClient {
		_ = uc.messages.MarkReadForClient(requestID)
	} else {
		_ = uc.messages.MarkReadForProvider(requestID)
	}

	return msgs, nil
}
