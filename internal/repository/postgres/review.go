package postgres

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/laboris/laboris-api/internal/domain"
)

type ReviewRepository struct {
	db *pgxpool.Pool
}

func NewReviewRepository(db *pgxpool.Pool) *ReviewRepository {
	return &ReviewRepository{db: db}
}

func (r *ReviewRepository) Upsert(rv *domain.Review) (*domain.Review, error) {
	err := r.db.QueryRow(context.Background(), `
		INSERT INTO reviews (professional_id, reviewer_id, rating, comment)
		VALUES ($1, $2, $3, NULLIF($4, ''))
		ON CONFLICT (professional_id, reviewer_id) DO UPDATE SET
			rating = EXCLUDED.rating, comment = EXCLUDED.comment, created_at = NOW()
		RETURNING id, created_at
	`, rv.ProfessionalID, rv.ReviewerID, rv.Rating, rv.Comment,
	).Scan(&rv.ID, &rv.CreatedAt)
	return rv, err
}

func (r *ReviewRepository) FindByProfessionalID(professionalID string) ([]domain.Review, error) {
	rows, err := r.db.Query(context.Background(), `
		SELECT rv.id, rv.professional_id, u.full_name, rv.rating, COALESCE(rv.comment,''), rv.created_at
		FROM reviews rv
		JOIN users u ON u.id = rv.reviewer_id
		WHERE rv.professional_id = $1
		ORDER BY rv.created_at DESC
	`, professionalID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make([]domain.Review, 0)
	for rows.Next() {
		var rv domain.Review
		if err := rows.Scan(&rv.ID, &rv.ProfessionalID, &rv.ReviewerName, &rv.Rating, &rv.Comment, &rv.CreatedAt); err != nil {
			return nil, err
		}
		result = append(result, rv)
	}
	return result, nil
}
