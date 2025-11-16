package main

import (
	"fmt"
	"src/internal/ml"
	packetslogic "src/utils/packetsLogic"
)


// handlePacket processa cada pacote recebido
func (rv *Rover) handlePacket(pkt ml.Packet) {
	// Closure que captura 'rv'
	processor := func(p ml.Packet) {
		switch p.MsgType {
		case ml.MSG_MISSION:
			rv.processMission(p)
		case ml.MSG_NO_MISSION:
			rv.missionReceivedChan <- false
		case ml.MSG_ACK:
			packetslogic.HandleAck(p, rv.window) // ✅ Usa 'p' (parâmetro da closure)
		default:
			fmt.Printf("⚠️ Tipo de pacote desconhecido: %d\n", p.MsgType)
		}
	}

	packetslogic.HandleOrderedPacket(
		pkt,
		&rv.expectedSeq,
		rv.buffer,
		&rv.bufferMu,
		rv.conn.conn,
		rv.conn.addr,
		rv.window,
		rv.id,
		processor,
		pkt.MsgType == ml.MSG_ACK, // ✅ Skip ordering para ACKs
	)
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
		rv.handlePacket(pkt)
	}
}
