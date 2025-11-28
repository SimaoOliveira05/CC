package devices

import (
	"sync"
	"time"
)

type MockBattery struct {
	level     uint8
	charging  bool
	lastCheck time.Time
	mu        sync.Mutex
}

func NewMockBattery(initialLevel uint8) *MockBattery {
	return &MockBattery{
		level:     initialLevel,
		charging:  false,
		lastCheck: time.Now(),
	}
}

func (b *MockBattery) GetLevel() uint8 {
	// Simula descarga gradual baseada em segundos (para testes mais rápidos)
	b.mu.Lock()
	defer b.mu.Unlock()
	elapsedSec := time.Since(b.lastCheck).Seconds()
	if !b.charging && elapsedSec > 0.5 {
		// Drain rate: 0.5% por segundo (aprox 30% por minuto)
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

func (b *MockBattery) IsCharging() bool {
	return b.charging
}

func (b *MockBattery) SetCharging(charging bool) {
	b.charging = charging
}

func (b *MockBattery) SetLevel(level uint8) {
	b.mu.Lock()
	defer b.mu.Unlock()
	if level > 100 {
		level = 100
	}
	b.level = level
}
