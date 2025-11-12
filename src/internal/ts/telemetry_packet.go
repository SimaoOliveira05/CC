package ts

import (
    "encoding/binary"
    "bytes"
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
    State         uint8              // Estado operacional
    Battery       uint8              // Nível de bateria (0-100%)
    Speed         float32            // Velocidade em m/s
    Temperature   int16              // Temperatura interna (°C * 10)
    WheelStatus   uint8              // Bits: wheel1|wheel2|wheel3|wheel4
}

// ToBytes serializa o pacote de telemetria
func (t *TelemetryPacket) ToBytes() []byte {
    buf := new(bytes.Buffer)
    binary.Write(buf, binary.BigEndian, t.RoverID)
    binary.Write(buf, binary.BigEndian, t.Timestamp)
    binary.Write(buf, binary.BigEndian, t.Position.Latitude)
    binary.Write(buf, binary.BigEndian, t.Position.Longitude)
    binary.Write(buf, binary.BigEndian, t.State)
    binary.Write(buf, binary.BigEndian, t.Battery)
    binary.Write(buf, binary.BigEndian, t.Speed)
    binary.Write(buf, binary.BigEndian, t.Temperature)
    binary.Write(buf, binary.BigEndian, t.WheelStatus)
    return buf.Bytes()
}

// FromBytes deserializa bytes em TelemetryPacket
func FromBytes(data []byte) (*TelemetryPacket, error) {
    buf := bytes.NewReader(data)
    t := &TelemetryPacket{}
    binary.Read(buf, binary.BigEndian, &t.RoverID)
    binary.Read(buf, binary.BigEndian, &t.Timestamp)
    binary.Read(buf, binary.BigEndian, &t.Position.Latitude)
    binary.Read(buf, binary.BigEndian, &t.Position.Longitude)
    binary.Read(buf, binary.BigEndian, &t.State)
    binary.Read(buf, binary.BigEndian, &t.Battery)
    binary.Read(buf, binary.BigEndian, &t.Speed)
    binary.Read(buf, binary.BigEndian, &t.Temperature)
    binary.Read(buf, binary.BigEndian, &t.WheelStatus)
    return t, nil
}