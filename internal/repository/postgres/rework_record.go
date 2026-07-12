package postgres

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/laboris/laboris-api/internal/domain"
)

type ReworkRecordRepository struct {
	db *pgxpool.Pool
}

func NewReworkRecordRepository(db *pgxpool.Pool) *ReworkRecordRepository {
	return &ReworkRecordRepository{db: db}
}

func (r *ReworkRecordRepository) Create(rec *domain.ReworkRecord) (*domain.ReworkRecord, error) {
	err := r.db.QueryRow(context.Background(), `
		INSERT INTO job_rework_records (job_id, cycle_number, notes)
		VALUES ($1, $2, NULLIF($3, ''))
		RETURNING id, created_at
	`, rec.JobID, rec.CycleNumber, rec.Notes).Scan(&rec.ID, &rec.CreatedAt)
	if err != nil {
		return nil, err
	}
	return rec, nil
}

func (r *ReworkRecordRepository) FindByJobID(jobID string) ([]domain.ReworkRecord, error) {
	rows, err := r.db.Query(context.Background(), `
		SELECT id, job_id, cycle_number, COALESCE(notes,''), quote_amount, scheduled_at, created_at
		FROM job_rework_records
		WHERE job_id = $1
		ORDER BY cycle_number ASC
	`, jobID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var records []domain.ReworkRecord
	for rows.Next() {
		var rec domain.ReworkRecord
		var quoteAmount *float64
		var scheduledAt *time.Time
		if err := rows.Scan(&rec.ID, &rec.JobID, &rec.CycleNumber, &rec.Notes, &quoteAmount, &scheduledAt, &rec.CreatedAt); err != nil {
			return nil, err
		}
		rec.QuoteAmount = quoteAmount
		rec.ScheduledAt = scheduledAt
		records = append(records, rec)
	}
	if records == nil {
		records = []domain.ReworkRecord{}
	}
	return records, nil
}

func (r *ReworkRecordRepository) UpdateQuoteAmount(jobID string, cycleNumber int, amount float64) error {
	_, err := r.db.Exec(context.Background(), `
		UPDATE job_rework_records SET quote_amount = $3
		WHERE job_id = $1 AND cycle_number = $2
	`, jobID, cycleNumber, amount)
	return err
}

func (r *ReworkRecordRepository) UpdateScheduledAt(jobID string, cycleNumber int, scheduledAt *time.Time) error {
	_, err := r.db.Exec(context.Background(), `
		UPDATE job_rework_records SET scheduled_at = $3
		WHERE job_id = $1 AND cycle_number = $2
	`, jobID, cycleNumber, scheduledAt)
	return err
}
