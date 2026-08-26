package postgres

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/laboris/laboris-api/internal/domain"
)

type AddressRepository struct {
	db *pgxpool.Pool
}

func NewAddressRepository(db *pgxpool.Pool) *AddressRepository {
	return &AddressRepository{db: db}
}

// hasActiveJobExpr es un domicilio con al menos un job en curso (estado no
// final) creado a partir de una solicitud que lo usó.
const hasActiveJobExpr = `EXISTS (
	SELECT 1 FROM requests rq
	JOIN jobs j ON j.request_id = rq.id
	WHERE rq.address_id = a.id AND j.status NOT IN ('completed', 'cancelled')
)`

func (r *AddressRepository) FindByUserID(userID string) ([]domain.Address, error) {
	rows, err := r.db.Query(context.Background(), `
		SELECT a.id, a.user_id, a.label, a.address, a.is_default, `+hasActiveJobExpr+`
		FROM addresses a
		WHERE a.user_id = $1
		ORDER BY a.is_default DESC, a.created_at ASC
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make([]domain.Address, 0)
	for rows.Next() {
		var a domain.Address
		if err := rows.Scan(&a.ID, &a.UserID, &a.Label, &a.Address, &a.IsDefault, &a.HasActiveJob); err != nil {
			return nil, err
		}
		result = append(result, a)
	}
	return result, nil
}

func (r *AddressRepository) FindByID(id string) (*domain.Address, error) {
	a := &domain.Address{}
	err := r.db.QueryRow(context.Background(), `
		SELECT id, user_id, label, address, lat, lng, is_default
		FROM addresses WHERE id = $1
	`, id).Scan(&a.ID, &a.UserID, &a.Label, &a.Address, &a.Lat, &a.Lng, &a.IsDefault)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return a, nil
}

func (r *AddressRepository) Create(a *domain.Address) (*domain.Address, error) {
	err := r.db.QueryRow(context.Background(), `
		INSERT INTO addresses (user_id, label, address, lat, lng, is_default)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id
	`, a.UserID, a.Label, a.Address, a.Lat, a.Lng, a.IsDefault).Scan(&a.ID)
	return a, err
}

func (r *AddressRepository) Update(id, label, address string, lat, lng float64) (*domain.Address, error) {
	a := &domain.Address{}
	err := r.db.QueryRow(context.Background(), `
		UPDATE addresses SET label = $2, address = $3, lat = $4, lng = $5
		WHERE id = $1
		RETURNING id, user_id, label, address, is_default
	`, id, label, address, lat, lng).Scan(&a.ID, &a.UserID, &a.Label, &a.Address, &a.IsDefault)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	return a, err
}

func (r *AddressRepository) Delete(id string) error {
	_, err := r.db.Exec(context.Background(), `DELETE FROM addresses WHERE id = $1`, id)
	return err
}

func (r *AddressRepository) SetDefault(userID, addressID string) error {
	tx, err := r.db.Begin(context.Background())
	if err != nil {
		return err
	}
	defer tx.Rollback(context.Background())

	if _, err := tx.Exec(context.Background(),
		`UPDATE addresses SET is_default = FALSE WHERE user_id = $1`, userID); err != nil {
		return err
	}
	if _, err := tx.Exec(context.Background(),
		`UPDATE addresses SET is_default = TRUE WHERE id = $1 AND user_id = $2`, addressID, userID); err != nil {
		return err
	}
	return tx.Commit(context.Background())
}

func (r *AddressRepository) ClearDefault(userID string) error {
	_, err := r.db.Exec(context.Background(),
		`UPDATE addresses SET is_default = FALSE WHERE user_id = $1`, userID)
	return err
}

func (r *AddressRepository) HasActiveJob(addressID string) (bool, error) {
	var has bool
	err := r.db.QueryRow(context.Background(), `
		SELECT EXISTS (
			SELECT 1 FROM requests rq
			JOIN jobs j ON j.request_id = rq.id
			WHERE rq.address_id = $1 AND j.status NOT IN ('completed', 'cancelled')
		)
	`, addressID).Scan(&has)
	return has, err
}

func (r *AddressRepository) CountByUserID(userID string) (int, error) {
	var count int
	err := r.db.QueryRow(context.Background(),
		`SELECT COUNT(*) FROM addresses WHERE user_id = $1`, userID).Scan(&count)
	return count, err
}
