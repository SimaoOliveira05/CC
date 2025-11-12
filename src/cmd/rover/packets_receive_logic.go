package main

import (
	"fmt"
	"src/internal/ml"
)

func (rv *Rover) packetHandler(pkt ml.Packet) {
	// Lógica para tratar o pacote recebido
	switch pkt.MsgType {

	case ml.MSG_MISSION:
		rv.handleMissionPacket(pkt)

	case ml.MSG_NO_MISSION:
		rv.handleMissionPacket(pkt)

	case ml.MSG_ACK:
		rv.window.mu.Lock()
		for i := rv.window.lastAckReceived + 1; i < int16(pkt.AckNum); i++ {
			if ch, exists := rv.window.window[uint32(i)]; exists {
				ch <- 1 // Sinaliza o ACK recebido
				delete(rv.window.window, uint32(i))
			}
		}
		rv.window.lastAckReceived = int16(pkt.AckNum - 1)
		rv.window.mu.Unlock()

	default:
		fmt.Printf("⚠️ Tipo de pacote desconhecido: %d\n", pkt.MsgType)
	}
}

// handleNoMissionPacket processa pacotes NO_MISSION com ordenação
func (rv *Rover) handleMissionPacket(pkt ml.Packet) {
	rv.bufferMu.Lock()
	defer rv.bufferMu.Unlock()

	seq := pkt.SeqNum
	expected := rv.expectedSeq

	switch {
	case seq == expected:
		// Pacote esperado
		switch pkt.MsgType {
		case ml.MSG_MISSION:
			rv.processMission(pkt)
			rv.sendAck(seq)
		case ml.MSG_NO_MISSION:
			rv.sendAck(seq)
			rv.missionReceivedChan <- false
		}
		rv.expectedSeq++

		// Processa pacotes bufferizados consecutivos
		for {
			if bufferedPkt, ok := rv.buffer[rv.expectedSeq]; ok {
				delete(rv.buffer, rv.expectedSeq)
				switch bufferedPkt.MsgType {
				case ml.MSG_NO_MISSION:
					rv.sendAck(rv.expectedSeq)
					rv.missionReceivedChan <- false
				case ml.MSG_MISSION:
					rv.processMission(bufferedPkt)
					rv.sendAck(rv.expectedSeq)
				}
				rv.expectedSeq++
			} else {
				break
			}
		}

	case seq > expected:
		// Fora de ordem — guarda no buffer e envia ACK cumulativo
		rv.buffer[seq] = pkt
		rv.sendAck(expected)

		// case seq < expected: pacote duplicado, ignora
	}
}

// processMission extrai e processa a missão
func (rv *Rover) processMission(pkt ml.Packet) {
	rv.missionReceivedChan <- true
	go rv.generate(ml.DataFromBytes(pkt.Payload))
}

func (rv *Rover) receiver() {
	buf := make([]byte, 2048)
	// Loop de recepção
	for {
		n, _, err := rv.conn.conn.ReadFromUDP(buf)
		if err != nil {
			fmt.Println("Erro ao ler pacote UDP:", err)
			continue
		}

		// Constrói o pacote a partir dos bytes recebidos e trata-o
		pkt := ml.FromBytes(buf[:n])
		rv.packetHandler(pkt)
	}
}
