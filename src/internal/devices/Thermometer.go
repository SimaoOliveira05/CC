package devices

import (
	"math/rand"
)

type MockThermometer struct {
	baseTemp float32
}

func NewMockThermometer() *MockThermometer {
	return &MockThermometer{
		baseTemp: -60.0, // temperatura base típica em simulação marciana
	}
}

func (t *MockThermometer) GetTemperature() float32 {
	// Simula variação de temperatura (±10°C)
	variation := (rand.Float32() - 0.5) * 20.0
	return t.baseTemp + variation
}

func (t *MockThermometer) GetOxygen() float32 {
	// Percentagem muito baixa (Mars-like), 0.0 - 1.0%
	return rand.Float32() * 0.5
}

func (t *MockThermometer) GetPressure() float32 {
	// Simula pressão em Pascals, exemplo 500-800 Pa
	return 500.0 + rand.Float32()*300.0
}

func (t *MockThermometer) GetHumidity() float32 {
	// Humidade relativa baixa 0-20%
	return rand.Float32() * 20.0
}

func (t *MockThermometer) GetWindSpeed() float32 {
	// Velocidade do vento 0-30 m/s
	return rand.Float32() * 30.0
}

func (t *MockThermometer) GetRadiation() float32 {
	// Níveis de radiação 0.0 - 0.5 (arbitrário)
	return rand.Float32() * 0.5
}
