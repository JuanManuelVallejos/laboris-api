package usecase

import (
	"errors"
	"fmt"
	"time"

	"github.com/laboris/laboris-api/internal/domain"
	"github.com/laboris/laboris-api/internal/realtime"
)

type JobUseCase struct {
	jobs          domain.JobRepository
	payments      domain.PaymentRepository
	users         domain.UserRepository
	professionals domain.ProfessionalRepository
	reworks       domain.ReworkRecordRepository
	notifications *NotificationUseCase
	hub           *realtime.Hub[*domain.Job]
	autoCloseDays int
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

func (uc *JobUseCase) SetHub(h *realtime.Hub[*domain.Job]) {
	uc.hub = h
}

func (uc *JobUseCase) SetAutoCloseDays(days int) {
	uc.autoCloseDays = days
}

// applyAutoCloseDeadline sets the (unpersisted) AutoCloseDeadline hint so the
// frontend can show a countdown without needing to know the configured days.
func applyAutoCloseDeadline(job *domain.Job, days int) {
	if days > 0 && job.Status == domain.JobStatusWorkDelivered && job.WorkDeliveredAt != nil {
		deadline := job.WorkDeliveredAt.Add(time.Duration(days) * 24 * time.Hour)
		job.AutoCloseDeadline = &deadline
	}
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
	_, _ = AutoCloseOverdueJobs(uc.jobs, uc.notifications, uc.autoCloseDays)
	job, err := uc.jobs.FindByID(jobID)
	if err != nil {
		return nil, err
	}
	if !uc.canAccess(user, job) {
		return nil, errors.New("forbidden")
	}
	applyAutoCloseDeadline(job, uc.autoCloseDays)
	job.ViewerIsClient = user.ID == job.ClientID
	job.ViewerIsProfessional = user.ID == job.ProfessionalUID
	applyJobAddressGating(job, job.ViewerIsClient)
	return job, nil
}

func (uc *JobUseCase) ListByUser(clerkID string) ([]domain.Job, error) {
	user, _, err := uc.resolveUser(clerkID)
	if err != nil {
		return nil, err
	}
	_, _ = AutoCloseOverdueJobs(uc.jobs, uc.notifications, uc.autoCloseDays)
	jobs, err := uc.jobs.FindByUserID(user.ID)
	if err != nil {
		return nil, err
	}
	for i := range jobs {
		applyAutoCloseDeadline(&jobs[i], uc.autoCloseDays)
		applyJobAddressGating(&jobs[i], jobs[i].ClientID == user.ID)
	}
	return jobs, nil
}

// ProfessionalStats resume la actividad de un profesional para la pestaña
// "Mi actividad": cuántos trabajos completó y cuánto ganó, mes a mes.
type ProfessionalStats struct {
	TotalCompleted  int                     `json:"totalCompleted"`
	TotalEarned     float64                 `json:"totalEarned"`
	MonthlyEarnings []domain.MonthlyEarning `json:"monthlyEarnings"`
}

func (uc *JobUseCase) GetProfessionalStats(clerkID string) (*ProfessionalStats, error) {
	_, prof, err := uc.resolveUser(clerkID)
	if err != nil {
		return nil, err
	}
	if prof == nil {
		return nil, errors.New("professional profile not found")
	}

	completed, err := uc.jobs.CountCompletedByProfessional(prof.ID)
	if err != nil {
		return nil, err
	}
	monthly, err := uc.payments.MonthlyEarningsByProfessional(prof.ID)
	if err != nil {
		return nil, err
	}

	stats := &ProfessionalStats{TotalCompleted: completed, MonthlyEarnings: monthly}
	for _, m := range monthly {
		stats.TotalEarned += m.Amount
	}
	return stats, nil
}

// ClientStats es el equivalente de ProfessionalStats para la pestaña "Mi
// actividad" del cliente: cuántos trabajos completó y cuánto gastó, mes a mes.
type ClientStats struct {
	TotalCompleted  int                     `json:"totalCompleted"`
	TotalSpent      float64                 `json:"totalSpent"`
	MonthlySpending []domain.MonthlyEarning `json:"monthlySpending"`
}

func (uc *JobUseCase) GetClientStats(clerkID string) (*ClientStats, error) {
	user, _, err := uc.resolveUser(clerkID)
	if err != nil {
		return nil, err
	}

	completed, err := uc.jobs.CountCompletedByClient(user.ID)
	if err != nil {
		return nil, err
	}
	monthly, err := uc.payments.MonthlySpendingByClient(user.ID)
	if err != nil {
		return nil, err
	}

	stats := &ClientStats{TotalCompleted: completed, MonthlySpending: monthly}
	for _, m := range monthly {
		stats.TotalSpent += m.Amount
	}
	return stats, nil
}

// applyJobAddressGating decide si Address queda completa o se recorta a lo
// sumo a nivel localidad — el cliente siempre ve completo, el profesional
// solo una vez confirmada la visita.
func applyJobAddressGating(job *domain.Job, isClient bool) {
	if isClient || domain.VisitConfirmed(job.Status) {
		job.AddressRevealed = true
		return
	}
	job.AddressRevealed = false
	job.Address = domain.CoarsenAddress(job.Address)
}

// ScheduleVisit: pending_visit → visit_proposed (professional proposes a date, optionally with a quote)
func (uc *JobUseCase) ScheduleVisit(clerkID, jobID string, scheduledAt time.Time, quoteAmount *float64) (*domain.Job, error) {
	user, prof, err := uc.resolveUser(clerkID)
	if err != nil {
		return nil, err
	}
	if !scheduledAt.After(time.Now()) {
		return nil, errors.New("scheduledAt must be in the future")
	}
	job, err := uc.jobs.FindByID(jobID)
	if err != nil {
		return nil, err
	}
	if prof == nil || prof.ID != job.ProfessionalID {
		return nil, errors.New("forbidden: only the professional can propose a visit date")
	}
	if err := validateTransition(job.Status, domain.JobStatusVisitProposed); err != nil {
		return nil, err
	}
	job.Status = domain.JobStatusVisitProposed
	job.VisitScheduledAt = &scheduledAt
	job.VisitQuoteAmount = quoteAmount
	job, err = uc.updateJob(job)
	if err != nil {
		return nil, err
	}
	// El que llama es el profesional y visit_proposed todavía no está
	// confirmado — su propia respuesta también queda con la dirección
	// recortada (updateJob solo garantiza esto para la copia del Hub).
	job.AddressRevealed = false
	job.Address = domain.CoarsenAddress(job.Address)
	msg := fmt.Sprintf("%s propuso una visita para el %s.", job.ProfessionalName, scheduledAt.Format("02/01 15:04"))
	if quoteAmount != nil {
		msg = fmt.Sprintf("%s propuso una visita para el %s, con un costo de $%.2f.", job.ProfessionalName, scheduledAt.Format("02/01 15:04"), *quoteAmount)
	}
	uc.notify(job.ClientID, "job_visit_proposed", msg+" Revisala y confirmala desde la app.", job.ID)
	_ = user
	return job, nil
}

// ConfirmVisit: visit_proposed → visit_scheduled (no quote proposed) or visit_quoted (quote proposed) —
// client confirms date and quote (if any) together, in a single approval step.
func (uc *JobUseCase) ConfirmVisit(clerkID, jobID string) (*domain.Job, error) {
	user, _, err := uc.resolveUser(clerkID)
	if err != nil {
		return nil, err
	}
	job, err := uc.jobs.FindByID(jobID)
	if err != nil {
		return nil, err
	}
	if user.ID != job.ClientID {
		return nil, errors.New("forbidden: only the client can confirm the visit")
	}
	target := domain.JobStatusVisitScheduled
	if job.VisitQuoteAmount != nil {
		target = domain.JobStatusVisitQuoted
	}
	if err := validateTransition(job.Status, target); err != nil {
		return nil, err
	}
	job.Status = target
	job, err = uc.updateJob(job)
	if err != nil {
		return nil, err
	}
	msg := fmt.Sprintf("%s confirmó la visita.", job.ClientName)
	if target == domain.JobStatusVisitQuoted {
		msg = fmt.Sprintf("%s confirmó la visita y la cotización ($%.2f). Queda pendiente el pago.", job.ClientName, *job.VisitQuoteAmount)
	}
	uc.notify(job.ProfessionalUID, "job_visit_confirmed", msg, job.ID)
	return job, nil
}

// DeclineVisit: visit_proposed → pending_visit (client asks for a different date/quote)
func (uc *JobUseCase) DeclineVisit(clerkID, jobID string) (*domain.Job, error) {
	user, _, err := uc.resolveUser(clerkID)
	if err != nil {
		return nil, err
	}
	job, err := uc.jobs.FindByID(jobID)
	if err != nil {
		return nil, err
	}
	if user.ID != job.ClientID {
		return nil, errors.New("forbidden: only the client can decline the visit")
	}
	if err := validateTransition(job.Status, domain.JobStatusPendingVisit); err != nil {
		return nil, err
	}
	job.Status = domain.JobStatusPendingVisit
	job.VisitScheduledAt = nil
	job.VisitQuoteAmount = nil
	job, err = uc.updateJob(job)
	if err != nil {
		return nil, err
	}
	uc.notify(job.ProfessionalUID, "job_visit_declined",
		fmt.Sprintf("%s rechazó la propuesta de visita. Proponé una nueva.", job.ClientName), job.ID)
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
	job, err = uc.updateJob(job)
	if err != nil {
		return nil, err
	}
	uc.notify(job.ClientID, "job_visit_quoted",
		fmt.Sprintf("%s envió la cotización de visita: $%.2f", job.ProfessionalName, amount), job.ID)
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
	job, err = uc.updateJob(job)
	if err != nil {
		return nil, err
	}
	uc.notify(job.ClientID, "job_work_quoted",
		fmt.Sprintf("%s envió cotización del trabajo: $%.2f", job.ProfessionalName, workAmount), job.ID)
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
	job, err = uc.updateJob(job)
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
		fmt.Sprintf("%s pagó la visita ($%.2f)", job.ClientName, amount), job.ID)
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
	job, err = uc.updateJob(job)
	if err != nil {
		return nil, err
	}
	uc.notify(job.ClientID, "job_visit_completed",
		fmt.Sprintf("%s confirmó que realizó la visita", job.ProfessionalName), job.ID)
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
	job, err = uc.updateJob(job)
	if err != nil {
		return nil, err
	}
	uc.notify(job.ClientID, "job_work_quoted",
		fmt.Sprintf("%s envió cotización del trabajo: $%.2f", job.ProfessionalName, amount), job.ID)
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
	job, err = uc.updateJob(job)
	if err != nil {
		return nil, err
	}
	uc.notify(job.ProfessionalUID, "job_work_approved",
		fmt.Sprintf("%s aprobó la cotización del trabajo", job.ClientName), job.ID)
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
	job, err = uc.updateJob(job)
	if err != nil {
		return nil, err
	}
	uc.notify(job.ClientID, "job_work_in_progress",
		fmt.Sprintf("%s comenzó el trabajo", job.ProfessionalName), job.ID)
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
	now := time.Now()
	job.WorkDeliveredAt = &now
	job, err = uc.updateJob(job)
	if err != nil {
		return nil, err
	}
	uc.notify(job.ClientID, "job_work_delivered",
		fmt.Sprintf("%s marcó el trabajo como entregado. Revisá y aprobá o pedí correcciones.", job.ProfessionalName), job.ID)
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
	if uc.reworks != nil {
		created, cerr := uc.reworks.Create(&domain.ReworkRecord{
			JobID:       job.ID,
			CycleNumber: job.ReworkCount,
			Notes:       notes,
		})
		if cerr == nil {
			job.ReworkRecords = append(job.ReworkRecords, *created)
		}
	}
	job, err = uc.updateJob(job)
	if err != nil {
		return nil, err
	}
	msg := fmt.Sprintf("%s solicitó correcciones (retrabajo #%d)", job.ClientName, job.ReworkCount)
	uc.notify(job.ProfessionalUID, "job_rework_requested", msg, job.ID)
	if job.ReworkCount > 2 {
		// notify admin users — for now notify professional as well with escalation flag
		uc.notify(job.ProfessionalUID, "job_rework_escalated",
			fmt.Sprintf("El trabajo lleva %d retrabajos. Un admin puede mediar si es necesario.", job.ReworkCount), job.ID)
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
	if uc.reworks != nil {
		_ = uc.reworks.UpdateQuoteAmount(job.ID, job.ReworkCount, amount)
	}
	for i := range job.ReworkRecords {
		if job.ReworkRecords[i].CycleNumber == job.ReworkCount {
			job.ReworkRecords[i].QuoteAmount = &amount
			break
		}
	}
	job, err = uc.updateJob(job)
	if err != nil {
		return nil, err
	}
	uc.notify(job.ClientID, "job_rework_quoted",
		fmt.Sprintf("%s cotizó las correcciones en $%.2f. Aprobá para retomar el trabajo.", job.ProfessionalName, amount), job.ID)
	return job, nil
}

// ApproveReworkQuote: rework_quoted → rework_accepted (client)
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
	if err := validateTransition(job.Status, domain.JobStatusReworkAccepted); err != nil {
		return nil, err
	}
	job.Status = domain.JobStatusReworkAccepted
	job, err = uc.updateJob(job)
	if err != nil {
		return nil, err
	}
	uc.notify(job.ProfessionalUID, "job_rework_quote_approved",
		fmt.Sprintf("%s aprobó la cotización de correcciones. Proponé una fecha para retomar el trabajo.", job.ClientName), job.ID)
	return job, nil
}

