package postgres

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/laboris/laboris-api/internal/domain"
)

type ProfessionalRepository struct {
	db *pgxpool.Pool
}

func NewProfessionalRepository(db *pgxpool.Pool) *ProfessionalRepository {
	return &ProfessionalRepository{db: db}
}

func (r *ProfessionalRepository) FindAll() ([]domain.Professional, error) {
	rows, err := r.db.Query(context.Background(), `
		SELECT p.id, p.user_id, u.full_name, p.trade, p.zone, p.bio, p.verified, p.status,
		       COALESCE(AVG(rv.rating), 0) AS rating
		FROM professionals p
		JOIN users u ON u.id = p.user_id
		LEFT JOIN reviews rv ON rv.professional_id = p.id
		WHERE p.status = 'active'
		GROUP BY p.id, u.full_name
		ORDER BY rating DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make([]domain.Professional, 0)
	for rows.Next() {
		var p domain.Professional
		if err := rows.Scan(&p.ID, &p.UserID, &p.Name, &p.Trade, &p.Zone, &p.Bio, &p.Verified, &p.Status, &p.Rating); err != nil {
			return nil, err
		}
		result = append(result, p)
	}
	return result, nil
}

func (r *ProfessionalRepository) FindByID(id string) (*domain.Professional, error) {
	p := &domain.Professional{}
	err := r.db.QueryRow(context.Background(), `
		SELECT p.id, p.user_id, u.full_name, p.trade, p.zone, p.bio, p.verified, p.status,
		       COALESCE(AVG(rv.rating), 0) AS rating
		FROM professionals p
		JOIN users u ON u.id = p.user_id
		LEFT JOIN reviews rv ON rv.professional_id = p.id
		WHERE p.id = $1
		GROUP BY p.id, u.full_name
	`, id).Scan(&p.ID, &p.UserID, &p.Name, &p.Trade, &p.Zone, &p.Bio, &p.Verified, &p.Status, &p.Rating)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	return p, err
}

func (r *ProfessionalRepository) FindByUserID(userID string) (*domain.Professional, error) {
	p := &domain.Professional{}
	err := r.db.QueryRow(context.Background(), `
		SELECT p.id, p.user_id, u.full_name, p.trade, p.zone, p.bio, p.verified, p.status,
		       COALESCE(AVG(rv.rating), 0) AS rating
		FROM professionals p
		JOIN users u ON u.id = p.user_id
		LEFT JOIN reviews rv ON rv.professional_id = p.id
		WHERE p.user_id = $1
		GROUP BY p.id, u.full_name
	`, userID).Scan(&p.ID, &p.UserID, &p.Name, &p.Trade, &p.Zone, &p.Bio, &p.Verified, &p.Status, &p.Rating)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	return p, err
}

func (r *ProfessionalRepository) UpdateByUserID(userID, trade, zone, bio string) (*domain.Professional, error) {
	p := &domain.Professional{}
	err := r.db.QueryRow(context.Background(), `
		UPDATE professionals SET trade = $2, zone = $3, bio = $4
		WHERE user_id = $1
		RETURNING id, user_id, trade, zone, bio, verified, status
	`, userID, trade, zone, bio).Scan(&p.ID, &p.UserID, &p.Trade, &p.Zone, &p.Bio, &p.Verified, &p.Status)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	return p, err
}

func (r *ProfessionalRepository) Create(p *domain.Professional) (*domain.Professional, error) {
	err := r.db.QueryRow(context.Background(),
		`INSERT INTO professionals (user_id, trade, zone, bio) VALUES ($1, $2, $3, $4)
		 ON CONFLICT (user_id) DO UPDATE SET trade = EXCLUDED.trade, zone = EXCLUDED.zone, bio = EXCLUDED.bio
		 RETURNING id, user_id, trade, zone, bio, verified, status`,
		p.UserID, p.Trade, p.Zone, p.Bio,
	).Scan(&p.ID, &p.UserID, &p.Trade, &p.Zone, &p.Bio, &p.Verified, &p.Status)
	return p, err
}

func (r *ProfessionalRepository) FindAllPaginated(page, limit int) ([]domain.Professional, int64, error) {
	offset := (page - 1) * limit

	var total int64
	if err := r.db.QueryRow(context.Background(), `SELECT COUNT(*) FROM professionals`).Scan(&total); err != nil {
		return nil, 0, err
	}

	rows, err := r.db.Query(context.Background(), `
		SELECT p.id, p.user_id, u.full_name, p.trade, p.zone, p.bio, p.verified, p.status,
		       COALESCE(AVG(rv.rating), 0) AS rating
		FROM professionals p
		JOIN users u ON u.id = p.user_id
		LEFT JOIN reviews rv ON rv.professional_id = p.id
		GROUP BY p.id, u.full_name
		ORDER BY p.status ASC, u.full_name ASC
		LIMIT $1 OFFSET $2
	`, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	result := make([]domain.Professional, 0)
	for rows.Next() {
		var p domain.Professional
		if err := rows.Scan(&p.ID, &p.UserID, &p.Name, &p.Trade, &p.Zone, &p.Bio, &p.Verified, &p.Status, &p.Rating); err != nil {
			return nil, 0, err
		}
		result = append(result, p)
	}
	return result, total, nil
}

func (r *ProfessionalRepository) SetVerified(id string, verified bool) error {
	_, err := r.db.Exec(context.Background(),
		`UPDATE professionals SET verified = $2 WHERE id = $1`, id, verified)
	return err
}

func (r *ProfessionalRepository) SetStatus(id string, status string) error {
	_, err := r.db.Exec(context.Background(),
		`UPDATE professionals SET status = $2 WHERE id = $1`, id, status)
	return err
}

func (r *ProfessionalRepository) Delete(id string) error {
	_, err := r.db.Exec(context.Background(),
		`DELETE FROM professionals WHERE id = $1`, id)
	return err
}
