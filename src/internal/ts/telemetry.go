package ts

import (
	"math/rand"
	"src/utils"
	"time"
)

// GenerateTelemetry generates a telemetry packet for a rover.
func GenerateTelemetry(roverID uint8, state uint8, position utils.Coordinate, battery uint8, speed float32, queueP1 uint8, queueP2 uint8, queueP3 uint8) *TelemetryPacket {
	return &TelemetryPacket{
		RoverID:      roverID,
		Timestamp:    time.Now().Unix(),
		Position:     position,
		State:        state,
		Battery:      battery,
		Speed:        speed,
		Temperature:  int16(20 + rand.Intn(30)), // 20-50°C
		WheelStatus:  0b1111,                    // All wheels OK
		QueueP1Count: queueP1,
		QueueP2Count: queueP2,
		QueueP3Count: queueP3,
	}
}
