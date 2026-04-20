package memory

import (
	"errors"

	"github.com/laboris/laboris-api/internal/domain"
)

type professionalRepository struct {
	data []domain.Professional
}

func NewProfessionalRepository() domain.ProfessionalRepository {
	return &professionalRepository{
		data: []domain.Professional{
			{ID: "1", Name: "Tomás Rivas", Trade: "plomero", Zone: "Zona Sur", Rating: 4.8, Verified: true},
			{ID: "2", Name: "Carlos Méndez", Trade: "electricista", Zone: "CABA", Rating: 4.5, Verified: true},
			{ID: "3", Name: "Roberto Giménez", Trade: "gasista", Zone: "CABA", Rating: 4.9, Verified: true},
		},
	}
}

func (r *professionalRepository) FindAll() ([]domain.Professional, error) {
	return r.data, nil
}

func (r *professionalRepository) FindByID(id string) (*domain.Professional, error) {
	for _, p := range r.data {
		if p.ID == id {
			return &p, nil
		}
	}
	return nil, errors.New("professional not found")
}
