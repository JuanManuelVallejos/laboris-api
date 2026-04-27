package usecase

import (
	"errors"
	"fmt"
	"time"

	"github.com/laboris/laboris-api/internal/domain"
)

type JobUseCase struct {
	jobs          domain.JobRepository
	payments      domain.PaymentRepository
	users         domain.UserRepository
	professionals domain.ProfessionalRepository
	reworks       domain.ReworkRecordRepository
	notifications *NotificationUseCase
}

func NewJobUseCase(
	jobs domain.JobRepository,
	payments domain.PaymentRepository,
	users domain.UserRepository,
	professionals domain.ProfessionalRepository,
	reworks domain.ReworkRecordRepository,
) *JobUseCase {
	return &JobUseCase{jobs: jobs, payments: payments, users: users, professionals: professionals, reworks: reworks}
}

func (uc *JobUseCase) SetNotifications(n *NotificationUseCase) {
	uc.notifications = n
}

// CreateFromRequest is called by RequestUseCase when a request is accepted.
func (uc *JobUseCase) CreateFromRequest(requestID, clientID, professionalID string) (*domain.Job, error) {
	return uc.jobs.Create(&domain.Job{
		RequestID:      requestID,
		ClientID:       clientID,
		ProfessionalID: professionalID,
	})
}

func (uc *JobUseCase) GetByID(clerkID, jobID string) (*domain.Job, error) {
	user, _, err := uc.resolveUser(clerkID)
	if err != nil {
		return nil, err
	}
	job, err := uc.jobs.FindByID(jobID)
	if err != nil {
		return nil, err
	}
	if !uc.canAccess(user, job) {
		return nil, errors.New("forbidden")
	}
	return job, nil
}

func (uc *JobUseCase) ListByUser(clerkID string) ([]domain.Job, error) {
	user, _, err := uc.resolveUser(clerkID)
	if err != nil {
		return nil, err
	}
	return uc.jobs.FindByUserID(user.ID)
}

// ScheduleVisit: pending_visit → visit_scheduled (professional)
func (uc *JobUseCase) ScheduleVisit(clerkID, jobID string, scheduledAt time.Time) (*domain.Job, error) {
	user, prof, err := uc.resolveUser(clerkID)
	if err != nil {
		return nil, err
	}
	job, err := uc.jobs.FindByID(jobID)
	if err != nil {
		return nil, err
	}
	if prof == nil || prof.ID != job.ProfessionalID {
		return nil, errors.New("forbidden: only the professional can schedule the visit")
	}
	if err := validateTransition(job.Status, domain.JobStatusVisitScheduled); err != nil {
		return nil, err
	}
	job.Status = domain.JobStatusVisitScheduled
	job.VisitScheduledAt = &scheduledAt
	job, err = uc.jobs.Update(job)
	if err != nil {
		return nil, err
	}
	uc.notify(job.ClientID, "job_visit_scheduled",
		fmt.Sprintf("%s agendó la visita para el %s", job.ProfessionalName, scheduledAt.Format("02/01 15:04")))
	_ = user
	return job, nil
}

// SubmitVisitQuote: visit_scheduled → visit_quoted (professional)
func (uc *JobUseCase) SubmitVisitQuote(clerkID, jobID string, amount float64) (*domain.Job, error) {
	_, prof, err := uc.resolveUser(clerkID)
	if err != nil {
		return nil, err
	}
	job, err := uc.jobs.FindByID(jobID)
	if err != nil {
		return nil, err
	}
	if prof == nil || prof.ID != job.ProfessionalID {
		return nil, errors.New("forbidden: only the professional can submit the visit quote")
	}
	if err := validateTransition(job.Status, domain.JobStatusVisitQuoted); err != nil {
		return nil, err
	}
	job.Status = domain.JobStatusVisitQuoted
	job.VisitQuoteAmount = &amount
	job, err = uc.jobs.Update(job)
	if err != nil {
		return nil, err
	}
	uc.notify(job.ClientID, "job_visit_quoted",
		fmt.Sprintf("%s envió la cotización de visita: $%.2f", job.ProfessionalName, amount))
	return job, nil
}

