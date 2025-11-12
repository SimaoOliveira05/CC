package main

import (
	"fmt"
	"src/internal/ml"
	"time"
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

	if _, err := ms.conn.WriteToUDP(ackPacket.ToBytes(), state.Addr); err != nil {
		fmt.Println("❌ Erro ao enviar ACK:", err)
		return
	}
	fmt.Printf("📤 ACK enviado para %s, AckNum: %d\n", state.Addr, ackNum)
}

// sendPacket envia um pacote e gerencia retransmissões até receber o ACK
func (ms *MotherShip) sendPacket(pkt ml.Packet, state *RoverState) {
	state.Window.mu.Lock()
	ch := make(chan int8, 1)
	state.Window.window[uint32(pkt.SeqNum)] = ch
	state.Window.mu.Unlock()
	go ms.packetManager(pkt, ch, state)
}

// packetManager gerencia o envio e retransmissão de um pacote até receber o ACK
func (ms *MotherShip) packetManager(pkt ml.Packet, ch chan int8, state *RoverState) {
	retries := 0
	maxRetries := 5
	for {
		// Envia o pacote
		_, err := ms.conn.WriteToUDP(pkt.ToBytes(), state.Addr)
		if err != nil {
			fmt.Println("❌ Erro ao enviar pacote:", err)
			return
		}
		if pkt.MsgType == ml.MSG_MISSION {
			fmt.Printf("📤 Missão enviada, SeqNum: %d\n", pkt.SeqNum)
		}
		if pkt.MsgType == ml.MSG_NO_MISSION {
			fmt.Printf("📤 NO_MISSION enviado, SeqNum: %d\n", pkt.SeqNum)
		}

		select {
		case <-ch:
			fmt.Printf("✅ ACK confirmado para SeqNum %d\n", pkt.SeqNum)
			return
		case <-time.After(100 * time.Millisecond):
			retries++
			if retries > maxRetries {
				fmt.Printf("❌ Falha ao receber ACK para SeqNum %d após %d tentativas. Abortando...\n", pkt.SeqNum, maxRetries)
				return
			}
			fmt.Printf("⏱️ Timeout esperando ACK para SeqNum %d. Retransmitindo (tentativa %d)...\n", pkt.SeqNum, retries)
		}
	}
}
