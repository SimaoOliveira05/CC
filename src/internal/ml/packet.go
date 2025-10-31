package ml

import (
    "bytes"
    "encoding/binary"
)

// Tipos de mensagens
const (
    MSG_REQUEST     = 0
    MSG_MISSION     = 1
    MSG_ACK         = 2
    MSG_REPORT      = 3
    MSG_MISSION_END = 4
)

// Estrutura base do pacote
type Packet struct {
    MsgType  uint8
    SeqNum   uint16
    AckNum   uint16
    Checksum uint8
    Payload  []byte
}

// Serializa o pacote em bytes
func (p *Packet) ToBytes() []byte {
    buf := new(bytes.Buffer)
    binary.Write(buf, binary.BigEndian, p.MsgType)
    binary.Write(buf, binary.BigEndian, p.SeqNum)
    binary.Write(buf, binary.BigEndian, p.AckNum)
    binary.Write(buf, binary.BigEndian, p.Checksum)
    buf.Write(p.Payload)
    return buf.Bytes()
}

// Lê bytes e cria um Packet
func FromBytes(data []byte) Packet {
    var p Packet
    buf := bytes.NewReader(data)
    binary.Read(buf, binary.BigEndian, &p.MsgType)
    binary.Read(buf, binary.BigEndian, &p.SeqNum)
    binary.Read(buf, binary.BigEndian, &p.AckNum)
    binary.Read(buf, binary.BigEndian, &p.Checksum)
    p.Payload = make([]byte, len(data)-6)
    buf.Read(p.Payload)
    return p
}

// Checksum simples
func Checksum(data []byte) uint8 {
    var sum uint32
    for _, b := range data {
        sum += uint32(b)
    }
    return uint8(sum % 256)
}
