package ts

import (
    "math/rand"
    "time"
    "src/utils"
)

// GenerateTelemetry gera dados de telemetria simulados
func GenerateTelemetry(roverID uint8, state uint8) *TelemetryPacket {
    return &TelemetryPacket{
        RoverID:     roverID,
        Timestamp:   time.Now().Unix(),
        Position:    utils.Coordinate{Latitude: rand.Float64()*90, Longitude: rand.Float64()*180},
        State:       state,
        Battery:     uint8(75 + rand.Intn(25)), // 75-100%
        Speed:       rand.Float32() * 2.0,      // 0-2 m/s
        Temperature: int16(20 + rand.Intn(30)), // 20-50°C
        WheelStatus: 0b1111,                    // Todas as rodas OK
    }
}