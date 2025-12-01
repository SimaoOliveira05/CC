package devices

import (
	"sync"
	"time"
)

// Battery interface
type Battery interface {
	GetLevel() uint8
	IsCharging() bool
}

// MockBattery simulate a device battery for testing purposes
type MockBattery struct {
	level     uint8
	charging  bool
	lastCheck time.Time
	mu        sync.Mutex
}

// NewMockBattery creates a new MockBattery with the specified initial level
func NewMockBattery(initialLevel uint8) *MockBattery {
	return &MockBattery{
		level:     initialLevel,
		charging:  false,
		lastCheck: time.Now(),
	}
}

// GetLevel returns the current battery level (0-100)
func (b *MockBattery) GetLevel() uint8 {
	// Simulate battery drain over time
	b.mu.Lock()
	defer b.mu.Unlock()
	elapsedSec := time.Since(b.lastCheck).Seconds()
	if !b.charging && elapsedSec > 0.5 {
		// Drain rate: 0.5% per second (approx 30% per minute)
		drain := uint8(elapsedSec * 0.5)
		if drain > 0 {
			if b.level > drain {
				b.level -= drain
			} else {
				b.level = 0
			}
			b.lastCheck = time.Now()
		}
	}
	return b.level
}

// IsCharging returns whether the battery is currently charging
func (b *MockBattery) IsCharging() bool {
	return b.charging
}

// SetLevel sets the battery level (0-100)
func (b *MockBattery) SetLevel(level uint8) {
	b.mu.Lock()
	defer b.mu.Unlock()
	if level > 100 {
		level = 100
	}
	b.level = level
}
