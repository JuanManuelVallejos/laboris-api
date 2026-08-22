package usecase

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/laboris/laboris-api/internal/domain"
	"github.com/laboris/laboris-api/internal/storage"
)

const maxPortfolioPhotoBytes = 5 << 20 // 5MB

// attachmentURLTTL is how long a signed read URL stays valid. Reused by
// professional.go/me.go when they sign PortfolioPhotos before responding.
const attachmentURLTTL = 24 * time.Hour

var allowedImageTypes = map[string]bool{
	"image/jpeg": true,
	"image/png":  true,
	"image/webp": true,
}

type AttachmentUseCase struct {
	attachments   domain.AttachmentRepository
	users         domain.UserRepository
	professionals domain.ProfessionalRepository
	storage       *storage.SupabaseClient
}

func NewAttachmentUseCase(
	attachments domain.AttachmentRepository,
	users domain.UserRepository,
	professionals domain.ProfessionalRepository,
	storageClient *storage.SupabaseClient,
) *AttachmentUseCase {
	return &AttachmentUseCase{
		attachments:   attachments,
		users:         users,
		professionals: professionals,
		storage:       storageClient,
	}
}

func (uc *AttachmentUseCase) UploadPortfolioPhoto(clerkID string, file io.Reader, filename string) (*domain.Attachment, error) {
	user, err := uc.users.FindByClerkID(clerkID)
	if err != nil || user == nil {
		return nil, errors.New("user not found")
	}
	prof, err := uc.professionals.FindByUserID(user.ID)
	if err != nil || prof == nil {
		return nil, errors.New("forbidden: only a professional can upload portfolio photos")
	}

	buf := make([]byte, maxPortfolioPhotoBytes+1)
	n, err := io.ReadFull(file, buf)
	if err != nil && err != io.ErrUnexpectedEOF && err != io.EOF {
		return nil, err
	}
	if n > maxPortfolioPhotoBytes {
		return nil, errors.New("file too large: max 5MB")
	}
	buf = buf[:n]

	contentType := http.DetectContentType(buf)
	if !allowedImageTypes[contentType] {
		return nil, fmt.Errorf("unsupported file type: %s (only jpeg, png or webp)", contentType)
	}

	ext := strings.ToLower(filepath.Ext(filename))
	if ext == "" {
		switch contentType {
		case "image/jpeg":
			ext = ".jpg"
		case "image/png":
			ext = ".png"
		case "image/webp":
			ext = ".webp"
		}
	}
	path := fmt.Sprintf("professional-portfolio/%s/%s%s", prof.ID, randomID(), ext)

	if err := uc.storage.Upload(context.Background(), path, bytes.NewReader(buf), contentType); err != nil {
		return nil, err
	}

	created, err := uc.attachments.Create(&domain.Attachment{
		Type:       domain.AttachmentTypeProfessionalPortfolio,
		OwnerID:    prof.ID,
		Path:       path,
		Filename:   filename,
		Extension:  strings.TrimPrefix(ext, "."),
		UploadedBy: user.ID,
	})
	if err != nil {
		return nil, err
	}

	signedURL, err := uc.storage.SignedURL(context.Background(), created.Path, attachmentURLTTL)
	if err != nil {
		return nil, err
	}
	created.URL = signedURL
	return created, nil
}

func (uc *AttachmentUseCase) DeletePortfolioPhoto(clerkID, attachmentID string) error {
	user, err := uc.users.FindByClerkID(clerkID)
	if err != nil || user == nil {
		return errors.New("user not found")
	}
	prof, err := uc.professionals.FindByUserID(user.ID)
	if err != nil || prof == nil {
		return errors.New("forbidden: only a professional can delete portfolio photos")
	}

	a, err := uc.attachments.FindByID(attachmentID)
	if err != nil {
		return err
	}
	if a == nil {
		return errors.New("attachment not found")
	}
	if a.Type != domain.AttachmentTypeProfessionalPortfolio || a.OwnerID != prof.ID {
		return errors.New("forbidden: this photo does not belong to your portfolio")
	}

	_ = uc.storage.Delete(context.Background(), a.Path)

	return uc.attachments.Delete(attachmentID)
}

// signPortfolioPhotos populates the (never-persisted) signed URL for each of
// a professional's portfolio photos. Best-effort: a single photo failing to
// sign doesn't fail the whole professional response, it's just left blank.
func signPortfolioPhotos(sc *storage.SupabaseClient, p *domain.Professional) {
	if sc == nil {
		return
	}
	for i := range p.PortfolioPhotos {
		url, err := sc.SignedURL(context.Background(), p.PortfolioPhotos[i].Path, attachmentURLTTL)
		if err == nil {
			p.PortfolioPhotos[i].URL = url
		}
	}
}

// randomID generates a random hex identifier for storage paths — stdlib
// only, no need for a dedicated uuid dependency for this single use.
func randomID() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}
