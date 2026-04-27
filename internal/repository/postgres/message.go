package postgres

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/laboris/laboris-api/internal/domain"
)

type MessageRepository struct {
	db *pgxpool.Pool
}

func NewMessageRepository(db *pgxpool.Pool) *MessageRepository {
	return &MessageRepository{db: db}
}

func (r *MessageRepository) Create(m *domain.Message) (*domain.Message, error) {
	err := r.db.QueryRow(context.Background(), `
		INSERT INTO messages (request_id, sender_id, sender_name, content)
		VALUES ($1, $2, $3, $4)
		RETURNING id, is_unread_for_client, is_unread_for_provider, created_at
	`, m.RequestID, m.SenderID, m.SenderName, m.Content,
	).Scan(&m.ID, &m.IsUnreadForClient, &m.IsUnreadForProvider, &m.CreatedAt)
	return m, err
}

func (r *MessageRepository) FindByRequestID(requestID string) ([]domain.Message, error) {
	rows, err := r.db.Query(context.Background(), `
		SELECT id, request_id, sender_id, sender_name, content,
		       is_unread_for_client, is_unread_for_provider, created_at
		FROM messages
		WHERE request_id = $1
		ORDER BY created_at ASC
	`, requestID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var msgs []domain.Message
	for rows.Next() {
		var m domain.Message
		if err := rows.Scan(&m.ID, &m.RequestID, &m.SenderID, &m.SenderName, &m.Content,
			&m.IsUnreadForClient, &m.IsUnreadForProvider, &m.CreatedAt); err != nil {
			return nil, err
		}
		msgs = append(msgs, m)
	}
	if msgs == nil {
		msgs = []domain.Message{}
	}
	return msgs, nil
}

func (r *MessageRepository) MarkReadForClient(requestID string) error {
	_, err := r.db.Exec(context.Background(),
		`UPDATE messages SET is_unread_for_client = FALSE WHERE request_id = $1`, requestID)
	return err
}

func (r *MessageRepository) MarkReadForProvider(requestID string) error {
	_, err := r.db.Exec(context.Background(),
		`UPDATE messages SET is_unread_for_provider = FALSE WHERE request_id = $1`, requestID)
	return err
}
