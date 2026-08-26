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
	ID               string `json:"id"`
	ClientID         string `json:"clientId"`
	ClientName       string `json:"clientName"`
	ProfessionalID   string `json:"professionalId"`
	ProfessionalName string `json:"professionalName"`
	// AddressID es el domicilio guardado elegido al pedir el presupuesto —
	// nil en solicitudes viejas, creadas antes de este sistema (legacy).
	AddressID       *string      `json:"-"`
	Description     string       `json:"description"`
	Status          string       `json:"status"`
	RejectionReason string       `json:"rejectionReason"`
	JobID           string       `json:"jobId,omitempty"`
	Photos          []Attachment `json:"photos"`
	CreatedAt       time.Time    `json:"createdAt"`
}

type RequestRepository interface {
	Create(r *Request) (*Request, error)
	FindByID(id string) (*Request, error)
	FindByProfessionalID(professionalID string) ([]Request, error)
	FindByClientID(clientID string) ([]Request, error)
	UpdateStatus(id, status, reason string) (*Request, error)
	MarkViewed(id string) error
}
