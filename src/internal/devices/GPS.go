package devices

import (
    "math/rand"
    "src/utils"
	"sync"
)

type MockGPS struct {
    position utils.Coordinate
    speed    float32
	mu       sync.Mutex
}

func NewMockGPS(initialPos utils.Coordinate) *MockGPS {
    return &MockGPS{
        position: initialPos,
        speed:    0.0,
    }
}

func (g *MockGPS) GetPosition() utils.Coordinate {
    // Simula movimento aleatório
    g.position.Latitude += (rand.Float64() - 0.5) * 0.0001
    g.position.Longitude += (rand.Float64() - 0.5) * 0.0001
    return g.position
}

func (g *MockGPS) GetSpeed() float32 {
    // Simula velocidade variável
    g.speed = rand.Float32() * 5.0 // 0-5 m/s
    return g.speed
}

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