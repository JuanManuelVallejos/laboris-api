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
	AddressID *string `json:"-"`
	// AddressLat/AddressLng son las coordenadas congeladas al momento de
	// crear la solicitud (mismo criterio que Address) — se usan para el
	// círculo aproximado que ve el profesional antes de que se revele la
	// dirección exacta. nil en solicitudes legacy sin coords congeladas.
	AddressLat *float64 `json:"-"`
	AddressLng *float64 `json:"-"`
	// Address es el texto del domicilio — completo si ya se reveló
	// (AddressRevealed), o recortado a lo sumo a nivel localidad si no.
	// Vacío en solicitudes legacy sin address_id.
	Address string `json:"address,omitempty"`
	// AddressRevealed indica si Address trae el domicilio completo. Para el
	// cliente siempre es true (es su propio domicilio); para el profesional
	// depende del estado del trabajo asociado — ver RequestUseCase.
	AddressRevealed bool `json:"addressRevealed"`
	// JobStatus es el status del job asociado (si existe) — interno, se usa
	// solo para decidir el gating de arriba, nunca se expone.
	JobStatus       string       `json:"-"`
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
