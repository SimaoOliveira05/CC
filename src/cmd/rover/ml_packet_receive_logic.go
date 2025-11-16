package main

import (
	"fmt"
	"src/internal/ml"
	pl "src/utils/packetsLogic"
)

// handlePacket processa cada pacote recebido
func (rover *Rover) handlePacket(pkt ml.Packet) {
	// Closure que captura 'rover'
	processor := func(p ml.Packet) {
		switch p.MsgType {
		case ml.MSG_MISSION:
			rover.processMission(p)
		case ml.MSG_NO_MISSION:
			rover.ML.MissionReceivedChan <- false
		case ml.MSG_ACK:
			pl.HandleAck(p, rover.ML.Window) // ✅ Usa 'p' (parâmetro da closure)
		default:
			fmt.Printf("⚠️ Tipo de pacote desconhecido: %d\n", p.MsgType)
		}
	}

	pl.HandleOrderedPacket(
		pkt,
		&rover.ML.ExpectedSeq,
		rover.ML.Buffer,
		&rover.ML.CondMu,
		rover.MLConn.Conn,
		rover.MLConn.Addr,
		rover.ML.Window,
		rover.ID,
		processor,
		pkt.MsgType == ml.MSG_ACK, // ✅ Skip ordering para ACKs
	)
}

// processMission extrai e processa a missão
func (rover *Rover) processMission(pkt ml.Packet) {
	rover.ML.MissionReceivedChan <- true
	go rover.generate(ml.DataFromBytes(pkt.Payload))
}

func (rover *Rover) receiver() {
	buf := make([]byte, 2048)
	// Loop de recepção
	for {
		n, _, err := rover.MLConn.Conn.ReadFromUDP(buf)
		if err != nil {
			fmt.Println("Erro ao ler pacote UDP:", err)
			continue
		}

		// Constrói o pacote a partir dos bytes recebidos e trata-o
		pkt := ml.FromBytes(buf[:n])
		rover.handlePacket(pkt)
	}
}
