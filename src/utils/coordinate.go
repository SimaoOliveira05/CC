package utils

import (
	"fmt"
)

// Raio médio da Terra em metros
const EarthRadius = 6371e3 // 6371 km

// Coordinate representa um ponto geográfico em graus decimais.
type Coordinate struct {
	Latitude  float64 // ex: 41.545
	Longitude float64 // ex: -8.421
}

// String devolve uma representação legível da coordenada.
func (c Coordinate) String() string {
	return fmt.Sprintf("(%.6f, %.6f)", c.Latitude, c.Longitude)
}
