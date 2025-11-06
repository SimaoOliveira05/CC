package main

import (
	"fmt"
	"src/internal/ml"
)

func packetHandler(rover *Rover, pkt ml.Packet, c *RoverMlConection, window *Window) {
	// Lógica para tratar o pacote recebido
	switch pkt.MsgType {

		case ml.MSG_MISSION:
			rover.missionReceivedChan <- true
			go generate(ml.DataFromBytes(pkt.Payload), rover, c, window)

		case ml.MSG_NO_MISSION:
			rover.missionReceivedChan <- false

		case ml.MSG_ACK:
			window.mu.Lock()
			for i := window.lastAckReceived + 1; i <= int16(pkt.AckNum); i++ {
				if ch, exists := window.window[uint32(i)]; exists {
					ch <- 1 // Sinaliza o ACK recebido
				}
			}
			window.lastAckReceived = int16(pkt.AckNum - 1)
			window.mu.Unlock()

		default:
			fmt.Printf("⚠️ Tipo de pacote desconhecido: %d\n", pkt.MsgType)
	}
}

func receiver(rover *Rover, c *RoverMlConection, window *Window) {
	buf := make([]byte, 2048)
	// Loop de recepção
	for {
		n, _, err := c.conn.ReadFromUDP(buf)
		if err != nil {
			fmt.Println("Erro ao ler pacote UDP:", err)
			continue
		}

		// Constrói o pacote a partir dos bytes recebidos e trata-o
		pkt := ml.FromBytes(buf[:n])
		fmt.Printf("📨 Pacote recebido do tipo: %d\n", pkt.MsgType)
		
		packetHandler(rover, pkt, c, window)
	}
}
