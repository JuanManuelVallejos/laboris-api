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
	// MonthlyEarningsByProfessional agrupa por mes (más reciente primero) el
	// dinero ya cobrado (paid/released) de los trabajos de este profesional
	// — se usa en las estadísticas de "Mi actividad".
	MonthlyEarningsByProfessional(professionalID string) ([]MonthlyEarning, error)
	// MonthlySpendingByClient es el equivalente para "Mi actividad" del cliente:
	// cuánto pagó (paid/released) este cliente, agrupado por mes.
	MonthlySpendingByClient(clientID string) ([]MonthlyEarning, error)
}

// MonthlyEarning es un renglón de "cuánto gané en tal mes" — Month va en
// formato "2026-08" (el frontend se encarga de darle formato legible).
type MonthlyEarning struct {
	Month     string  `json:"month"`
	Amount    float64 `json:"amount"`
	JobsCount int     `json:"jobsCount"`
}
