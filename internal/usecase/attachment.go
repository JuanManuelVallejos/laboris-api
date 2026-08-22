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
// professional.go/me.go/request.go when they sign an entity's photos before
// responding.
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
	requests      domain.RequestRepository
	storage       *storage.SupabaseClient
}

func NewAttachmentUseCase(
	attachments domain.AttachmentRepository,
	users domain.UserRepository,
	professionals domain.ProfessionalRepository,
	requests domain.RequestRepository,
	storageClient *storage.SupabaseClient,
) *AttachmentUseCase {
	return &AttachmentUseCase{
		attachments:   attachments,
		users:         users,
		professionals: professionals,
		requests:      requests,
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

	pathPrefix := fmt.Sprintf("professional-portfolio/%s", prof.ID)
	return uc.uploadAndCreate(file, filename, domain.AttachmentTypeProfessionalPortfolio, prof.ID, user.ID, pathPrefix)
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

// UploadRequestPhoto lets the client who created a Request attach a photo to
// it, only while the request hasn't been reviewed yet (status "pending") —
// once the professional has opened it, the attachments the decision was
// based on shouldn't change out from under them.
func (uc *AttachmentUseCase) UploadRequestPhoto(clerkID, requestID string, file io.Reader, filename string) (*domain.Attachment, error) {
	user, err := uc.users.FindByClerkID(clerkID)
	if err != nil || user == nil {
		return nil, errors.New("user not found")
	}
	req, err := uc.requests.FindByID(requestID)
	if err != nil {
		return nil, err
	}
	if req == nil {
		return nil, errors.New("request not found")
	}
	if req.ClientID != user.ID {
		return nil, errors.New("forbidden: only the request's author can attach photos")
	}
	if req.Status != domain.RequestStatusPending {
		return nil, errors.New("cannot attach photos after the request has been reviewed")
	}

	pathPrefix := fmt.Sprintf("request-photos/%s", req.ID)
	return uc.uploadAndCreate(file, filename, domain.AttachmentTypeRequestPhoto, req.ID, user.ID, pathPrefix)
}

// uploadAndCreate validates the file, uploads it to storage under
// {pathPrefix}/{randomID}{ext}, persists the Attachment row, and returns it
// with a freshly signed URL. Shared by every "attach a photo to X" flow.
func (uc *AttachmentUseCase) uploadAndCreate(file io.Reader, filename, attachmentType, ownerID, uploadedBy, pathPrefix string) (*domain.Attachment, error) {
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
	path := fmt.Sprintf("%s/%s%s", pathPrefix, randomID(), ext)

	if err := uc.storage.Upload(context.Background(), path, bytes.NewReader(buf), contentType); err != nil {
		return nil, err
	}

	created, err := uc.attachments.Create(&domain.Attachment{
		Type:       attachmentType,
		OwnerID:    ownerID,
		Path:       path,
		Filename:   filename,
		Extension:  strings.TrimPrefix(ext, "."),
		UploadedBy: uploadedBy,
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

// signAttachments populates the (never-persisted) signed URL for each
// attachment in photos, in place. Best-effort: a single photo failing to
// sign doesn't fail the whole response, it's just left blank.
func signAttachments(sc *storage.SupabaseClient, photos []domain.Attachment) {
	if sc == nil {
		return
	}
	for i := range photos {
		url, err := sc.SignedURL(context.Background(), photos[i].Path, attachmentURLTTL)
		if err == nil {
			photos[i].URL = url
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
