package ml

import (
	"encoding/binary"
)

// Message types.
const (
	MSG_MISSION PacketType = iota
	MSG_NO_MISSION
	MSG_ACK
	MSG_REPORT
	MSG_REQUEST
)

// PacketType represents the type of message
type PacketType uint8

// Packet is the base structure of the packet.
type Packet struct {
	RoverId  uint8
	MsgType  PacketType
	SeqNum   uint32
	AckNum   uint32
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
const PacketHeaderSize = 11 // 1 (RoverId) + 1 (MsgType) + 4 (SeqNum) + 4 (AckNum) + 1 (Checksum) - Payload é variável

// Enconde serializes the packet into bytes.
func (p *Packet) Encode() []byte {
	// Calculate total size of the packet
	totalSize := PacketHeaderSize + len(p.Payload)
	data := make([]byte, totalSize)

	data[0] = p.RoverId
	data[1] = uint8(p.MsgType)
	binary.BigEndian.PutUint32(data[2:6], p.SeqNum)
	binary.BigEndian.PutUint32(data[6:10], p.AckNum)
	data[10] = p.Checksum
	copy(data[PacketHeaderSize:], p.Payload)

	return data
}

// Decode deserializes bytes into a Packet (BigEndian).
func (p *Packet) Decode(data []byte) {
	p.RoverId = data[0]
	p.MsgType = PacketType(data[1])
	p.SeqNum = binary.BigEndian.Uint32(data[2:6])
	p.AckNum = binary.BigEndian.Uint32(data[6:10])
	p.Checksum = data[10]

	if len(data) > PacketHeaderSize {
		p.Payload = make([]byte, len(data)-PacketHeaderSize)
		copy(p.Payload, data[PacketHeaderSize:])
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
