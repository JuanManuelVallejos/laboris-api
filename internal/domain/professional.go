package domain

type Professional struct {
	ID       string  `json:"id"`
	Name     string  `json:"name"`
	Trade    string  `json:"trade"`
	Zone     string  `json:"zone"`
	Rating   float64 `json:"rating"`
	Verified bool    `json:"verified"`
}

type ProfessionalRepository interface {
	FindAll() ([]Professional, error)
	FindByID(id string) (*Professional, error)
}
