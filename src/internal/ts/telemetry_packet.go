package ts

import (
	"encoding/binary"
	"math"
	"src/utils"
)

// Operational states of the rover.
const (
    STATE_IDLE        = 0  // Idle
    STATE_IN_MISSION  = 1  // In mission
    STATE_TRAVELING   = 2  // Traveling
    STATE_ERROR       = 3  // Error
)

type TelemetryPacket struct {
    RoverID       uint8              // Rover ID
    Timestamp     int64              // Unix timestamp
    Position      utils.Coordinate   // (Latitude, Longitude)
    State         uint8              // Operational state (4 bits)
    Battery       uint8              // Battery level (0-100%)
    Speed         float32            // Speed in m/s
    Temperature   int16              // Internal temperature (°C * 10)
    WheelStatus   uint8              // Status of the 4 wheels (4 bits) wheel1|wheel2|wheel3|wheel4
}

// TelemetryPacketSize is the size in bytes of the serialized TelemetryPacket.
const TelemetryPacketSize = 33 // 1 (RoverID) + 8 (Timestamp) + 8 (Latitude) + 8 (Longitude) + 1 (State + WheelStatus) + 1 (Battery) + 4 (Speed) + 2 (Temperature)

// Encode serializes the TelemetryPacket data to bytes (BigEndian).
func (t *TelemetryPacket) Encode() []byte {
    data := make([]byte, TelemetryPacketSize)
    data[0] = t.RoverID
    binary.BigEndian.PutUint64(data[1:], uint64(t.Timestamp))
    binary.BigEndian.PutUint64(data[9:], math.Float64bits(t.Position.Latitude))
    binary.BigEndian.PutUint64(data[17:], math.Float64bits(t.Position.Longitude))
    // Combine State (bits 4-7) and WheelStatus (bits 0-3) into one byte
    StateAndWheelsStatus := (t.State << 4) | (t.WheelStatus & 0x0F)
    data[25] = StateAndWheelsStatus
    data[26] = t.Battery
    binary.BigEndian.PutUint32(data[27:], math.Float32bits(t.Speed))
    binary.BigEndian.PutUint16(data[31:], uint16(t.Temperature))
    return data
}

// Decode deserializes bytes into TelemetryPacket data (BigEndian).
func (t *TelemetryPacket) Decode(data []byte) error {
    t.RoverID = data[0]
    t.Timestamp = int64(binary.BigEndian.Uint64(data[1:]))
    t.Position.Latitude = math.Float64frombits(binary.BigEndian.Uint64(data[9:]))
    t.Position.Longitude = math.Float64frombits(binary.BigEndian.Uint64(data[17:]))
    StateAndWheelsStatus := data[25]
    t.State = StateAndWheelsStatus >> 4
    t.WheelStatus = StateAndWheelsStatus & 0x0F
    t.Battery = data[26]
    t.Speed = math.Float32frombits(binary.BigEndian.Uint32(data[27:]))
    t.Temperature = int16(binary.BigEndian.Uint16(data[31:]))
    return nil
}