package usecase

import (
	"errors"

	"github.com/laboris/laboris-api/internal/domain"
)

// ErrReviewNotEligible se devuelve si el cliente nunca tuvo un trabajo
// completado con este profesional — evita reseñas de gente que nunca lo
// contrató.
var ErrReviewNotEligible = errors.New("solo podés dejar una reseña si ya tuviste un trabajo completado con este profesional")

type ReviewUseCase struct {
	reviews       domain.ReviewRepository
	users         domain.UserRepository
	professionals domain.ProfessionalRepository
	jobs          domain.JobRepository
}

func NewReviewUseCase(reviews domain.ReviewRepository, users domain.UserRepository, professionals domain.ProfessionalRepository, jobs domain.JobRepository) *ReviewUseCase {
	return &ReviewUseCase{reviews: reviews, users: users, professionals: professionals, jobs: jobs}
}

// Create crea o actualiza (upsert) la reseña del cliente para este
// profesional — reenviar la actualiza en vez de duplicarla.
func (uc *ReviewUseCase) Create(clerkID, professionalID string, rating int, comment string) (*domain.Review, error) {
	if rating < 1 || rating > 5 {
		return nil, errors.New("el puntaje debe estar entre 1 y 5")
	}

	user, err := uc.users.FindByClerkID(clerkID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, ErrUserNotOnboarded
	}

	prof, err := uc.professionals.FindByID(professionalID)
	if err != nil {
		return nil, err
	}
	if prof == nil {
		return nil, errors.New("professional not found")
	}

	jobs, err := uc.jobs.FindByUserID(user.ID)
	if err != nil {
		return nil, err
	}
	eligible := false
	for _, j := range jobs {
		if j.ProfessionalID == professionalID && j.Status == domain.JobStatusCompleted {
			eligible = true
			break
		}
	}
	if !eligible {
		return nil, ErrReviewNotEligible
	}

	return uc.reviews.Upsert(&domain.Review{
		ProfessionalID: professionalID,
		ReviewerID:     user.ID,
		Rating:         rating,
		Comment:        comment,
	})
}

func (uc *ReviewUseCase) ListByProfessional(professionalID string) ([]domain.Review, error) {
	return uc.reviews.FindByProfessionalID(professionalID)
}
