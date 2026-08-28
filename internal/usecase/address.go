package usecase

import (
	"context"
	"errors"

	"github.com/laboris/laboris-api/internal/domain"
	"github.com/laboris/laboris-api/internal/geocoding"
)

// ErrAddressInUse se devuelve al intentar editar/borrar un domicilio que
// tiene un trabajo en curso creado con él.
var ErrAddressInUse = errors.New("no podés editar ni borrar un domicilio con un trabajo en curso")

// ErrAddressNotFound cubre tanto "no existe" como "no es tuyo" — mismo
// domicilio de respuesta para no filtrar cuál es el caso real.
var ErrAddressNotFound = errors.New("domicilio no encontrado")

type AddressUseCase struct {
	addresses domain.AddressRepository
	users     domain.UserRepository
	geo       *geocoding.Client
}

func NewAddressUseCase(addresses domain.AddressRepository, users domain.UserRepository, geo *geocoding.Client) *AddressUseCase {
	return &AddressUseCase{addresses: addresses, users: users, geo: geo}
}

// List devuelve los domicilios guardados del cliente. Si todavía no tiene
// ninguno pero sí tiene users.home_* cargado (onboarding de antes de este
// sistema), se migra automáticamente a un primer domicilio "Casa" — así
// ningún usuario existente pierde su domicilio ni hace falta tocar el
// onboarding ya probado.
func (uc *AddressUseCase) List(clerkID string) ([]domain.Address, error) {
	user, err := uc.users.FindByClerkID(clerkID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, ErrUserNotOnboarded
	}
	addrs, err := uc.addresses.FindByUserID(user.ID)
	if err != nil {
		return nil, err
	}
	if len(addrs) == 0 && user.HasHomeAddress() {
		created, err := uc.addresses.CreateIfNotExists(&domain.Address{
			UserID: user.ID, Label: "Casa", Address: user.HomeAddress,
			Lat: *user.HomeLat, Lng: *user.HomeLng, IsDefault: true,
		})
		if err != nil {
			return nil, err
		}
		addrs = []domain.Address{*created}
	}
	return addrs, nil
}

// Create agrega un domicilio nuevo. El primero que carga un usuario queda
// como default automáticamente y sincroniza users.home_* (la caché que usa
// el listado de profesionales por cercanía).
func (uc *AddressUseCase) Create(clerkID, label, address string) (*domain.Address, error) {
	user, err := uc.users.FindByClerkID(clerkID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, ErrUserNotOnboarded
	}
	lat, lng, err := uc.geo.Geocode(context.Background(), address)
	if err != nil {
		return nil, err
	}
	count, err := uc.addresses.CountByUserID(user.ID)
	if err != nil {
		return nil, err
	}
	isDefault := count == 0
	created, err := uc.addresses.Create(&domain.Address{
		UserID: user.ID, Label: label, Address: address, Lat: lat, Lng: lng, IsDefault: isDefault,
	})
	if err != nil {
		return nil, err
	}
	if isDefault {
		if err := uc.users.UpdateHomeAddress(user.ID, address, lat, lng); err != nil {
			return nil, err
		}
	}
	return created, nil
}

func (uc *AddressUseCase) Update(clerkID, addressID, label, address string) (*domain.Address, error) {
	user, existing, err := uc.ownedAddress(clerkID, addressID)
	if err != nil {
		return nil, err
	}
	if existing == nil {
		return nil, ErrAddressNotFound
	}
	active, err := uc.addresses.HasActiveJob(addressID)
	if err != nil {
		return nil, err
	}
	if active {
		return nil, ErrAddressInUse
	}
	lat, lng, err := uc.geo.Geocode(context.Background(), address)
	if err != nil {
		return nil, err
	}
	updated, err := uc.addresses.Update(addressID, label, address, lat, lng)
	if err != nil {
		return nil, err
	}
	if existing.IsDefault {
		if err := uc.users.UpdateHomeAddress(user.ID, address, lat, lng); err != nil {
			return nil, err
		}
	}
	return updated, nil
}

func (uc *AddressUseCase) Delete(clerkID, addressID string) error {
	user, existing, err := uc.ownedAddress(clerkID, addressID)
	if err != nil {
		return err
	}
	if existing == nil {
		return ErrAddressNotFound
	}
	active, err := uc.addresses.HasActiveJob(addressID)
	if err != nil {
		return err
	}
	if active {
		return ErrAddressInUse
	}
	if err := uc.addresses.Delete(addressID); err != nil {
		return err
	}
	if !existing.IsDefault {
		return nil
	}

	remaining, err := uc.addresses.FindByUserID(user.ID)
	if err != nil {
		return err
	}
	if len(remaining) == 0 {
		return uc.users.ClearHomeAddress(user.ID)
	}
	// Promueve el domicilio guardado más reciente como nuevo default —
	// FindByUserID ordena is_default DESC, created_at ASC, así que el
	// último de la lista es el creado más recientemente.
	next := remaining[len(remaining)-1]
	if err := uc.addresses.SetDefault(user.ID, next.ID); err != nil {
		return err
	}
	return uc.users.UpdateHomeAddress(user.ID, next.Address, next.Lat, next.Lng)
}

func (uc *AddressUseCase) SetDefault(clerkID, addressID string) error {
	user, existing, err := uc.ownedAddress(clerkID, addressID)
	if err != nil {
		return err
	}
	if existing == nil {
		return ErrAddressNotFound
	}
	if err := uc.addresses.SetDefault(user.ID, addressID); err != nil {
		return err
	}
	return uc.users.UpdateHomeAddress(user.ID, existing.Address, existing.Lat, existing.Lng)
}

// ownedAddress valida que addressID pertenezca al clerkID dado — mismo
// criterio de pertenencia que ErrSelfRequest en RequestUseCase.
func (uc *AddressUseCase) ownedAddress(clerkID, addressID string) (*domain.User, *domain.Address, error) {
	user, err := uc.users.FindByClerkID(clerkID)
	if err != nil {
		return nil, nil, err
	}
	if user == nil {
		return nil, nil, ErrUserNotOnboarded
	}
	addr, err := uc.addresses.FindByID(addressID)
	if err != nil {
		return nil, nil, err
	}
	if addr == nil || addr.UserID != user.ID {
		return user, nil, nil
	}
	return user, addr, nil
}
