package packetslogic

import (
	"fmt"
	"src/internal/ml"
)

// handleACK processa confirmações de entrega
func HandleAck(p ml.Packet, window *Window) {
	window.Mu.Lock()
	for i := window.LastAckReceived + 1; i < int16(p.AckNum); i++ {
		if ch, exists := window.Window[uint32(i)]; exists {
			ch <- 1 // Sinaliza o ACK recebido
			delete(window.Window, uint32(i))
		}
	}
	window.LastAckReceived = int16(p.AckNum - 1)
	window.Mu.Unlock()
	fmt.Printf("✅ ACK recebido, AckNum: %d (confirmou até SeqNum %d)\n", p.AckNum, p.AckNum-1)
}
