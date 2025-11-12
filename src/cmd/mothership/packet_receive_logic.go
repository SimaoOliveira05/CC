package main

import (
	"fmt"
	"src/internal/ml"
	"sync"
)

// receiver lê continuamente pacotes UDP
func (ms *MotherShip) receiver() {
	buf := make([]byte, 1024)

	for {
		n, addr, err := ms.conn.ReadFromUDP(buf)
		if err != nil {
			fmt.Println("Erro a ler pacote:", err)
			continue
		}

		packet := ml.FromBytes(buf[:n])
		roverID := addr.String()

		ms.mu.Lock()
		state, exists := ms.rovers[roverID]
		if !exists {
			state = &RoverState{
				Addr:        addr,
				SeqNum:      0,
				ExpectedSeq: packet.SeqNum,
				Buffer:      make(map[uint16]ml.Packet),
				Window: &Window{
					lastAckReceived: -1,
					window:          make(map[uint32](chan int8)),
					mu:              sync.Mutex{},
				},
			}
			ms.rovers[roverID] = state
			fmt.Printf("🆕 Novo rover registado: %s\n", roverID)
		}
		ms.mu.Unlock()

		// Criar goroutine para processar o pacote
		go ms.handlePacket(state, packet)
	}
}
