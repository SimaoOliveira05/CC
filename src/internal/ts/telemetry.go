package ts

import (
    "math/rand"
    "time"
    "src/utils"
)

// GenerateTelemetry gera dados de telemetria simulados
func GenerateTelemetry(roverID uint8, state uint8, position utils.Coordinate, battery uint8, speed float32) *TelemetryPacket {
    return &TelemetryPacket{
        RoverID:     roverID,
        Timestamp:   time.Now().Unix(),
        Position:    position,
        State:       state,
        Battery:     battery,
        Speed:       speed,
        Temperature: int16(20 + rand.Intn(30)), // 20-50°C
        WheelStatus: 0b1111,                    // Todas as rodas OK
    }
}