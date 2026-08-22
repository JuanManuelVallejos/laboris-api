package domain

import "time"

const (
	AttachmentTypeProfessionalPortfolio = "professional_portfolio"
)

// Attachment is a generic, polymorphic reference to a file uploaded to
// storage (Supabase Storage today). Type + OwnerID identify what the file
// belongs to — e.g. AttachmentTypeProfessionalPortfolio + a professional's
// ID for a portfolio photo. Future types (request photos, review images,
// verification documents) reuse the same table without a schema change.
type Attachment struct {
	ID         string    `json:"id"`
	Type       string    `json:"type"`
	OwnerID    string    `json:"ownerId"`
	Path       string    `json:"-"`   // storage path, persisted, never serialized directly
	URL        string    `json:"url"` // signed URL, computed on read, never persisted
	Filename   string    `json:"filename"`
	Extension  string    `json:"extension"`
	UploadedBy string    `json:"uploadedBy"`
	CreatedAt  time.Time `json:"createdAt"`
}

type AttachmentRepository interface {
	Create(a *Attachment) (*Attachment, error)
	FindByTypeAndOwner(attachmentType, ownerID string) ([]Attachment, error)
	FindByID(id string) (*Attachment, error)
	Delete(id string) error
}
