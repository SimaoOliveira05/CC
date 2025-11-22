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
    // Simula descarga gradual
    elapsed := time.Since(b.lastCheck).Minutes()
    if !b.charging && elapsed > 1 {
        if b.level > 0 {
            b.level -= uint8(elapsed * 0.5) // 0.5% por minuto
            if b.level > 100 {
                b.level = 0 // evitar underflow
            }
        }
        b.lastCheck = time.Now()
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