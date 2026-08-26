package postgres

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/laboris/laboris-api/internal/domain"
)

type ProfessionalRepository struct {
	db *pgxpool.Pool
}

func NewProfessionalRepository(db *pgxpool.Pool) *ProfessionalRepository {
	return &ProfessionalRepository{db: db}
}

// distanceKmExpr calcula la distancia (Haversine, en km) entre el punto
// ($1, $2) = (lat, lng) del cliente y el domicilio del profesional. El
// LEAST/GREATEST evita que un redondeo de punto flotante empuje el argumento
// de acos() apenas fuera de [-1, 1], lo que rompería la query con NaN.
const distanceKmExpr = `6371 * acos(LEAST(1, GREATEST(-1,
	cos(radians($1)) * cos(radians(p.home_lat)) * cos(radians(p.home_lng) - radians($2)) +
	sin(radians($1)) * sin(radians(p.home_lat))
)))`

// FindNear devuelve los profesionales activos cuyo radio de alcance cubre al
// cliente ubicado en (clientLat, clientLng) — reemplaza al viejo filtro por
// zona. Los que no cargaron domicilio/radio quedan afuera (home_lat/home_lng/
// radius_km NULL nunca matchean).
func (r *ProfessionalRepository) FindNear(clientLat, clientLng float64) ([]domain.Professional, error) {
	rows, err := r.db.Query(context.Background(), `
		SELECT p.id, p.user_id, u.full_name, COALESCE(u.avatar_url, '') AS avatar_url, p.trade,
		       COALESCE(p.home_address, ''), p.radius_km, p.bio, p.verified, p.status,
		       COALESCE(AVG(rv.rating), 0) AS rating,
		       `+distanceKmExpr+` AS distance_km
		FROM professionals p
		JOIN users u ON u.id = p.user_id
		LEFT JOIN reviews rv ON rv.professional_id = p.id
		WHERE p.status = 'active'
		  AND p.home_lat IS NOT NULL AND p.home_lng IS NOT NULL AND p.radius_km IS NOT NULL
		GROUP BY p.id, u.full_name, u.avatar_url
		HAVING `+distanceKmExpr+` <= p.radius_km
		ORDER BY distance_km ASC
	`, clientLat, clientLng)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make([]domain.Professional, 0)
	for rows.Next() {
		var p domain.Professional
		if err := rows.Scan(
			&p.ID, &p.UserID, &p.Name, &p.AvatarURL, &p.Trade,
			&p.HomeAddress, &p.RadiusKm, &p.Bio, &p.Verified, &p.Status,
			&p.Rating, &p.DistanceKm,
		); err != nil {
			return nil, err
		}
		result = append(result, p)
	}
	return result, nil
}

func (r *ProfessionalRepository) FindByID(id string) (*domain.Professional, error) {
	p := &domain.Professional{}
	err := r.db.QueryRow(context.Background(), `
		SELECT p.id, p.user_id, u.full_name, COALESCE(u.avatar_url, '') AS avatar_url, p.trade,
		       COALESCE(p.home_address, ''), p.radius_km, p.bio, p.verified, p.status,
		       COALESCE(AVG(rv.rating), 0) AS rating
		FROM professionals p
		JOIN users u ON u.id = p.user_id
		LEFT JOIN reviews rv ON rv.professional_id = p.id
		WHERE p.id = $1
		GROUP BY p.id, u.full_name, u.avatar_url
	`, id).Scan(&p.ID, &p.UserID, &p.Name, &p.AvatarURL, &p.Trade, &p.HomeAddress, &p.RadiusKm, &p.Bio, &p.Verified, &p.Status, &p.Rating)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	photos, err := r.fetchPortfolioPhotos(p.ID)
	if err != nil {
		return nil, err
	}
	p.PortfolioPhotos = photos
	return p, nil
}

func (r *ProfessionalRepository) FindByUserID(userID string) (*domain.Professional, error) {
	p := &domain.Professional{}
	err := r.db.QueryRow(context.Background(), `
		SELECT p.id, p.user_id, u.full_name, COALESCE(u.avatar_url, '') AS avatar_url, p.trade,
		       COALESCE(p.home_address, ''), p.radius_km, p.bio, p.verified, p.status,
		       COALESCE(AVG(rv.rating), 0) AS rating
		FROM professionals p
		JOIN users u ON u.id = p.user_id
		LEFT JOIN reviews rv ON rv.professional_id = p.id
		WHERE p.user_id = $1
		GROUP BY p.id, u.full_name, u.avatar_url
	`, userID).Scan(&p.ID, &p.UserID, &p.Name, &p.AvatarURL, &p.Trade, &p.HomeAddress, &p.RadiusKm, &p.Bio, &p.Verified, &p.Status, &p.Rating)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	photos, err := r.fetchPortfolioPhotos(p.ID)
	if err != nil {
		return nil, err
	}
	p.PortfolioPhotos = photos
	return p, nil
}

