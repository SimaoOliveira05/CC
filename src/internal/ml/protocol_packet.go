package ml

import (
	"encoding/binary"
)

// Message types.
const (
	MSG_REQUEST      = 0
	MSG_MISSION      = 1
	MSG_ACK          = 2
	MSG_REPORT       = 3
	MSG_NO_MISSION   = 4
	MSG_STATE_UPDATE = 5
)

// Packet is the base structure of the packet.
type Packet struct {
	RoverId  uint8
	MsgType  uint8
	SeqNum   uint16
	AckNum   uint16
	Checksum uint8
	Payload  []byte
}

// PacketHeaderSize is the size of the packet header in bytes.
const PacketHeaderSize = 7 // 1 (RoverId) + 1 (MsgType) + 2 (SeqNum) + 2 (AckNum) + 1 (Checksum) - Payload é variável


// Enconde serializes the packet into bytes.
func (p *Packet) Encode() []byte {
	// Calculate total size of the packet
    totalSize := PacketHeaderSize + len(p.Payload)
    data := make([]byte, totalSize)
    
    data[0] = p.RoverId
    data[1] = p.MsgType
    binary.BigEndian.PutUint16(data[2:], p.SeqNum)
    binary.BigEndian.PutUint16(data[4:], p.AckNum)
    data[6] = p.Checksum
    copy(data[7:], p.Payload)
    
    return data
}

// Decode deserializes bytes into a Packet (BigEndian).
func (p *Packet) Decode(data []byte) {  
    p.RoverId = data[0]
    p.MsgType = data[1]
    p.SeqNum = binary.BigEndian.Uint16(data[2:])
    p.AckNum = binary.BigEndian.Uint16(data[4:])
    p.Checksum = data[6]
    
    if len(data) > PacketHeaderSize {
        p.Payload = make([]byte, len(data)-PacketHeaderSize)
        copy(p.Payload, data[7:])
    }
}

// Simple checksum calculation.
func Checksum(data []byte) uint8 {
	var sum uint32
	for _, b := range data {
		sum += uint32(b)
	}
	return uint8(sum % 256)
}
