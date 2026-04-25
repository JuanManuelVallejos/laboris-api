package domain

import "time"

type Notification struct {
	ID        string    `json:"id"`
	UserID    string    `json:"userId"`
	Type      string    `json:"type"`
	Message   string    `json:"message"`
	Read      bool      `json:"read"`
	CreatedAt time.Time `json:"createdAt"`
}

type NotificationRepository interface {
	Create(n *Notification) (*Notification, error)
	FindByUserID(userID string) ([]Notification, error)
	CountUnread(userID string) (int, error)
	MarkAllRead(userID string) error
}
