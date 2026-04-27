package postgres

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/laboris/laboris-api/internal/domain"
)

type JobRepository struct {
	db *pgxpool.Pool
}

func NewJobRepository(db *pgxpool.Pool) *JobRepository {
	return &JobRepository{db: db}
}

const jobSelectCols = `
	j.id, j.request_id,
	j.client_id,      uc.full_name,
	j.professional_id, up.full_name, p.user_id,
	j.status,
	j.visit_scheduled_at, j.visit_quote_amount, j.work_quote_amount,
	j.work_description,   j.rework_count,       j.rework_notes,
	j.rework_quote_amount,
	j.cancel_reason,      j.completed_at,        j.cancelled_at,
	j.created_at,         j.updated_at`

const jobJoins = `
	FROM jobs j
	JOIN users uc ON uc.id = j.client_id
	JOIN professionals p ON p.id = j.professional_id
	JOIN users up ON up.id = p.user_id`

func scanJob(row interface{ Scan(...any) error }) (*domain.Job, error) {
	j := &domain.Job{}
	var (
		visitScheduledAt  *time.Time
		visitQuoteAmount  *float64
		workQuoteAmount   *float64
		workDescription   *string
		reworkNotes       *string
		reworkQuoteAmount *float64
		cancelReason      *string
		completedAt       *time.Time
		cancelledAt       *time.Time
	)
	if err := row.Scan(
		&j.ID, &j.RequestID,
		&j.ClientID, &j.ClientName,
		&j.ProfessionalID, &j.ProfessionalName, &j.ProfessionalUID,
		&j.Status,
		&visitScheduledAt, &visitQuoteAmount, &workQuoteAmount,
		&workDescription, &j.ReworkCount, &reworkNotes,
		&reworkQuoteAmount,
		&cancelReason, &completedAt, &cancelledAt,
		&j.CreatedAt, &j.UpdatedAt,
	); err != nil {
		return nil, err
	}
	j.VisitScheduledAt = visitScheduledAt
	j.VisitQuoteAmount = visitQuoteAmount
	j.WorkQuoteAmount = workQuoteAmount
	if workDescription != nil {
		j.WorkDescription = *workDescription
	}
	if reworkNotes != nil {
		j.ReworkNotes = *reworkNotes
	}
	j.ReworkQuoteAmount = reworkQuoteAmount
	if cancelReason != nil {
		j.CancelReason = *cancelReason
	}
	j.CompletedAt = completedAt
	j.CancelledAt = cancelledAt
	j.Payments = []domain.Payment{}
	return j, nil
}

func (r *JobRepository) Create(j *domain.Job) (*domain.Job, error) {
	err := r.db.QueryRow(context.Background(), `
		INSERT INTO jobs (request_id, client_id, professional_id)
		VALUES ($1, $2, $3)
		RETURNING id, status, created_at, updated_at
	`, j.RequestID, j.ClientID, j.ProfessionalID,
	).Scan(&j.ID, &j.Status, &j.CreatedAt, &j.UpdatedAt)
	if err != nil {
		return nil, err
	}
	j.Payments = []domain.Payment{}
	return j, nil
}

func (r *JobRepository) FindByID(id string) (*domain.Job, error) {
	row := r.db.QueryRow(context.Background(),
		`SELECT `+jobSelectCols+jobJoins+` WHERE j.id = $1`, id)
	j, err := scanJob(row)
	if err != nil {
		return nil, err
	}
	payments, err := r.fetchPayments(j.ID)
	if err != nil {
		return nil, err
	}
	j.Payments = payments
	return j, nil
}

func (r *JobRepository) FindByUserID(userID string) ([]domain.Job, error) {
	rows, err := r.db.Query(context.Background(),
		`SELECT `+jobSelectCols+jobJoins+`
		 WHERE j.client_id = $1 OR p.user_id = $1
		 ORDER BY j.updated_at DESC`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var jobs []domain.Job
	for rows.Next() {
		j, err := scanJob(rows)
		if err != nil {
			return nil, err
		}
		jobs = append(jobs, *j)
	}
	if jobs == nil {
		jobs = []domain.Job{}
	}
	return jobs, nil
}

func (r *JobRepository) FindByRequestID(requestID string) (*domain.Job, error) {
	row := r.db.QueryRow(context.Background(),
		`SELECT `+jobSelectCols+jobJoins+` WHERE j.request_id = $1`, requestID)
	return scanJob(row)
}

func (r *JobRepository) Update(j *domain.Job) (*domain.Job, error) {
	err := r.db.QueryRow(context.Background(), `
		UPDATE jobs SET
			status              = $2,
			visit_scheduled_at  = $3,
			visit_quote_amount  = $4,
			work_quote_amount   = $5,
			work_description    = NULLIF($6,''),
			rework_count        = $7,
			rework_notes        = NULLIF($8,''),
			rework_quote_amount = $9,
			cancel_reason       = NULLIF($10,''),
			completed_at        = $11,
			cancelled_at        = $12,
			updated_at          = NOW()
		WHERE id = $1
		RETURNING updated_at
	`,
		j.ID, j.Status,
		j.VisitScheduledAt, j.VisitQuoteAmount, j.WorkQuoteAmount,
		j.WorkDescription, j.ReworkCount, j.ReworkNotes,
		j.ReworkQuoteAmount,
		j.CancelReason, j.CompletedAt, j.CancelledAt,
	).Scan(&j.UpdatedAt)
	return j, err
}

func (r *JobRepository) fetchPayments(jobID string) ([]domain.Payment, error) {
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
