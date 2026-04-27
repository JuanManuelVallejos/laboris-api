package domain

import "time"

const (
	PaymentTypeVisit = "visit"
	PaymentTypeWork  = "work"

	PaymentStatusPending  = "pending"
	PaymentStatusPaid     = "paid"
	PaymentStatusReleased = "released"
	PaymentStatusRefunded = "refunded"
)

type Payment struct {
	ID          string    `json:"id"`
	JobID       string    `json:"jobId"`
	Type        string    `json:"type"`
	Amount      float64   `json:"amount"`
	Status      string    `json:"status"`
	Provider    string    `json:"provider"`
	ProviderRef string    `json:"providerRef,omitempty"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

type PaymentRepository interface {
	Create(p *Payment) (*Payment, error)
	FindByJobID(jobID string) ([]Payment, error)
	UpdateStatus(id, status string) error
}
