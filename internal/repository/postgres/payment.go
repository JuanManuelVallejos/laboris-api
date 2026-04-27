package postgres

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/laboris/laboris-api/internal/domain"
)

type PaymentRepository struct {
	db *pgxpool.Pool
}

func NewPaymentRepository(db *pgxpool.Pool) *PaymentRepository {
	return &PaymentRepository{db: db}
}

func (r *PaymentRepository) Create(p *domain.Payment) (*domain.Payment, error) {
	err := r.db.QueryRow(context.Background(), `
		INSERT INTO payments (job_id, type, amount, status, provider, provider_ref)
		VALUES ($1, $2, $3, $4, $5, NULLIF($6,''))
		RETURNING id, created_at, updated_at
	`, p.JobID, p.Type, p.Amount, p.Status, p.Provider, p.ProviderRef,
	).Scan(&p.ID, &p.CreatedAt, &p.UpdatedAt)
	return p, err
}

func (r *PaymentRepository) FindByJobID(jobID string) ([]domain.Payment, error) {
	rows, err := r.db.Query(context.Background(), `
		SELECT id, job_id, type, amount, status, provider, COALESCE(provider_ref,''), created_at, updated_at
		FROM payments WHERE job_id = $1 ORDER BY created_at ASC
	`, jobID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var payments []domain.Payment
	for rows.Next() {
		var p domain.Payment
		if err := rows.Scan(&p.ID, &p.JobID, &p.Type, &p.Amount, &p.Status, &p.Provider, &p.ProviderRef, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, err
		}
		payments = append(payments, p)
	}
	if payments == nil {
		payments = []domain.Payment{}
	}
	return payments, nil
}

func (r *PaymentRepository) UpdateStatus(id, status string) error {
	_, err := r.db.Exec(context.Background(),
		`UPDATE payments SET status = $2, updated_at = NOW() WHERE id = $1`, id, status)
	return err
}
