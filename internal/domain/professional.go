package domain

type Professional struct {
	ID          string  `json:"id"`
	UserID      string  `json:"userId"`
	Name        string  `json:"name"`
	AvatarURL   string  `json:"avatarUrl"`
	Trade       string  `json:"trade"`
	HomeAddress string  `json:"homeAddress"`
	// HomeLat/HomeLng nunca se serializan — la posición exacta del
	// profesional no se expone por API, solo la distancia calculada
	// (DistanceKm) cuando corresponde a un listado geo-filtrado.
	HomeLat  *float64 `json:"-"`
	HomeLng  *float64 `json:"-"`
	RadiusKm *int     `json:"radiusKm"`
	// DistanceKm se completa al leer en un listado filtrado por cercanía
	// (FindNear) — no es una columna.
	DistanceKm      *float64     `json:"distanceKm,omitempty"`
	Bio             string       `json:"bio"`
	Rating          float64      `json:"rating"`
	Verified        bool         `json:"verified"`
	Status          string       `json:"status"`
	PortfolioPhotos []Attachment `json:"portfolioPhotos"`
}

type ProfessionalRepository interface {
	// FindNear devuelve los profesionales activos, con domicilio y radio
	// cargados, cuyo radio de alcance cubre (clientLat, clientLng) —
	// reemplaza al viejo listado por zona. DistanceKm queda completo en
	// cada resultado.
	FindNear(clientLat, clientLng float64) ([]Professional, error)
	FindByID(id string) (*Professional, error)
	FindByUserID(userID string) (*Professional, error)
	Create(p *Professional) (*Professional, error)
	UpdateByUserID(userID, trade, homeAddress, bio string, homeLat, homeLng float64, radiusKm int) (*Professional, error)
	FindAllPaginated(page, limit int) ([]Professional, int64, error)
	SetVerified(id string, verified bool) error
	SetStatus(id string, status string) error
	Delete(id string) error
}
