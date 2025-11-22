package devices

import (
	"math/rand"
)

type MockThermometer struct {
	baseTemp float32
}

func NewMockThermometer() *MockThermometer {
	return &MockThermometer{
		baseTemp: 20.0, // temperatura base
	}
}

func (t *MockThermometer) GetTemperature() float32 {
	// Simula variação de temperatura
	variation := (rand.Float32() - 0.5) * 5.0 // ±2.5°C
	return t.baseTemp + variation
}