// SkipVisit: pending_visit → work_quoted (professional, sets visit_quote_amount = 0)
func (uc *JobUseCase) SkipVisit(clerkID, jobID string, workAmount float64, workDescription string) (*domain.Job, error) {
	_, prof, err := uc.resolveUser(clerkID)
	if err != nil {
		return nil, err
	}
	job, err := uc.jobs.FindByID(jobID)
	if err != nil {
		return nil, err
	}
	if prof == nil || prof.ID != job.ProfessionalID {
		return nil, errors.New("forbidden: only the professional can skip the visit")
	}
	if err := validateTransition(job.Status, domain.JobStatusWorkQuoted); err != nil {
		return nil, err
	}
	zero := 0.0
	job.Status = domain.JobStatusWorkQuoted
	job.VisitQuoteAmount = &zero
	job.WorkQuoteAmount = &workAmount
	job.WorkDescription = workDescription
	job, err = uc.jobs.Update(job)
	if err != nil {
		return nil, err
	}
	uc.notify(job.ClientID, "job_work_quoted",
		fmt.Sprintf("%s envió cotización del trabajo: $%.2f", job.ProfessionalName, workAmount))
	return job, nil
}

// PayVisit: visit_quoted → visit_paid (client, mocked)
func (uc *JobUseCase) PayVisit(clerkID, jobID string) (*domain.Job, error) {
	user, _, err := uc.resolveUser(clerkID)
	if err != nil {
		return nil, err
	}
	job, err := uc.jobs.FindByID(jobID)
	if err != nil {
		return nil, err
	}
	if user.ID != job.ClientID {
		return nil, errors.New("forbidden: only the client can pay for the visit")
	}
	if err := validateTransition(job.Status, domain.JobStatusVisitPaid); err != nil {
		return nil, err
	}
	job.Status = domain.JobStatusVisitPaid
	job, err = uc.jobs.Update(job)
	if err != nil {
		return nil, err
	}
	amount := 0.0
	if job.VisitQuoteAmount != nil {
		amount = *job.VisitQuoteAmount
	}
	_, _ = uc.payments.Create(&domain.Payment{
		JobID:    job.ID,
		Type:     domain.PaymentTypeVisit,
		Amount:   amount,
		Status:   domain.PaymentStatusPaid,
		Provider: "mock",
	})
	uc.notify(job.ProfessionalUID, "job_visit_paid",
		fmt.Sprintf("%s pagó la visita ($%.2f)", job.ClientName, amount))
	return job, nil
}

// CompleteVisit: visit_paid → visit_completed (professional)
func (uc *JobUseCase) CompleteVisit(clerkID, jobID string) (*domain.Job, error) {
	_, prof, err := uc.resolveUser(clerkID)
	if err != nil {
		return nil, err
	}
	job, err := uc.jobs.FindByID(jobID)
	if err != nil {
		return nil, err
	}
	if prof == nil || prof.ID != job.ProfessionalID {
		return nil, errors.New("forbidden: only the professional can confirm the visit")
	}
	if err := validateTransition(job.Status, domain.JobStatusVisitCompleted); err != nil {
		return nil, err
	}
	job.Status = domain.JobStatusVisitCompleted
	job, err = uc.jobs.Update(job)
	if err != nil {
		return nil, err
	}
	uc.notify(job.ClientID, "job_visit_completed",
		fmt.Sprintf("%s confirmó que realizó la visita", job.ProfessionalName))
	return job, nil
}

