package ts

import (
	"bytes"
	"encoding/binary"
	"src/utils"
)

// Estado operacional do rover
const (
    STATE_IDLE        = 0  // Parado
    STATE_IN_MISSION  = 1  // Em missão
    STATE_TRAVELING   = 2  // A caminho
    STATE_ERROR       = 3  // Erro
)

type TelemetryPacket struct {
    RoverID       uint8              // ID do rover
    Timestamp     int64              // Unix timestamp
    Position      utils.Coordinate   // (Latitude, Longitude)
    State         uint8              // Estado operacional (4 bits)
    Battery       uint8              // Nível de bateria (0-100%)
    Speed         float32            // Velocidade em m/s
    Temperature   int16              // Temperatura interna (°C * 10)
    WheelStatus   uint8              // Status das 4 rodas (4 bits) wheel1|wheel2|wheel3|wheel4
}

// ToBytes serializa o pacote de telemetria
// State e WheelStatus são combinados num único byte: [4 bits State | 4 bits WheelStatus]
func (t *TelemetryPacket) ToBytes() []byte {
    buf := new(bytes.Buffer)
    binary.Write(buf, binary.BigEndian, t.RoverID)
    binary.Write(buf, binary.BigEndian, t.Timestamp)
    binary.Write(buf, binary.BigEndian, t.Position.Latitude)
    binary.Write(buf, binary.BigEndian, t.Position.Longitude)
    // Combina State (bits 4-5) e WheelStatus (bits 0-3) num byte
    StateAndWheelsStatus := (t.State << 4) | (t.WheelStatus & 0x0F)
    binary.Write(buf, binary.BigEndian, StateAndWheelsStatus)
    binary.Write(buf, binary.BigEndian, t.Battery)
    binary.Write(buf, binary.BigEndian, t.Speed)
    binary.Write(buf, binary.BigEndian, t.Temperature)
    return buf.Bytes()
}

// FromBytes deserializa bytes em TelemetryPacket
// State e WheelStatus são extraídos de 1 byte combinado: [4 bits State | 4 bits WheelStatus]
func FromBytes(data []byte) (*TelemetryPacket, error) {
    buf := bytes.NewReader(data)
    t := &TelemetryPacket{}
    binary.Read(buf, binary.BigEndian, &t.RoverID)
    binary.Read(buf, binary.BigEndian, &t.Timestamp)
    binary.Read(buf, binary.BigEndian, &t.Position.Latitude)
    binary.Read(buf, binary.BigEndian, &t.Position.Longitude)
    // Lê o byte combinado e separa State e WheelStatus
    var StateAndWheelsStatus uint8
    binary.Read(buf, binary.BigEndian, &StateAndWheelsStatus)
    t.State = StateAndWheelsStatus >> 4       // 4 bits para State (bits 4-7)
    t.WheelStatus = StateAndWheelsStatus & 0x0F         // 4 bits para WheelStatus (bits 0-3)
    binary.Read(buf, binary.BigEndian, &t.Battery)
    binary.Read(buf, binary.BigEndian, &t.Speed)
    binary.Read(buf, binary.BigEndian, &t.Temperature)
    return t, nil
}