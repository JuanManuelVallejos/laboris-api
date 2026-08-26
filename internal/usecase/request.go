package usecase

import (
	"errors"
	"fmt"

	"github.com/laboris/laboris-api/internal/domain"
	"github.com/laboris/laboris-api/internal/storage"
)

type RequestUseCase struct {
	requests      domain.RequestRepository
	users         domain.UserRepository
	professionals domain.ProfessionalRepository
	addresses     domain.AddressRepository
	notifications *NotificationUseCase
	jobs          domain.JobRepository
	storage       *storage.SupabaseClient
	autoCloseDays int
}

func NewRequestUseCase(requests domain.RequestRepository, users domain.UserRepository, professionals domain.ProfessionalRepository) *RequestUseCase {
	return &RequestUseCase{requests: requests, users: users, professionals: professionals}
}

func (uc *RequestUseCase) SetNotifications(n *NotificationUseCase) {
	uc.notifications = n
}

func (uc *RequestUseCase) SetAddressRepository(addresses domain.AddressRepository) {
	uc.addresses = addresses
}

func (uc *RequestUseCase) SetJobRepository(jobs domain.JobRepository) {
	uc.jobs = jobs
}

func (uc *RequestUseCase) SetStorage(sc *storage.SupabaseClient) {
	uc.storage = sc
}

func (uc *RequestUseCase) SetAutoCloseDays(days int) {
	uc.autoCloseDays = days
}

var ErrSelfRequest = errors.New("no podés solicitar presupuesto a vos mismo")

func (uc *RequestUseCase) Create(clerkID, professionalID, description, addressID string) (*domain.Request, error) {
	user, err := uc.users.FindByClerkID(clerkID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, errors.New("user not found")
	}
	if !user.HasHomeAddress() {
		return nil, ErrAddressRequired
	}

	var addrID *string
	if uc.addresses != nil {
		addr, err := uc.addresses.FindByID(addressID)
		if err != nil {
			return nil, err
		}
		if addr == nil || addr.UserID != user.ID {
			return nil, ErrAddressNotFound
		}
		addrID = &addr.ID
	}

	if uc.professionals != nil {
		prof, err := uc.professionals.FindByID(professionalID)
		if err != nil {
			return nil, err
		}
		if prof != nil && prof.UserID == user.ID {
			return nil, ErrSelfRequest
		}
	}

	req, err := uc.requests.Create(&domain.Request{
		ClientID:       user.ID,
		ProfessionalID: professionalID,
		Description:    description,
		AddressID:      addrID,
	})
	if err != nil {
		return nil, err
	}

	if uc.notifications != nil && uc.professionals != nil {
		prof, lookupErr := uc.professionals.FindByID(professionalID)
		if lookupErr == nil && prof != nil {
			msg := fmt.Sprintf("Nueva solicitud de %s", user.FullName)
			_ = uc.notifications.CreateForUser(prof.UserID, "new_request", msg, req.ID)
		}
	}

	return req, nil
}

func (uc *RequestUseCase) ListReceivedByProfessional(clerkID string) ([]domain.Request, error) {
	user, err := uc.users.FindByClerkID(clerkID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, errors.New("user not found")
	}
	prof, err := uc.professionals.FindByUserID(user.ID)
	if err != nil {
		return nil, err
	}
	if prof == nil {
		return nil, errors.New("professional profile not found")
	}
	_, _ = AutoCloseOverdueJobs(uc.jobs, uc.notifications, uc.autoCloseDays)
	return uc.requests.FindByProfessionalID(prof.ID)
}

// GetReceivedRequestDetail is how a professional opens a single request they
// received. Seeing the full description + photos here is what marks the
// request "viewed" (pending → viewed) — listing the inbox no longer does.
func (uc *RequestUseCase) GetReceivedRequestDetail(clerkID, requestID string) (*domain.Request, error) {
	user, err := uc.users.FindByClerkID(clerkID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, errors.New("user not found")
	}
	prof, err := uc.professionals.FindByUserID(user.ID)
	if err != nil {
		return nil, err
	}
	if prof == nil {
		return nil, errors.New("professional profile not found")
	}
	rq, err := uc.requests.FindByID(requestID)
	if err != nil {
		return nil, err
	}
	if rq == nil {
		return nil, errors.New("request not found")
	}
	if rq.ProfessionalID != prof.ID {
		return nil, errors.New("forbidden")
	}
	if rq.Status == domain.RequestStatusPending {
		if err := uc.requests.MarkViewed(rq.ID); err == nil {
			rq.Status = domain.RequestStatusViewed
		}
	}
	signAttachments(uc.storage, rq.Photos)
	return rq, nil
}

func (uc *RequestUseCase) ListSentByClient(clerkID string) ([]domain.Request, error) {
	user, err := uc.users.FindByClerkID(clerkID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, errors.New("user not found")
	}
	_, _ = AutoCloseOverdueJobs(uc.jobs, uc.notifications, uc.autoCloseDays)
	return uc.requests.FindByClientID(user.ID)
}

func (uc *RequestUseCase) UpdateStatus(clerkID, id, status, reason string) (*domain.Request, error) {
	if status != domain.RequestStatusAccepted && status != domain.RequestStatusRejected {
		return nil, errors.New("invalid status")
	}
	if status == "rejected" && reason == "" {
		return nil, errors.New("rejection reason is required")
	}

	user, err := uc.users.FindByClerkID(clerkID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, errors.New("user not found")
	}
	prof, err := uc.professionals.FindByUserID(user.ID)
	if err != nil {
		return nil, err
	}
	if prof == nil {
		return nil, errors.New("forbidden: only a professional can accept or reject a request")
	}
	existing, err := uc.requests.FindByID(id)
	if err != nil {
		return nil, err
	}
	if existing == nil {
		return nil, errors.New("request not found")
	}
	if existing.ProfessionalID != prof.ID {
		return nil, errors.New("forbidden")
	}
	if existing.Status != domain.RequestStatusPending && existing.Status != domain.RequestStatusViewed {
		return nil, errors.New("this request was already resolved")
	}

	rq, err := uc.requests.UpdateStatus(id, status, reason)
	if err != nil {
		return nil, err
	}

	var jobID string
	if status == "accepted" && uc.jobs != nil {
		job, jerr := uc.jobs.Create(&domain.Job{
			RequestID:      rq.ID,
			ClientID:       rq.ClientID,
			ProfessionalID: rq.ProfessionalID,
		})
		if jerr == nil && job != nil {
			jobID = job.ID
			rq.JobID = jobID
		}
	}

	if uc.notifications != nil && uc.professionals != nil {
		prof, lookupErr := uc.professionals.FindByID(rq.ProfessionalID)
		if lookupErr == nil && prof != nil {
			var msg string
			if status == "accepted" {
				msg = fmt.Sprintf("%s aceptó tu solicitud", prof.Name)
			} else {
				msg = fmt.Sprintf("%s rechazó tu solicitud", prof.Name)
			}
			_ = uc.notifications.CreateForUser(rq.ClientID, "request_"+status, msg, jobID)
		}
	}

	return rq, nil
}