// SubmitWorkQuote: visit_completed → work_quoted (professional)
func (uc *JobUseCase) SubmitWorkQuote(clerkID, jobID string, amount float64, description string) (*domain.Job, error) {
	_, prof, err := uc.resolveUser(clerkID)
	if err != nil {
		return nil, err
	}
	job, err := uc.jobs.FindByID(jobID)
	if err != nil {
		return nil, err
	}
	if prof == nil || prof.ID != job.ProfessionalID {
		return nil, errors.New("forbidden: only the professional can submit the work quote")
	}
	if err := validateTransition(job.Status, domain.JobStatusWorkQuoted); err != nil {
		return nil, err
	}
	job.Status = domain.JobStatusWorkQuoted
	job.WorkQuoteAmount = &amount
	job.WorkDescription = description
	job, err = uc.jobs.Update(job)
	if err != nil {
		return nil, err
	}
	uc.notify(job.ClientID, "job_work_quoted",
		fmt.Sprintf("%s envió cotización del trabajo: $%.2f", job.ProfessionalName, amount))
	return job, nil
}

// ApproveWorkQuote: work_quoted → work_approved (client)
func (uc *JobUseCase) ApproveWorkQuote(clerkID, jobID string) (*domain.Job, error) {
	user, _, err := uc.resolveUser(clerkID)
	if err != nil {
		return nil, err
	}
	job, err := uc.jobs.FindByID(jobID)
	if err != nil {
		return nil, err
	}
	if user.ID != job.ClientID {
		return nil, errors.New("forbidden: only the client can approve the work quote")
	}
	if err := validateTransition(job.Status, domain.JobStatusWorkApproved); err != nil {
		return nil, err
	}
	job.Status = domain.JobStatusWorkApproved
	job, err = uc.jobs.Update(job)
	if err != nil {
		return nil, err
	}
	uc.notify(job.ProfessionalUID, "job_work_approved",
		fmt.Sprintf("%s aprobó la cotización del trabajo", job.ClientName))
	return job, nil
}

// StartWork: work_approved → work_in_progress (professional)
func (uc *JobUseCase) StartWork(clerkID, jobID string) (*domain.Job, error) {
	_, prof, err := uc.resolveUser(clerkID)
	if err != nil {
		return nil, err
	}
	job, err := uc.jobs.FindByID(jobID)
	if err != nil {
		return nil, err
	}
	if prof == nil || prof.ID != job.ProfessionalID {
		return nil, errors.New("forbidden: only the professional can start the work")
	}
	if err := validateTransition(job.Status, domain.JobStatusWorkInProgress); err != nil {
		return nil, err
	}
	job.Status = domain.JobStatusWorkInProgress
	job, err = uc.jobs.Update(job)
	if err != nil {
		return nil, err
	}
	uc.notify(job.ClientID, "job_work_in_progress",
		fmt.Sprintf("%s comenzó el trabajo", job.ProfessionalName))
	return job, nil
}

// DeliverWork: work_in_progress → work_delivered (professional)
func (uc *JobUseCase) DeliverWork(clerkID, jobID string) (*domain.Job, error) {
	_, prof, err := uc.resolveUser(clerkID)
	if err != nil {
		return nil, err
	}
	job, err := uc.jobs.FindByID(jobID)
	if err != nil {
		return nil, err
	}
	if prof == nil || prof.ID != job.ProfessionalID {
		return nil, errors.New("forbidden: only the professional can mark work as delivered")
	}
	if err := validateTransition(job.Status, domain.JobStatusWorkDelivered); err != nil {
		return nil, err
	}
	job.Status = domain.JobStatusWorkDelivered
	job, err = uc.jobs.Update(job)
	if err != nil {
		return nil, err
	}
	uc.notify(job.ClientID, "job_work_delivered",
		fmt.Sprintf("%s marcó el trabajo como entregado. Revisá y aprobá o pedí correcciones.", job.ProfessionalName))
	return job, nil
}

