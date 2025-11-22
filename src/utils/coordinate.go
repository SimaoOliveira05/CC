package utils

import (
	"fmt"
)

// Coordinate representa um ponto geográfico em graus decimais.
type Coordinate struct {
	Latitude  float64 `json:"latitude"`  // ex: 41.545
	Longitude float64 `json:"longitude"` // ex: -8.421
}

// String devolve uma representação legível da coordenada.
func (c Coordinate) String() string {
	return fmt.Sprintf("(%.6f, %.6f)", c.Latitude, c.Longitude)
}
