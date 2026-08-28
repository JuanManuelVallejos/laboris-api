package domain

type Address struct {
	ID     string `json:"id"`
	UserID string `json:"-"`
	Label  string `json:"label"`
	// Address es el texto formateado que devuelve Google Places — el usuario
	// nunca escribe una dirección libre, así que este texto es siempre una
	// dirección real y específica.
	Address string `json:"address"`
	// Lat/Lng nunca se serializan, mismo criterio de privacidad que
	// Professional.HomeLat/HomeLng y User.HomeLat/HomeLng.
	Lat       float64 `json:"-"`
	Lng       float64 `json:"-"`
	IsDefault bool    `json:"isDefault"`
	// HasActiveJob se completa al leer (FindByUserID) — true si hay un job en
	// estado no final creado con este domicilio. Bloquea edición/borrado.
	HasActiveJob bool `json:"hasActiveJob"`
}

type AddressRepository interface {
	FindByUserID(userID string) ([]Address, error)
	FindByID(id string) (*Address, error)
	Create(a *Address) (*Address, error)
	// CreateIfNotExists inserta el domicilio salvo que ya exista uno igual
	// (mismo user_id + address) — en ese caso devuelve el existente en vez
	// de duplicarlo. Pensado para la migración automática de users.home_address
	// en AddressUseCase.List, que puede dispararse en paralelo desde varias
	// pantallas.
	CreateIfNotExists(a *Address) (*Address, error)
	Update(id, label, address string, lat, lng float64) (*Address, error)
	Delete(id string) error
	SetDefault(userID, addressID string) error
	// ClearDefault saca is_default a todos los domicilios del usuario —
	// usado cuando se borra el único domicilio que quedaba.
	ClearDefault(userID string) error
	HasActiveJob(addressID string) (bool, error)
	CountByUserID(userID string) (int, error)
}
