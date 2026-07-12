package usecase

import (
	"time"

	"github.com/laboris/laboris-api/internal/domain"
)

// AutoCloseOverdueJobs marks jobs stuck in work_delivered as completed once the
// configured grace period since delivery has elapsed. It takes no usecase-level
// state — only interfaces and primitives — so it can be invoked today as a side
// effect of listing jobs/requests, and later lifted unchanged into a real
// scheduled trigger (cron, lambda, etc.) once one exists.
func AutoCloseOverdueJobs(jobs domain.JobRepository, notifications *NotificationUseCase, days int) (int, error) {
	if days <= 0 || jobs == nil {
		return 0, nil
	}
	cutoff := time.Now().Add(-time.Duration(days) * 24 * time.Hour)
	overdue, err := jobs.FindOverdueDelivered(cutoff)
	if err != nil {
		return 0, err
	}
	closed := 0
	for i := range overdue {
		job := &overdue[i]
		job.Status = domain.JobStatusCompleted
		now := time.Now()
		job.CompletedAt = &now
		job.AutoCompleted = true
		if _, err := jobs.Update(job); err != nil {
			continue
		}
		closed++
		if notifications != nil {
			msg := "El trabajo se marcó como completado automáticamente porque no se confirmó la entrega a tiempo."
			_ = notifications.CreateForUser(job.ClientID, "job_auto_completed", msg)
			_ = notifications.CreateForUser(job.ProfessionalUID, "job_auto_completed", msg)
		}
	}
	return closed, nil
}
