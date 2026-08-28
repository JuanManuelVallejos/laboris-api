package domain

import "time"

type Review struct {
	ID             string `json:"id"`
	ProfessionalID string `json:"professionalId"`
	// ReviewerID nunca se expone — solo el nombre, para mostrar quién dejó
	// la reseña sin filtrar el id interno de otro usuario.
	ReviewerID   string    `json:"-"`
	ReviewerName string    `json:"reviewerName"`
	Rating       int       `json:"rating"`
	Comment      string    `json:"comment"`
	CreatedAt    time.Time `json:"createdAt"`
}

type ReviewRepository interface {
	// Upsert crea o actualiza la reseña de este reviewer para este
	// profesional — un cliente solo puede tener una reseña vigente por
	// profesional (UNIQUE en la tabla), volver a enviar la actualiza.
	Upsert(r *Review) (*Review, error)
	FindByProfessionalID(professionalID string) ([]Review, error)
}
