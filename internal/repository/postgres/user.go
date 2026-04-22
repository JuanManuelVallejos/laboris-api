package postgres

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/laboris/laboris-api/internal/domain"
)

type UserRepository struct {
	db *pgxpool.Pool
}

func NewUserRepository(db *pgxpool.Pool) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) FindByClerkID(clerkID string) (*domain.User, error) {
	u := &domain.User{}
	err := r.db.QueryRow(context.Background(),
		`SELECT id, clerk_id, email, full_name, created_at FROM users WHERE clerk_id = $1`,
		clerkID,
	).Scan(&u.ID, &u.ClerkID, &u.Email, &u.FullName, &u.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return u, nil
}

func (r *UserRepository) Create(user *domain.User) (*domain.User, error) {
	err := r.db.QueryRow(context.Background(),
		`INSERT INTO users (clerk_id, email, full_name) VALUES ($1, $2, $3)
		 RETURNING id, clerk_id, email, full_name, created_at`,
		user.ClerkID, user.Email, user.FullName,
	).Scan(&user.ID, &user.ClerkID, &user.Email, &user.FullName, &user.CreatedAt)
	return user, err
}

func (r *UserRepository) AddRole(userID string, role string) error {
	_, err := r.db.Exec(context.Background(),
		`INSERT INTO user_roles (user_id, role) VALUES ($1, $2) ON CONFLICT DO NOTHING`,
		userID, role,
	)
	return err
}
