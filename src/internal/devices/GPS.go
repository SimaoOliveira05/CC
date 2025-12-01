package devices

import (
    "math/rand"
    "src/utils"
	"sync"
)

// GPS interface
type GPS interface {
	GetPosition() utils.Coordinate
	GetSpeed() float32
	GetAltitude() float32
}

// MockGPS simulates a GPS device for testing purposes
type MockGPS struct {
    position utils.Coordinate
    speed    float32
	mu       sync.Mutex
}

// NewMockGPS creates a new MockGPS with the specified initial position
func NewMockGPS(initialPos utils.Coordinate) *MockGPS {
    return &MockGPS{
        position: initialPos,
        speed:    0.0,
    }
}

// GetPosition returns the current GPS position
func (g *MockGPS) GetPosition() utils.Coordinate {
    // Simulate random movement
    g.position.Latitude += (rand.Float64() - 0.5) * 0.0001
    g.position.Longitude += (rand.Float64() - 0.5) * 0.0001
    return g.position
}

// GetSpeed returns the current GPS speed
func (g *MockGPS) GetSpeed() float32 {
    // Simulate variable speed
    g.speed = rand.Float32() * 5.0 // 0-5 m/s
    return g.speed
}

// GetAltitude returns the current GPS altitude
func (g *MockGPS) GetAltitude() float32 {
    return 100.0 + rand.Float32()*50.0 // 100-150m
}

func (g *MockGPS) SetPosition(pos utils.Coordinate) {
    g.mu.Lock()
    defer g.mu.Unlock()
    g.position = pos
}

func (g *MockGPS) SetSpeed(speed float32) {
    g.mu.Lock()
    defer g.mu.Unlock()
    g.speed = speed
}