// RequestRework: work_delivered → rework_requested (client)
func (uc *JobUseCase) RequestRework(clerkID, jobID string, notes string) (*domain.Job, error) {
	user, _, err := uc.resolveUser(clerkID)
	if err != nil {
		return nil, err
	}
	job, err := uc.jobs.FindByID(jobID)
	if err != nil {
		return nil, err
	}
	if user.ID != job.ClientID {
		return nil, errors.New("forbidden: only the client can request rework")
	}
	if err := validateTransition(job.Status, domain.JobStatusReworkRequested); err != nil {
		return nil, err
	}
	job.Status = domain.JobStatusReworkRequested
	job.ReworkCount++
	job.ReworkNotes = notes
	job, err = uc.jobs.Update(job)
	if err != nil {
		return nil, err
	}
	if uc.reworks != nil {
		_, _ = uc.reworks.Create(&domain.ReworkRecord{
			JobID:       job.ID,
			CycleNumber: job.ReworkCount,
			Notes:       notes,
		})
	}
	msg := fmt.Sprintf("%s solicitó correcciones (retrabajo #%d)", job.ClientName, job.ReworkCount)
	uc.notify(job.ProfessionalUID, "job_rework_requested", msg)
	if job.ReworkCount > 2 {
		// notify admin users — for now notify professional as well with escalation flag
		uc.notify(job.ProfessionalUID, "job_rework_escalated",
			fmt.Sprintf("El trabajo lleva %d retrabajos. Un admin puede mediar si es necesario.", job.ReworkCount))
	}
	return job, nil
}

// SubmitReworkQuote: rework_requested → rework_quoted (professional, extra cost)
func (uc *JobUseCase) SubmitReworkQuote(clerkID, jobID string, amount float64) (*domain.Job, error) {
	_, prof, err := uc.resolveUser(clerkID)
	if err != nil {
		return nil, err
	}
	job, err := uc.jobs.FindByID(jobID)
	if err != nil {
		return nil, err
	}
	if prof == nil || prof.ID != job.ProfessionalID {
		return nil, errors.New("forbidden: only the professional can submit a rework quote")
	}
	if err := validateTransition(job.Status, domain.JobStatusReworkQuoted); err != nil {
		return nil, err
	}
	job.Status = domain.JobStatusReworkQuoted
	job.ReworkQuoteAmount = &amount
	job, err = uc.jobs.Update(job)
	if err != nil {
		return nil, err
	}
	if uc.reworks != nil {
		_ = uc.reworks.UpdateQuoteAmount(job.ID, job.ReworkCount, amount)
	}
	uc.notify(job.ClientID, "job_rework_quoted",
		fmt.Sprintf("%s cotizó las correcciones en $%.2f. Aprobá para retomar el trabajo.", job.ProfessionalName, amount))
	return job, nil
}

// ApproveReworkQuote: rework_quoted → work_in_progress (client)
func (uc *JobUseCase) ApproveReworkQuote(clerkID, jobID string) (*domain.Job, error) {
	user, _, err := uc.resolveUser(clerkID)
	if err != nil {
		return nil, err
	}
	job, err := uc.jobs.FindByID(jobID)
	if err != nil {
		return nil, err
	}
	if user.ID != job.ClientID {
		return nil, errors.New("forbidden: only the client can approve the rework quote")
	}
	if err := validateTransition(job.Status, domain.JobStatusWorkInProgress); err != nil {
		return nil, err
	}
	job.Status = domain.JobStatusWorkInProgress
	job, err = uc.jobs.Update(job)
	if err != nil {
		return nil, err
	}
	uc.notify(job.ProfessionalUID, "job_rework_quote_approved",
		fmt.Sprintf("%s aprobó la cotización de correcciones. Retomá el trabajo.", job.ClientName))
	return job, nil
}

// AcceptRework: rework_requested → work_in_progress (professional, no extra cost)
func (uc *JobUseCase) AcceptRework(clerkID, jobID string) (*domain.Job, error) {
	_, prof, err := uc.resolveUser(clerkID)
	if err != nil {
		return nil, err
	}
	job, err := uc.jobs.FindByID(jobID)
	if err != nil {
		return nil, err
	}
	if prof == nil || prof.ID != job.ProfessionalID {
		return nil, errors.New("forbidden: only the professional can accept rework")
	}
	if err := validateTransition(job.Status, domain.JobStatusWorkInProgress); err != nil {
		return nil, err
	}
	job.Status = domain.JobStatusWorkInProgress
	job.ReworkQuoteAmount = nil // no extra cost — clear any amount from a previous cycle
	job, err = uc.jobs.Update(job)
	if err != nil {
		return nil, err
	}
	uc.notify(job.ClientID, "job_rework_accepted",
		fmt.Sprintf("%s aceptó las correcciones y retomó el trabajo", job.ProfessionalName))
	return job, nil
}

