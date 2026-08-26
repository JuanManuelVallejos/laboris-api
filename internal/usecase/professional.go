package usecase

import (
	"errors"

	"github.com/laboris/laboris-api/internal/domain"
	"github.com/laboris/laboris-api/internal/storage"
)

// ErrAddressRequired se devuelve cuando una acción necesita el domicilio del
// usuario (para calcular distancia) y todavía no lo cargó.
var ErrAddressRequired = errors.New("necesitás cargar tu domicilio antes de continuar")

// ErrUserNotOnboarded se devuelve cuando el clerkID no tiene ninguna fila en
// nuestra base (a diferencia de ErrAddressRequired, que asume que el usuario
// existe y solo le falta el domicilio) — pasa cuando Clerk cree que el
// onboarding está completo pero nunca llegó a crearse del lado nuestro.
var ErrUserNotOnboarded = errors.New("todavía no completaste el registro")

type ProfessionalUseCase interface {
	GetAll(clerkID string) ([]domain.Professional, error)
	GetByID(id string) (*domain.Professional, error)
	// CheckAddressDistance valida un domicilio guardado del cliente contra el
	// radio de alcance de un profesional puntual — se usa cuando el cliente
	// cambia de domicilio ya dentro del flujo de pedir presupuesto, después
	// de haber entrado por el listado filtrado por su domicilio activo.
	CheckAddressDistance(clerkID, professionalID, addressID string) (distanceKm float64, withinRadius bool, err error)
}

type professionalUseCase struct {
	repo      domain.ProfessionalRepository
	users     domain.UserRepository
	addresses domain.AddressRepository
	storage   *storage.SupabaseClient
}

func NewProfessionalUseCase(repo domain.ProfessionalRepository, users domain.UserRepository, addresses domain.AddressRepository, storageClient *storage.SupabaseClient) ProfessionalUseCase {
	return &professionalUseCase{repo: repo, users: users, addresses: addresses, storage: storageClient}
}

// GetAll devuelve los profesionales cuyo radio de alcance cubre el domicilio
// de quien pregunta — reemplaza al viejo listado por zona. Requiere que el
// usuario ya haya cargado su domicilio.
func (uc *professionalUseCase) GetAll(clerkID string) ([]domain.Professional, error) {
	user, err := uc.users.FindByClerkID(clerkID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, ErrUserNotOnboarded
	}
	if !user.HasHomeAddress() {
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

func (uc *professionalUseCase) CheckAddressDistance(clerkID, professionalID, addressID string) (float64, bool, error) {
	user, err := uc.users.FindByClerkID(clerkID)
	if err != nil {
		return 0, false, err
	}
	if user == nil {
		return 0, false, ErrUserNotOnboarded
	}

	addr, err := uc.addresses.FindByID(addressID)
	if err != nil {
		return 0, false, err
	}
	if addr == nil || addr.UserID != user.ID {
		return 0, false, ErrAddressNotFound
	}

	distanceKm, radiusKm, err := uc.repo.DistanceToPoint(professionalID, addr.Lat, addr.Lng)
	if err != nil {
		return 0, false, err
	}
	return distanceKm, distanceKm <= float64(radiusKm), nil
}
