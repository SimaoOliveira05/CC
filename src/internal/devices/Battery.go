package devices

import (
	"src/config"
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
		// Drain rate from config
		drain := uint8(elapsedSec * config.BATTERY_DRAIN_RATE)
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

// StartCharging initiates battery charging
func (b *MockBattery) StartCharging() {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.charging = true
	b.lastCheck = time.Now()
}

// StopCharging stops battery charging
func (b *MockBattery) StopCharging() {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.charging = false
	b.lastCheck = time.Now()
}

// Recharge simulates battery recharging over time
// Returns true when fully charged
func (b *MockBattery) Recharge() bool {
	b.mu.Lock()
	defer b.mu.Unlock()

	if !b.charging {
		b.charging = true
		b.lastCheck = time.Now()
	}

	elapsedSec := time.Since(b.lastCheck).Seconds()
	if elapsedSec > 0.5 {
		// Charge rate from config
		charge := uint8(elapsedSec * config.BATTERY_CHARGE_RATE)
		if charge > 0 {
			if b.level+charge < 100 {
				b.level += charge
			} else {
				b.level = 100
			}
			b.lastCheck = time.Now()
		}
	}

	return b.level >= 100
}

// IsCritical returns true if battery is at critical level
func (b *MockBattery) IsCritical() bool {
	return b.GetLevel() < config.CRITICAL_BATTERY_LEVEL
}
