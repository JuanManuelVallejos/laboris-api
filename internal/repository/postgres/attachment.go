package postgres

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/laboris/laboris-api/internal/domain"
)

type AttachmentRepository struct {
	db *pgxpool.Pool
}

func NewAttachmentRepository(db *pgxpool.Pool) *AttachmentRepository {
	return &AttachmentRepository{db: db}
}

func (r *AttachmentRepository) Create(a *domain.Attachment) (*domain.Attachment, error) {
	err := r.db.QueryRow(context.Background(), `
		INSERT INTO attachments (type, owner_id, path, filename, extension, uploaded_by)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, created_at
	`, a.Type, a.OwnerID, a.Path, a.Filename, a.Extension, a.UploadedBy).Scan(&a.ID, &a.CreatedAt)
	if err != nil {
		return nil, err
	}
	return a, nil
}

func (r *AttachmentRepository) FindByTypeAndOwner(attachmentType, ownerID string) ([]domain.Attachment, error) {
	rows, err := r.db.Query(context.Background(), `
		SELECT id, type, owner_id, path, filename, extension, uploaded_by, created_at
		FROM attachments
		WHERE type = $1 AND owner_id = $2
		ORDER BY created_at ASC
	`, attachmentType, ownerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make([]domain.Attachment, 0)
	for rows.Next() {
		var a domain.Attachment
		if err := rows.Scan(&a.ID, &a.Type, &a.OwnerID, &a.Path, &a.Filename, &a.Extension, &a.UploadedBy, &a.CreatedAt); err != nil {
			return nil, err
		}
		result = append(result, a)
	}
	return result, nil
}

func (r *AttachmentRepository) FindByID(id string) (*domain.Attachment, error) {
	a := &domain.Attachment{}
	err := r.db.QueryRow(context.Background(), `
		SELECT id, type, owner_id, path, filename, extension, uploaded_by, created_at
		FROM attachments
		WHERE id = $1
	`, id).Scan(&a.ID, &a.Type, &a.OwnerID, &a.Path, &a.Filename, &a.Extension, &a.UploadedBy, &a.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	return a, err
}

func (r *AttachmentRepository) Delete(id string) error {
	_, err := r.db.Exec(context.Background(), `DELETE FROM attachments WHERE id = $1`, id)
	return err
}
