package domain

import "time"

type Review struct {
	ID             string    `json:"id"`
	ProfessionalID string    `json:"professionalId"`
	ReviewerID     string    `json:"reviewerId"`
	Rating         int       `json:"rating"`
	Comment        string    `json:"comment"`
	CreatedAt      time.Time `json:"createdAt"`
}
