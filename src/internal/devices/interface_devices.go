package devices

import "src/utils"

type GPS interface {
	GetPosition() utils.Coordinate
	GetSpeed() float32
	GetAltitude() float32
}

type Thermometer interface {
	// Methods to read full environmental data used in EnvReportData
	GetTemperature() float32
	GetOxygen() float32
	GetPressure() float32
	GetHumidity() float32
	GetWindSpeed() float32
	GetRadiation() float32
}

type Battery interface {
	GetLevel() uint8
	IsCharging() bool
}