// AcceptRework: rework_requested → rework_accepted (professional, no extra cost)
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
	if err := validateTransition(job.Status, domain.JobStatusReworkAccepted); err != nil {
		return nil, err
	}
	job.Status = domain.JobStatusReworkAccepted
	job.ReworkQuoteAmount = nil // no extra cost — clear any amount from a previous cycle
	if uc.reworks != nil {
		_ = uc.reworks.MarkNoCharge(job.ID, job.ReworkCount)
	}
	for i := range job.ReworkRecords {
		if job.ReworkRecords[i].CycleNumber == job.ReworkCount {
			job.ReworkRecords[i].NoCharge = true
			break
		}
	}
	job, err = uc.updateJob(job)
	if err != nil {
		return nil, err
	}
	uc.notify(job.ClientID, "job_rework_accepted",
		fmt.Sprintf("%s aceptó las correcciones sin costo extra. Va a proponer una fecha para retomar el trabajo.", job.ProfessionalName), job.ID)
	return job, nil
}

// ScheduleReworkVisit: rework_accepted → rework_visit_proposed (professional proposes a date)
func (uc *JobUseCase) ScheduleReworkVisit(clerkID, jobID string, scheduledAt time.Time) (*domain.Job, error) {
	_, prof, err := uc.resolveUser(clerkID)
	if err != nil {
		return nil, err
	}
	if !scheduledAt.After(time.Now()) {
		return nil, errors.New("scheduledAt must be in the future")
	}
	job, err := uc.jobs.FindByID(jobID)
	if err != nil {
		return nil, err
	}
	if prof == nil || prof.ID != job.ProfessionalID {
		return nil, errors.New("forbidden: only the professional can propose a rework date")
	}
	if err := validateTransition(job.Status, domain.JobStatusReworkVisitProposed); err != nil {
		return nil, err
	}
	job.Status = domain.JobStatusReworkVisitProposed
	if uc.reworks != nil {
		_ = uc.reworks.UpdateScheduledAt(job.ID, job.ReworkCount, &scheduledAt)
	}
	for i := range job.ReworkRecords {
		if job.ReworkRecords[i].CycleNumber == job.ReworkCount {
			job.ReworkRecords[i].ScheduledAt = &scheduledAt
			break
		}
	}
	job, err = uc.updateJob(job)
	if err != nil {
		return nil, err
	}
	uc.notify(job.ClientID, "job_rework_visit_proposed",
		fmt.Sprintf("%s propuso el %s para retomar el trabajo. Confirmalo desde la app.", job.ProfessionalName, scheduledAt.Format("02/01 15:04")), job.ID)
	return job, nil
}

