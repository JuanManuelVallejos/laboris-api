package domain

import "time"

type ReworkRecord struct {
	ID          string    `json:"id"`
	JobID       string    `json:"jobId"`
	CycleNumber int       `json:"cycleNumber"`
	Notes       string    `json:"notes,omitempty"`
	QuoteAmount *float64  `json:"quoteAmount,omitempty"`
	CreatedAt   time.Time `json:"createdAt"`
}

type ReworkRecordRepository interface {
	Create(r *ReworkRecord) (*ReworkRecord, error)
	FindByJobID(jobID string) ([]ReworkRecord, error)
	UpdateQuoteAmount(jobID string, cycleNumber int, amount float64) error
}
