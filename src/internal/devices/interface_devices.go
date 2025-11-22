package devices

import "src/utils"

type GPS interface {
	GetPosition() utils.Coordinate
	GetSpeed() float32
	GetAltitude() float32
}

type Thermometer interface {
	GetTemperature() float32
}

type Battery interface {
	GetLevel() uint8
	IsCharging() bool
}