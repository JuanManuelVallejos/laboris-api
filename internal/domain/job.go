package domain

import "time"

const (
	JobStatusPendingVisit    = "pending_visit"
	JobStatusVisitScheduled  = "visit_scheduled"
	JobStatusVisitQuoted     = "visit_quoted"
	JobStatusVisitPaid       = "visit_paid"
	JobStatusVisitCompleted  = "visit_completed"
	JobStatusWorkQuoted      = "work_quoted"
	JobStatusWorkApproved    = "work_approved"
	JobStatusWorkInProgress  = "work_in_progress"
	JobStatusWorkDelivered   = "work_delivered"
	JobStatusReworkRequested = "rework_requested"
	JobStatusReworkQuoted    = "rework_quoted"
	JobStatusCompleted       = "completed"
	JobStatusCancelled       = "cancelled"
)

// ValidTransitions defines allowed state transitions for a Job.
var ValidTransitions = map[string]map[string]bool{
	JobStatusPendingVisit:    {JobStatusVisitScheduled: true, JobStatusWorkQuoted: true, JobStatusCancelled: true},
	JobStatusVisitScheduled:  {JobStatusVisitCompleted: true, JobStatusVisitQuoted: true, JobStatusCancelled: true},
	JobStatusVisitQuoted:     {JobStatusVisitPaid: true, JobStatusCancelled: true},
	JobStatusVisitPaid:       {JobStatusVisitCompleted: true, JobStatusCancelled: true},
	JobStatusVisitCompleted:  {JobStatusWorkQuoted: true, JobStatusCancelled: true},
	JobStatusWorkQuoted:      {JobStatusWorkApproved: true, JobStatusCancelled: true},
	JobStatusWorkApproved:    {JobStatusWorkInProgress: true, JobStatusCancelled: true},
	JobStatusWorkInProgress:  {JobStatusWorkDelivered: true, JobStatusCancelled: true},
	JobStatusWorkDelivered:   {JobStatusReworkRequested: true, JobStatusCompleted: true},
	JobStatusReworkRequested: {JobStatusReworkQuoted: true, JobStatusWorkInProgress: true, JobStatusCancelled: true},
	JobStatusReworkQuoted:    {JobStatusWorkInProgress: true, JobStatusCancelled: true},
	JobStatusCompleted:       {},
	JobStatusCancelled:       {},
}

type Job struct {
	ID                string     `json:"id"`
	RequestID         string     `json:"requestId"`
	ClientID          string     `json:"clientId"`
	ClientName        string     `json:"clientName"`
	ProfessionalID    string     `json:"professionalId"`
	ProfessionalName  string     `json:"professionalName"`
	ProfessionalUID   string     `json:"-"` // professional's user_id — used for auth, not exposed
	Status            string     `json:"status"`
	VisitScheduledAt  *time.Time `json:"visitScheduledAt,omitempty"`
	VisitQuoteAmount  *float64   `json:"visitQuoteAmount,omitempty"`
	WorkQuoteAmount   *float64   `json:"workQuoteAmount,omitempty"`
	WorkDescription   string     `json:"workDescription,omitempty"`
	ReworkCount       int            `json:"reworkCount"`
	ReworkNotes       string         `json:"reworkNotes,omitempty"`
	ReworkQuoteAmount *float64       `json:"reworkQuoteAmount,omitempty"`
	CancelReason      string         `json:"cancelReason,omitempty"`
	CompletedAt       *time.Time     `json:"completedAt,omitempty"`
	CancelledAt       *time.Time     `json:"cancelledAt,omitempty"`
	Payments          []Payment      `json:"payments"`
	ReworkRecords     []ReworkRecord `json:"reworkRecords"`
	CreatedAt         time.Time      `json:"createdAt"`
	UpdatedAt         time.Time      `json:"updatedAt"`
}

type JobRepository interface {
	Create(j *Job) (*Job, error)
	FindByID(id string) (*Job, error)
	FindByUserID(userID string) ([]Job, error)
	FindByRequestID(requestID string) (*Job, error)
	Update(j *Job) (*Job, error)
}
