package utils

import (
	"fmt"
)

// Coordinate represents a geographical coordinate with latitude and longitude.
type Coordinate struct {
	Latitude  float64 `json:"latitude"`  // ex: 41.545
	Longitude float64 `json:"longitude"` // ex: -8.421
}

// String returns a human-readable representation of the coordinate.
func (c Coordinate) String() string {
	return fmt.Sprintf("(%.6f, %.6f)", c.Latitude, c.Longitude)
}
