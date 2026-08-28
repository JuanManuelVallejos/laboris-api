package domain

import (
	"hash/fnv"
	"math"
	"math/rand"
	"strings"
)

// CoarsenAddress recorta el primer segmento separado por coma (calle +
// altura) de una dirección formateada por Google, dejando el resto tal cual
// (localidad/provincia/país) — así el profesional puede ubicar la zona sin
// ver el domicilio exacto. Si no hay coma (formato inesperado), devuelve una
// cadena vacía antes que arriesgarse a mostrar la dirección completa.
func CoarsenAddress(full string) string {
	idx := strings.Index(full, ",")
	if idx == -1 {
		return ""
	}
	return strings.TrimSpace(full[idx+1:])
}

// JitterCoords devuelve un punto desplazado de forma pseudoaleatoria pero
// determinística (misma seed → mismo resultado siempre, para que el círculo
// aproximado no salte en cada visita) hasta maxMeters de (lat, lng) real —
// el punto real siempre cae dentro de un círculo dibujado sobre el resultado
// con un radio mayor a maxMeters.
func JitterCoords(lat, lng float64, seed string, maxMeters float64) (float64, float64) {
	h := fnv.New64a()
	_, _ = h.Write([]byte(seed))
	rng := rand.New(rand.NewSource(int64(h.Sum64())))

	angle := rng.Float64() * 2 * math.Pi
	distance := rng.Float64() * maxMeters

	const metersPerDegreeLat = 111320.0
	dLat := (distance * math.Cos(angle)) / metersPerDegreeLat
	dLng := (distance * math.Sin(angle)) / (metersPerDegreeLat * math.Cos(lat*math.Pi/180))

	return lat + dLat, lng + dLng
}