// ApproveDelivery: work_delivered → completed (client, mocked payment release)
func (uc *JobUseCase) ApproveDelivery(clerkID, jobID string) (*domain.Job, error) {
	user, _, err := uc.resolveUser(clerkID)
	if err != nil {
		return nil, err
	}
	job, err := uc.jobs.FindByID(jobID)
	if err != nil {
		return nil, err
	}
	if user.ID != job.ClientID {
		return nil, errors.New("forbidden: only the client can approve the delivery")
	}
	if err := validateTransition(job.Status, domain.JobStatusCompleted); err != nil {
		return nil, err
	}
	now := time.Now()
	job.Status = domain.JobStatusCompleted
	job.CompletedAt = &now
	job, err = uc.jobs.Update(job)
	if err != nil {
		return nil, err
	}
	amount := 0.0
	if job.WorkQuoteAmount != nil {
		amount = *job.WorkQuoteAmount
	}
	_, _ = uc.payments.Create(&domain.Payment{
		JobID:    job.ID,
		Type:     domain.PaymentTypeWork,
		Amount:   amount,
		Status:   domain.PaymentStatusReleased,
		Provider: "mock",
	})
	uc.notify(job.ProfessionalUID, "job_completed",
		fmt.Sprintf("%s aprobó el trabajo. El pago fue liberado ($%.2f).", job.ClientName, amount))
	return job, nil
}

// Cancel: any non-terminal state → cancelled (client or professional)
func (uc *JobUseCase) Cancel(clerkID, jobID string, reason string) (*domain.Job, error) {
	user, prof, err := uc.resolveUser(clerkID)
	if err != nil {
		return nil, err
	}
	job, err := uc.jobs.FindByID(jobID)
	if err != nil {
		return nil, err
	}
	if user.ID != job.ClientID && (prof == nil || prof.ID != job.ProfessionalID) {
		return nil, errors.New("forbidden")
	}
	if err := validateTransition(job.Status, domain.JobStatusCancelled); err != nil {
		return nil, err
	}
	now := time.Now()
	job.Status = domain.JobStatusCancelled
	job.CancelReason = reason
	job.CancelledAt = &now
	job, err = uc.jobs.Update(job)
	if err != nil {
		return nil, err
	}
	cancellerName := job.ClientName
	notifyID := job.ProfessionalUID
	if prof != nil && prof.ID == job.ProfessionalID {
		cancellerName = job.ProfessionalName
		notifyID = job.ClientID
	}
	uc.notify(notifyID, "job_cancelled",
		fmt.Sprintf("%s canceló el trabajo. Motivo: %s", cancellerName, reason))
	return job, nil
}

// --- helpers ---

func (uc *JobUseCase) resolveUser(clerkID string) (*domain.User, *domain.Professional, error) {
	user, err := uc.users.FindByClerkID(clerkID)
	if err != nil || user == nil {
		return nil, nil, errors.New("user not found")
	}
	prof, _ := uc.professionals.FindByUserID(user.ID)
	return user, prof, nil
}

func (uc *JobUseCase) canAccess(user *domain.User, job *domain.Job) bool {
	return user.ID == job.ClientID || user.ID == job.ProfessionalUID
}

func (uc *JobUseCase) notify(userID, notifType, message string) {
	if uc.notifications != nil && userID != "" {
		_ = uc.notifications.CreateForUser(userID, notifType, message)
	}
}

func validateTransition(from, to string) error {
	targets, ok := domain.ValidTransitions[from]
	if !ok {
		return fmt.Errorf("invalid current state: %s", from)
	}
	if !targets[to] {
		return fmt.Errorf("invalid transition: %s → %s", from, to)
	}
	return nil
}
