package devices

import (
	"math/rand"
)

// Thermometer interface
type Thermometer interface {
	// Methods to read full environmental data used in EnvReportData
	GetTemperature() float32
	GetOxygen() float32
	GetPressure() float32
	GetHumidity() float32
	GetWindSpeed() float32
	GetRadiation() float32
}

// MockThermometer simulates a thermometer device for testing purposes reading all environmental data
type MockThermometer struct {
	baseTemp float32
}

// NewMockThermometer creates a new MockThermometer
func NewMockThermometer() *MockThermometer {
	return &MockThermometer{
		baseTemp: -60.0, // base typical temperature on Mars in °C
	}
}

// GetTemperature returns the current temperature reading
func (t *MockThermometer) GetTemperature() float32 {
	// Simulates temperature variation (±10°C)
	variation := (rand.Float32() - 0.5) * 20.0
	return t.baseTemp + variation
}

// GetOxygen returns the current oxygen level reading
func (t *MockThermometer) GetOxygen() float32 {
	// Very low percentage (Mars-like), 0.0 - 1.0%
	return rand.Float32() * 0.5
}

// GetPressure returns the current pressure reading in Pascals
func (t *MockThermometer) GetPressure() float32 {
	// Simulates pressure in Pascals, example 500-800 Pa
	return 500.0 + rand.Float32()*300.0
}

// GetHumidity returns the current humidity level reading
func (t *MockThermometer) GetHumidity() float32 {
	// Low relative humidity 0-20%
	return rand.Float32() * 20.0
}

// GetWindSpeed returns the current wind speed reading
func (t *MockThermometer) GetWindSpeed() float32 {
	// Wind speed 0-30 m/s
	return rand.Float32() * 30.0
}

// GetRadiation returns the current radiation level reading
func (t *MockThermometer) GetRadiation() float32 {
	// Radiation levels 0.0 - 0.5 (arbitrary)
	return rand.Float32() * 0.5
}
