package ml

import (
	"encoding/binary"
)

// Message types.
const (
    MSG_MISSION    PacketType = 0
    MSG_NO_MISSION PacketType = 1
    MSG_ACK        PacketType = 2
    MSG_REPORT     PacketType = 3
    MSG_REQUEST    PacketType = 4
)

// PacketType represents the type of message
type PacketType uint8

// Packet is the base structure of the packet.
type Packet struct {
	RoverId  uint8
	MsgType  PacketType
	SeqNum   uint16
	AckNum   uint16
	Checksum uint8
	Payload  []byte
}

// String returns the string representation of PacketType
func (pt PacketType) String() string {
    switch pt {
    case MSG_MISSION:
        return "MSG_MISSION"
    case MSG_NO_MISSION:
        return "MSG_NO_MISSION"
    case MSG_ACK:
        return "MSG_ACK"
    case MSG_REPORT:
        return "MSG_REPORT"
    case MSG_REQUEST:
        return "MSG_REQUEST"
    default:
        return "UNKNOWN"
    }
}

// PacketHeaderSize is the size of the packet header in bytes.
const PacketHeaderSize = 7 // 1 (RoverId) + 1 (MsgType) + 2 (SeqNum) + 2 (AckNum) + 1 (Checksum) - Payload é variável


// Enconde serializes the packet into bytes.
func (p *Packet) Encode() []byte {
	// Calculate total size of the packet
    totalSize := PacketHeaderSize + len(p.Payload)
    data := make([]byte, totalSize)
    
    data[0] = p.RoverId
    data[1] = uint8(p.MsgType)
    binary.BigEndian.PutUint16(data[2:], p.SeqNum)
    binary.BigEndian.PutUint16(data[4:], p.AckNum)
    data[6] = p.Checksum
    copy(data[7:], p.Payload)
    
    return data
}

// Decode deserializes bytes into a Packet (BigEndian).
func (p *Packet) Decode(data []byte) {  
    p.RoverId = data[0]
    p.MsgType = PacketType(data[1])
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