// fetchPortfolioPhotos loads a professional's portfolio attachments directly
// (same pattern as JobRepository.fetchReworkRecords) rather than depending
// on AttachmentRepository, keeping repositories independent of each other.
func (r *ProfessionalRepository) fetchPortfolioPhotos(professionalID string) ([]domain.Attachment, error) {
	rows, err := r.db.Query(context.Background(), `
		SELECT id, type, owner_id, path, filename, extension, uploaded_by, created_at
		FROM attachments
		WHERE type = $1 AND owner_id = $2
		ORDER BY created_at ASC
	`, domain.AttachmentTypeProfessionalPortfolio, professionalID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	photos := make([]domain.Attachment, 0)
	for rows.Next() {
		var a domain.Attachment
		if err := rows.Scan(&a.ID, &a.Type, &a.OwnerID, &a.Path, &a.Filename, &a.Extension, &a.UploadedBy, &a.CreatedAt); err != nil {
			return nil, err
		}
		photos = append(photos, a)
	}
	return photos, nil
}

func (r *ProfessionalRepository) UpdateByUserID(userID, trade, homeAddress, bio string, homeLat, homeLng float64, radiusKm int) (*domain.Professional, error) {
	p := &domain.Professional{}
	err := r.db.QueryRow(context.Background(), `
		UPDATE professionals SET trade = $2, home_address = $3, home_lat = $4, home_lng = $5, radius_km = $6, bio = $7
		WHERE user_id = $1
		RETURNING id, user_id, trade, home_address, radius_km, bio, verified, status
	`, userID, trade, homeAddress, homeLat, homeLng, radiusKm, bio).
		Scan(&p.ID, &p.UserID, &p.Trade, &p.HomeAddress, &p.RadiusKm, &p.Bio, &p.Verified, &p.Status)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	return p, err
}

func (r *ProfessionalRepository) Create(p *domain.Professional) (*domain.Professional, error) {
	err := r.db.QueryRow(context.Background(),
		`INSERT INTO professionals (user_id, trade, home_address, home_lat, home_lng, radius_km, bio)
		 VALUES ($1, $2, $3, $4, $5, $6, $7)
		 ON CONFLICT (user_id) DO UPDATE SET
		   trade = EXCLUDED.trade, home_address = EXCLUDED.home_address,
		   home_lat = EXCLUDED.home_lat, home_lng = EXCLUDED.home_lng,
		   radius_km = EXCLUDED.radius_km, bio = EXCLUDED.bio
		 RETURNING id, user_id, trade, home_address, radius_km, bio, verified, status`,
		p.UserID, p.Trade, p.HomeAddress, p.HomeLat, p.HomeLng, p.RadiusKm, p.Bio,
	).Scan(&p.ID, &p.UserID, &p.Trade, &p.HomeAddress, &p.RadiusKm, &p.Bio, &p.Verified, &p.Status)
	return p, err
}

func (r *ProfessionalRepository) FindAllPaginated(page, limit int) ([]domain.Professional, int64, error) {
	offset := (page - 1) * limit

	var total int64
	if err := r.db.QueryRow(context.Background(), `SELECT COUNT(*) FROM professionals`).Scan(&total); err != nil {
		return nil, 0, err
	}

	rows, err := r.db.Query(context.Background(), `
		SELECT p.id, p.user_id, u.full_name, COALESCE(u.avatar_url, '') AS avatar_url, p.trade,
		       COALESCE(p.home_address, ''), p.radius_km, p.bio, p.verified, p.status,
		       COALESCE(AVG(rv.rating), 0) AS rating
		FROM professionals p
		JOIN users u ON u.id = p.user_id
		LEFT JOIN reviews rv ON rv.professional_id = p.id
		GROUP BY p.id, u.full_name, u.avatar_url
		ORDER BY p.status ASC, u.full_name ASC
		LIMIT $1 OFFSET $2
	`, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	result := make([]domain.Professional, 0)
	for rows.Next() {
		var p domain.Professional
		if err := rows.Scan(&p.ID, &p.UserID, &p.Name, &p.AvatarURL, &p.Trade, &p.HomeAddress, &p.RadiusKm, &p.Bio, &p.Verified, &p.Status, &p.Rating); err != nil {
			return nil, 0, err
		}
		result = append(result, p)
	}
	return result, total, nil
}

func (r *ProfessionalRepository) SetVerified(id string, verified bool) error {
	_, err := r.db.Exec(context.Background(),
		`UPDATE professionals SET verified = $2 WHERE id = $1`, id, verified)
	return err
}

func (r *ProfessionalRepository) SetStatus(id string, status string) error {
	_, err := r.db.Exec(context.Background(),
		`UPDATE professionals SET status = $2 WHERE id = $1`, id, status)
	return err
}

func (r *ProfessionalRepository) Delete(id string) error {
	_, err := r.db.Exec(context.Background(),
		`DELETE FROM professionals WHERE id = $1`, id)
	return err
}
