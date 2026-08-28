package domain

import "time"

const (
	JobStatusPendingVisit        = "pending_visit"
	JobStatusVisitProposed       = "visit_proposed"
	JobStatusVisitScheduled      = "visit_scheduled"
	JobStatusVisitQuoted         = "visit_quoted"
	JobStatusVisitPaid           = "visit_paid"
	JobStatusVisitCompleted      = "visit_completed"
	JobStatusWorkQuoted          = "work_quoted"
	JobStatusWorkApproved        = "work_approved"
	JobStatusWorkInProgress      = "work_in_progress"
	JobStatusWorkDelivered       = "work_delivered"
	JobStatusReworkRequested     = "rework_requested"
	JobStatusReworkQuoted        = "rework_quoted"
	JobStatusReworkAccepted      = "rework_accepted"
	JobStatusReworkVisitProposed = "rework_visit_proposed"
	JobStatusCompleted           = "completed"
	JobStatusCancelled           = "cancelled"
)

// VisitConfirmed indica si, para este status, ya corresponde revelarle al
// profesional la dirección exacta del trabajo — recién una vez que el
// cliente confirmó una visita propuesta (con o sin cotización), o el
// trabajo se saltó la visita directamente. Antes de eso (sin trabajo
// todavía, o con una visita solo propuesta sin confirmar) se mantiene
// oculta. Ninguna transición vuelve a pending_visit/visit_proposed desde un
// estado posterior, así que el reveal es permanente para ese trabajo.
func VisitConfirmed(status string) bool {
	switch status {
	case "", JobStatusPendingVisit, JobStatusVisitProposed:
		return false
	default:
		return true
	}
}

// ValidTransitions defines allowed state transitions for a Job.
var ValidTransitions = map[string]map[string]bool{
	JobStatusPendingVisit:        {JobStatusVisitProposed: true, JobStatusWorkQuoted: true, JobStatusCancelled: true},
	JobStatusVisitProposed:       {JobStatusVisitScheduled: true, JobStatusVisitQuoted: true, JobStatusPendingVisit: true, JobStatusCancelled: true},
	JobStatusVisitScheduled:      {JobStatusVisitCompleted: true, JobStatusVisitQuoted: true, JobStatusCancelled: true},
	JobStatusVisitQuoted:         {JobStatusVisitPaid: true, JobStatusCancelled: true},
	JobStatusVisitPaid:           {JobStatusVisitCompleted: true, JobStatusCancelled: true},
	JobStatusVisitCompleted:      {JobStatusWorkQuoted: true, JobStatusCancelled: true},
	JobStatusWorkQuoted:          {JobStatusWorkApproved: true, JobStatusCancelled: true},
	JobStatusWorkApproved:        {JobStatusWorkInProgress: true, JobStatusCancelled: true},
	JobStatusWorkInProgress:      {JobStatusWorkDelivered: true, JobStatusCancelled: true},
	JobStatusWorkDelivered:       {JobStatusReworkRequested: true, JobStatusCompleted: true},
	JobStatusReworkRequested:     {JobStatusReworkQuoted: true, JobStatusReworkAccepted: true, JobStatusCancelled: true},
	JobStatusReworkQuoted:        {JobStatusReworkAccepted: true, JobStatusCancelled: true},
	JobStatusReworkAccepted:      {JobStatusReworkVisitProposed: true, JobStatusCancelled: true},
	JobStatusReworkVisitProposed: {JobStatusWorkInProgress: true, JobStatusReworkAccepted: true, JobStatusCancelled: true},
	JobStatusCompleted:           {},
	JobStatusCancelled:           {},
}

type Job struct {
	ID               string `json:"id"`
	RequestID        string `json:"requestId"`
	ClientID         string `json:"clientId"`
	ClientName       string `json:"clientName"`
	ProfessionalID   string `json:"professionalId"`
	ProfessionalName string `json:"professionalName"`
	ProfessionalUID  string `json:"-"` // professional's user_id — used for auth, not exposed
	// Address es el domicilio congelado al momento de crear la solicitud
	// (requests.address_snapshot) — no cambia aunque el cliente después
	// edite o borre ese domicilio guardado. Vacío en trabajos legacy. Para
	// el profesional queda recortado a lo sumo a nivel localidad hasta que
	// AddressRevealed sea true.
	Address string `json:"address,omitempty"`
	// AddressRevealed indica si Address trae el domicilio completo. Para el
	// cliente siempre es true; para el profesional, solo una vez confirmada
	// la visita (VisitConfirmed). Los objetos que se difunden por el Hub de
	// SSE siempre lo fuerzan a false — ver JobUseCase.updateJob.
	AddressRevealed   bool       `json:"addressRevealed"`
	Status            string     `json:"status"`
	VisitScheduledAt  *time.Time `json:"visitScheduledAt,omitempty"`
	VisitQuoteAmount  *float64   `json:"visitQuoteAmount,omitempty"`
	WorkQuoteAmount   *float64   `json:"workQuoteAmount,omitempty"`
	WorkDescription   string     `json:"workDescription,omitempty"`
	ReworkCount       int        `json:"reworkCount"`
	ReworkNotes       string     `json:"reworkNotes,omitempty"`
	ReworkQuoteAmount *float64   `json:"reworkQuoteAmount,omitempty"`
	CancelReason      string     `json:"cancelReason,omitempty"`
	CompletedAt       *time.Time `json:"completedAt,omitempty"`
	CancelledAt       *time.Time `json:"cancelledAt,omitempty"`
	WorkDeliveredAt   *time.Time `json:"workDeliveredAt,omitempty"`
	AutoCompleted     bool       `json:"autoCompleted"`
	// AutoCloseDeadline is computed on read (WorkDeliveredAt + configured grace
	// period) and never persisted — it's not a column, just a convenience for callers.
	AutoCloseDeadline *time.Time     `json:"autoCloseDeadline,omitempty"`
	Payments          []Payment      `json:"payments"`
	ReworkRecords     []ReworkRecord `json:"reworkRecords"`
	CreatedAt         time.Time      `json:"createdAt"`
	UpdatedAt         time.Time      `json:"updatedAt"`
	// ViewerIsClient / ViewerIsProfessional indican, para quien pidió este
	// job, si es el cliente o el profesional — se completan al leer (mismo
	// patrón que AutoCloseDeadline), no son columnas. Con doble rol, un
	// usuario puede tener el rol "professional" en general y aun así ser el
	// cliente de este job puntual, así que no alcanza con mirar sus roles
	// globales. No usar en objetos que se difunden por el Hub de SSE — ahí
	// el mismo job se comparte entre los dos participantes.
	ViewerIsClient       bool `json:"viewerIsClient"`
	ViewerIsProfessional bool `json:"viewerIsProfessional"`
}

type JobRepository interface {
	Create(j *Job) (*Job, error)
	FindByID(id string) (*Job, error)
	FindByUserID(userID string) ([]Job, error)
	FindByRequestID(requestID string) (*Job, error)
	FindOverdueDelivered(before time.Time) ([]Job, error)
	Update(j *Job) (*Job, error)
	// CountCompletedByProfessional se usa para las estadísticas de "Mi
	// actividad" — cuántos trabajos completó este profesional en total.
	CountCompletedByProfessional(professionalID string) (int, error)
	// CountCompletedByClient es el equivalente para "Mi actividad" del cliente.
	CountCompletedByClient(clientID string) (int, error)
}
