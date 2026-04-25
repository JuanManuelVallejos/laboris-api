package domain

type Professional struct {
	ID       string  `json:"id"`
	UserID   string  `json:"userId"`
	Name     string  `json:"name"`
	Trade    string  `json:"trade"`
	Zone     string  `json:"zone"`
	Bio      string  `json:"bio"`
	Rating   float64 `json:"rating"`
	Verified bool    `json:"verified"`
	Status   string  `json:"status"`
}

type ProfessionalRepository interface {
	FindAll() ([]Professional, error)
	FindByID(id string) (*Professional, error)
	FindByUserID(userID string) (*Professional, error)
	Create(p *Professional) (*Professional, error)
	UpdateByUserID(userID, trade, zone, bio string) (*Professional, error)
	FindAllPaginated(page, limit int) ([]Professional, int64, error)
	SetVerified(id string, verified bool) error
	SetStatus(id string, status string) error
	Delete(id string) error
}