// ConfirmReworkVisit: rework_visit_proposed → work_in_progress (client confirms the proposed date)
func (uc *JobUseCase) ConfirmReworkVisit(clerkID, jobID string) (*domain.Job, error) {
	user, _, err := uc.resolveUser(clerkID)
	if err != nil {
		return nil, err
	}
	job, err := uc.jobs.FindByID(jobID)
	if err != nil {
		return nil, err
	}
	if user.ID != job.ClientID {
		return nil, errors.New("forbidden: only the client can confirm the rework date")
	}
	if err := validateTransition(job.Status, domain.JobStatusWorkInProgress); err != nil {
		return nil, err
	}
	job.Status = domain.JobStatusWorkInProgress
	job, err = uc.updateJob(job)
	if err != nil {
		return nil, err
	}
	uc.notify(job.ProfessionalUID, "job_rework_visit_confirmed",
		fmt.Sprintf("%s confirmó la fecha del retrabajo.", job.ClientName), job.ID)
	return job, nil
}

// DeclineReworkVisit: rework_visit_proposed → rework_accepted (client asks for a different date)
func (uc *JobUseCase) DeclineReworkVisit(clerkID, jobID string) (*domain.Job, error) {
	user, _, err := uc.resolveUser(clerkID)
	if err != nil {
		return nil, err
	}
	job, err := uc.jobs.FindByID(jobID)
	if err != nil {
		return nil, err
	}
	if user.ID != job.ClientID {
		return nil, errors.New("forbidden: only the client can decline the rework date")
	}
	if err := validateTransition(job.Status, domain.JobStatusReworkAccepted); err != nil {
		return nil, err
	}
	job.Status = domain.JobStatusReworkAccepted
	if uc.reworks != nil {
		_ = uc.reworks.UpdateScheduledAt(job.ID, job.ReworkCount, nil)
	}
	for i := range job.ReworkRecords {
		if job.ReworkRecords[i].CycleNumber == job.ReworkCount {
			job.ReworkRecords[i].ScheduledAt = nil
			break
		}
	}
	job, err = uc.updateJob(job)
	if err != nil {
		return nil, err
	}
	uc.notify(job.ProfessionalUID, "job_rework_visit_declined",
		fmt.Sprintf("%s rechazó la fecha propuesta para el retrabajo. Proponé una nueva.", job.ClientName), job.ID)
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
	job, err = uc.updateJob(job)
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
		fmt.Sprintf("%s aprobó el trabajo. El pago fue liberado ($%.2f).", job.ClientName, amount), job.ID)
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
	job, err = uc.updateJob(job)
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
		fmt.Sprintf("%s canceló el trabajo. Motivo: %s", cancellerName, reason), job.ID)
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

