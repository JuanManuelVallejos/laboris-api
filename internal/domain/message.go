package domain

import "time"

type Message struct {
	ID                  string    `json:"id"`
	RequestID           string    `json:"requestId"`
	SenderID            string    `json:"senderId"`
	SenderName          string    `json:"senderName"`
	Content             string    `json:"content"`
	IsUnreadForClient   bool      `json:"isUnreadForClient"`
	IsUnreadForProvider bool      `json:"isUnreadForProvider"`
	CreatedAt           time.Time `json:"createdAt"`
}

type MessageRepository interface {
	Create(m *Message) (*Message, error)
	FindByRequestID(requestID string) ([]Message, error)
	MarkReadForClient(requestID string) error
	MarkReadForProvider(requestID string) error
}
