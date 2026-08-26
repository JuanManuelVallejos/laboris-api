package domain

import "time"

type User struct {
	ID          string     `json:"id"`
	ClerkID     string     `json:"clerkId"`
	Email       string     `json:"email"`
	FullName    string     `json:"fullName"`
	AvatarURL   string     `json:"avatarUrl"`
	HomeAddress string     `json:"homeAddress"`
	HomeLat     *float64   `json:"-"`
	HomeLng     *float64   `json:"-"`
	CreatedAt   time.Time  `json:"createdAt"`
	DeletedAt   *time.Time `json:"deletedAt"`
}

// HasHomeAddress indica si el usuario ya cargó su domicilio — requisito
// para ver el listado de profesionales y para pedir presupuesto.
func (u *User) HasHomeAddress() bool {
	return u.HomeLat != nil && u.HomeLng != nil
}

type UserWithRoles struct {
	User
	Roles []string `json:"roles"`
}

type UserRepository interface {
	FindByClerkID(clerkID string) (*User, error)
	Create(user *User) (*User, error)
	AddRole(userID string, role string) error
	FindAllPaginated(page, limit int) ([]UserWithRoles, int64, error)
	SoftDeleteByClerkID(clerkID string) error
	UpdateAvatarURLByClerkID(clerkID, avatarURL string) error
	UpdateHomeAddress(userID, address string, lat, lng float64) error
	// ClearHomeAddress se usa cuando un cliente borra su último domicilio
	// guardado — vuelve a quedar sin domicilio (gate de ErrAddressRequired).
	ClearHomeAddress(userID string) error
}
