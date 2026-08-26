package usecase

import (
	"errors"

	"github.com/laboris/laboris-api/internal/domain"
	"github.com/laboris/laboris-api/internal/storage"
)

// ErrAddressRequired se devuelve cuando una acción necesita el domicilio del
// usuario (para calcular distancia) y todavía no lo cargó.
var ErrAddressRequired = errors.New("necesitás cargar tu domicilio antes de continuar")

type ProfessionalUseCase interface {
	GetAll(clerkID string) ([]domain.Professional, error)
	GetByID(id string) (*domain.Professional, error)
}

type professionalUseCase struct {
	repo    domain.ProfessionalRepository
	users   domain.UserRepository
	storage *storage.SupabaseClient
}

func NewProfessionalUseCase(repo domain.ProfessionalRepository, users domain.UserRepository, storageClient *storage.SupabaseClient) ProfessionalUseCase {
	return &professionalUseCase{repo: repo, users: users, storage: storageClient}
}

// GetAll devuelve los profesionales cuyo radio de alcance cubre el domicilio
// de quien pregunta — reemplaza al viejo listado por zona. Requiere que el
// usuario ya haya cargado su domicilio.
func (uc *professionalUseCase) GetAll(clerkID string) ([]domain.Professional, error) {
	user, err := uc.users.FindByClerkID(clerkID)
	if err != nil {
		return nil, err
	}
	if user == nil || !user.HasHomeAddress() {
		return nil, ErrAddressRequired
	}
	return uc.repo.FindNear(*user.HomeLat, *user.HomeLng)
}

func (uc *professionalUseCase) GetByID(id string) (*domain.Professional, error) {
	p, err := uc.repo.FindByID(id)
	if err != nil || p == nil {
		return p, err
	}
	signAttachments(uc.storage, p.PortfolioPhotos)
	return p, nil
}
