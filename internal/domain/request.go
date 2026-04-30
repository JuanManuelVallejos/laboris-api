package domain

import "time"

const (
	RequestStatusPending  = "pending"
	RequestStatusViewed   = "viewed"
	RequestStatusAccepted = "accepted"
	RequestStatusRejected = "rejected"
	RequestStatusExpired  = "expired"
)

type Request struct {
	ID               string    `json:"id"`
	ClientID         string    `json:"clientId"`
	ClientName       string    `json:"clientName"`
	ProfessionalID   string    `json:"professionalId"`
	ProfessionalName string    `json:"professionalName"`
	Description      string    `json:"description"`
	Status           string    `json:"status"`
	RejectionReason  string    `json:"rejectionReason"`
	JobID            string    `json:"jobId,omitempty"`
	CreatedAt        time.Time `json:"createdAt"`
}

type RequestRepository interface {
	Create(r *Request) (*Request, error)
	FindByID(id string) (*Request, error)
	FindByProfessionalID(professionalID string) ([]Request, error)
	FindByClientID(clientID string) ([]Request, error)
	UpdateStatus(id, status, reason string) (*Request, error)
	MarkAllPendingAsViewed(professionalID string) error
}
