package domain

import "time"

type User struct {
	ID        string    `json:"id"`
	ClerkID   string    `json:"clerkId"`
	Email     string    `json:"email"`
	FullName  string    `json:"fullName"`
	CreatedAt time.Time `json:"createdAt"`
}

type UserRepository interface {
	FindByClerkID(clerkID string) (*User, error)
	Create(user *User) (*User, error)
	AddRole(userID string, role string) error
}
