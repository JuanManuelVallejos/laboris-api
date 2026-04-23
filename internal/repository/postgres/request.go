package postgres

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/laboris/laboris-api/internal/domain"
)

type RequestRepository struct {
	db *pgxpool.Pool
}

func NewRequestRepository(db *pgxpool.Pool) *RequestRepository {
	return &RequestRepository{db: db}
}

func (r *RequestRepository) Create(req *domain.Request) (*domain.Request, error) {
	err := r.db.QueryRow(context.Background(), `
		INSERT INTO requests (client_id, professional_id, description)
		VALUES ($1, $2, $3)
		RETURNING id, client_id, professional_id, description, status, COALESCE(rejection_reason,''), created_at
	`, req.ClientID, req.ProfessionalID, req.Description,
	).Scan(&req.ID, &req.ClientID, &req.ProfessionalID, &req.Description, &req.Status, &req.RejectionReason, &req.CreatedAt)
	return req, err
}

func (r *RequestRepository) FindByProfessionalID(professionalID string) ([]domain.Request, error) {
	rows, err := r.db.Query(context.Background(), `
		SELECT rq.id, rq.client_id, uc.full_name, rq.professional_id, up.full_name,
		       rq.description, rq.status, COALESCE(rq.rejection_reason,''), rq.created_at
		FROM requests rq
		JOIN users uc ON uc.id = rq.client_id
		JOIN professionals p ON p.id = rq.professional_id
		JOIN users up ON up.id = p.user_id
		WHERE rq.professional_id = $1
		ORDER BY rq.created_at DESC
	`, professionalID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make([]domain.Request, 0)
	for rows.Next() {
		var rq domain.Request
		if err := rows.Scan(&rq.ID, &rq.ClientID, &rq.ClientName, &rq.ProfessionalID, &rq.ProfessionalName,
			&rq.Description, &rq.Status, &rq.RejectionReason, &rq.CreatedAt); err != nil {
			return nil, err
		}
		result = append(result, rq)
	}
	return result, nil
}

func (r *RequestRepository) FindByClientID(clientID string) ([]domain.Request, error) {
	rows, err := r.db.Query(context.Background(), `
		SELECT rq.id, rq.client_id, uc.full_name, rq.professional_id, up.full_name,
		       rq.description, rq.status, COALESCE(rq.rejection_reason,''), rq.created_at
		FROM requests rq
		JOIN users uc ON uc.id = rq.client_id
		JOIN professionals p ON p.id = rq.professional_id
		JOIN users up ON up.id = p.user_id
		WHERE rq.client_id = $1
		ORDER BY rq.created_at DESC
	`, clientID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make([]domain.Request, 0)
	for rows.Next() {
		var rq domain.Request
		if err := rows.Scan(&rq.ID, &rq.ClientID, &rq.ClientName, &rq.ProfessionalID, &rq.ProfessionalName,
			&rq.Description, &rq.Status, &rq.RejectionReason, &rq.CreatedAt); err != nil {
			return nil, err
		}
		result = append(result, rq)
	}
	return result, nil
}

func (r *RequestRepository) UpdateStatus(id, status, reason string) (*domain.Request, error) {
	rq := &domain.Request{}
	err := r.db.QueryRow(context.Background(), `
		UPDATE requests SET status = $2, rejection_reason = NULLIF($3, '')
		WHERE id = $1
		RETURNING id, client_id, professional_id, description, status, COALESCE(rejection_reason,''), created_at
	`, id, status, reason).Scan(&rq.ID, &rq.ClientID, &rq.ProfessionalID, &rq.Description, &rq.Status, &rq.RejectionReason, &rq.CreatedAt)
	return rq, err
}