func (uc *JobUseCase) notify(userID, notifType, message, entityID string) {
	if uc.notifications != nil && userID != "" {
		_ = uc.notifications.CreateForUser(userID, notifType, message, entityID)
	}
}

// updateJob persists the job and, on success, publishes the new state to any
// live subscribers of its detail page — the single choke point every
// transition goes through, so every actionable propagates in real time for
// free without each method having to remember to publish.
func (uc *JobUseCase) updateJob(job *domain.Job) (*domain.Job, error) {
	job, err := uc.jobs.Update(job)
	if err != nil {
		return nil, err
	}
	if uc.hub != nil {
		// El mismo objeto se comparte entre cliente y profesional (ver
		// realtime.Hub) — no hay forma de mandar un Address distinto según
		// quién lo reciba, así que la copia publicada siempre usa la
		// versión más restrictiva. El frontend nunca confía en este campo
		// al llegar por SSE; lo vuelve a pedir por REST cuando hace falta.
		broadcast := *job
		broadcast.Address = domain.CoarsenAddress(job.Address)
		broadcast.AddressRevealed = false
		uc.hub.Publish(broadcast.ID, &broadcast)
	}
	return job, nil
}

func (uc *JobUseCase) Subscribe(clerkID, jobID string) (<-chan *domain.Job, func(), error) {
	user, _, err := uc.resolveUser(clerkID)
	if err != nil {
		return nil, nil, err
	}
	job, err := uc.jobs.FindByID(jobID)
	if err != nil {
		return nil, nil, err
	}
	if !uc.canAccess(user, job) {
		return nil, nil, errors.New("forbidden")
	}
	if uc.hub == nil {
		return nil, nil, errors.New("realtime not configured")
	}
	ch, unsub := uc.hub.Subscribe(jobID)
	return ch, unsub, nil
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
