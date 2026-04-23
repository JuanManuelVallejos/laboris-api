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

func (r *professionalRepository) FindByUserID(userID string) (*domain.Professional, error) {
	for _, p := range r.data {
		if p.UserID == userID {
			return &p, nil
		}
	}
	return nil, nil
}

func (r *professionalRepository) UpdateByUserID(userID, trade, zone, bio string) (*domain.Professional, error) {
	for i, p := range r.data {
		if p.UserID == userID {
			r.data[i].Trade = trade
			r.data[i].Zone = zone
			r.data[i].Bio = bio
			return &r.data[i], nil
		}
	}
	return nil, nil
}

func (r *professionalRepository) Create(p *domain.Professional) (*domain.Professional, error) {
	r.data = append(r.data, *p)
	return p, nil
}
