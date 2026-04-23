package domain

import "time"

type Request struct {
	ID             string    `json:"id"`
	ClientID       string    `json:"clientId"`
	ClientName     string    `json:"clientName"`
	ProfessionalID string    `json:"professionalId"`
	Description    string    `json:"description"`
	Status         string    `json:"status"`
	CreatedAt      time.Time `json:"createdAt"`
}

type RequestRepository interface {
	Create(r *Request) (*Request, error)
	FindByProfessionalID(professionalID string) ([]Request, error)
	FindByClientID(clientID string) ([]Request, error)
	UpdateStatus(id, status string) (*Request, error)
}
