package domain

import "time"

type ReworkRecord struct {
	ID          string     `json:"id"`
	JobID       string     `json:"jobId"`
	CycleNumber int        `json:"cycleNumber"`
	Notes       string     `json:"notes,omitempty"`
	QuoteAmount *float64   `json:"quoteAmount,omitempty"`
	NoCharge    bool       `json:"noCharge"`
	ScheduledAt *time.Time `json:"scheduledAt,omitempty"`
	CreatedAt   time.Time  `json:"createdAt"`
}

type ReworkRecordRepository interface {
	Create(r *ReworkRecord) (*ReworkRecord, error)
	FindByJobID(jobID string) ([]ReworkRecord, error)
	UpdateQuoteAmount(jobID string, cycleNumber int, amount float64) error
	UpdateScheduledAt(jobID string, cycleNumber int, scheduledAt *time.Time) error
	MarkNoCharge(jobID string, cycleNumber int) error
}
