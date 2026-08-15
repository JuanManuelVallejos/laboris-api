package postgres

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/laboris/laboris-api/internal/domain"
)

type NotificationRepository struct {
	db *pgxpool.Pool
}

func NewNotificationRepository(db *pgxpool.Pool) *NotificationRepository {
	return &NotificationRepository{db: db}
}

func (r *NotificationRepository) Create(n *domain.Notification) (*domain.Notification, error) {
	// entity_id is a UUID column — binding it through NULLIF($4, '') makes
	// Postgres try to cast the literal '' to uuid at prepare time and fail
	// with "invalid input syntax for type uuid" on every insert. Pass a real
	// nil instead so pgx sends SQL NULL directly when there's no entity.
	var entityIDParam *string
	if n.EntityID != "" {
		entityIDParam = &n.EntityID
	}
	var entityID *string
	err := r.db.QueryRow(context.Background(), `
		INSERT INTO notifications (user_id, type, message, entity_id)
		VALUES ($1, $2, $3, $4)
		RETURNING id, user_id, type, message, entity_id, read, created_at
	`, n.UserID, n.Type, n.Message, entityIDParam).Scan(&n.ID, &n.UserID, &n.Type, &n.Message, &entityID, &n.Read, &n.CreatedAt)
	if entityID != nil {
		n.EntityID = *entityID
	}
	return n, err
}

func (r *NotificationRepository) FindByUserID(userID string) ([]domain.Notification, error) {
	rows, err := r.db.Query(context.Background(), `
		SELECT id, user_id, type, message, entity_id, read, created_at
		FROM notifications
		WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT 50
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make([]domain.Notification, 0)
	for rows.Next() {
		var n domain.Notification
		var entityID *string
		if err := rows.Scan(&n.ID, &n.UserID, &n.Type, &n.Message, &entityID, &n.Read, &n.CreatedAt); err != nil {
			return nil, err
		}
		if entityID != nil {
			n.EntityID = *entityID
		}
		result = append(result, n)
	}
	return result, nil
}

func (r *NotificationRepository) CountUnread(userID string) (int, error) {
	var count int
	err := r.db.QueryRow(context.Background(), `
		SELECT COUNT(*) FROM notifications WHERE user_id = $1 AND read = false
	`, userID).Scan(&count)
	return count, err
}

func (r *NotificationRepository) MarkAllRead(userID string) error {
	_, err := r.db.Exec(context.Background(), `
		UPDATE notifications SET read = true WHERE user_id = $1 AND read = false
	`, userID)
	return err
}
