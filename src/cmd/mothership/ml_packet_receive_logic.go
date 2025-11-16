package main

import (
	"fmt"
	"src/internal/ml"
	"src/utils/packetsLogic"
	"src/internal/core"
	"sync"
)

// receiver lê continuamente pacotes UDP
func (ms *MotherShip) receiver() {
	buf := make([]byte, 1024)

	for {
		n, addr, err := ms.Conn.ReadFromUDP(buf)
		if err != nil {
			fmt.Println("Erro a ler pacote:", err)
			continue
		}

		packet := ml.FromBytes(buf[:n])
		roverID := packet.RoverId

		ms.Mu.Lock()
		state, exists := ms.Rovers[roverID]
		if !exists {
			state = &core.RoverState{
				Addr:        addr,
				SeqNum:      0,
				ExpectedSeq: packet.SeqNum,
				Buffer:      make(map[uint16]ml.Packet),
				Window: &packetslogic.Window{
					LastAckReceived: -1,
					Window:          make(map[uint32](chan int8)),
					Mu:              sync.Mutex{},
				},
			}
			ms.Rovers[roverID] = state
			fmt.Printf("🆕 Novo rover registado: %d\n", roverID)
		}
		ms.Mu.Unlock()

		// Criar goroutine para processar o pacote
		go ms.handlePacket(state, packet)
	}
}
