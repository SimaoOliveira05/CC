package main

import (
	"src/internal/ml"
	"src/utils/packetsLogic"
)

func (ms *MotherShip) sendAck(state *RoverState, ackNum uint16) {
	ackPacket := ml.Packet{
		RoverId: 0,
		MsgType: ml.MSG_ACK,
		SeqNum:  0,
		AckNum:  ackNum + 1,
		Payload: []byte{},
	}
	ackPacket.Checksum = ml.Checksum(ackPacket.Payload)

	packetslogic.PacketManager(ms.conn, state.Addr, ackPacket, state.Window)
}


