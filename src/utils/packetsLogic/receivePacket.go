package packetslogic

import (
	"net"
	"src/internal/ml"
	"sync"
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
}

// PacketProcessor é a função callback para processar um pacote após ordenação
type PacketProcessor func(pkt ml.Packet)

// HandleOrderedPacket processa pacotes com ordenação e verificação de checksum
// Parâmetros:
//   - pkt: pacote recebido
//   - expectedSeq: ponteiro para o número de sequência esperado
//   - buffer: buffer de pacotes fora de ordem
//   - mu: mutex para proteger acesso ao estado
//   - conn: conexão UDP
//   - addr: endereço do remetente
//   - window: janela de controlo de fluxo
//   - roverID: ID do rover (0 para MotherShip)
//   - processor: função callback para processar o pacote
//   - skipOrdering: se true, processa sem ordenação (ex: ACKs)
func HandleOrderedPacket(
	pkt ml.Packet,
	expectedSeq *uint16,
	buffer map[uint16]ml.Packet,
	mu *sync.Mutex,
	conn *net.UDPConn,
	addr *net.UDPAddr,
	window *Window,
	roverID uint8,
	processor PacketProcessor,
	skipOrdering bool,
	autoAck bool,
	logf func(level string, msg string, meta any),
) {
	// 1. Verificar checksum ANTES de adquirir lock
	expectedChecksum := ml.Checksum(pkt.Payload)
	if pkt.Checksum != expectedChecksum {
		logf("ERROR", "Checksum inválido, pacote descartado", map[string]any{
			"addr":     addr.String(),
			"expected": expectedChecksum,
			"received": pkt.Checksum,
		})
		return
	}

	// 2. Se for para processar sem ordenação (ACKs), processa diretamente
	if skipOrdering {
		go processor(pkt)
		return
	}

	// 3. Lógica de ordenação com janela deslizante
	mu.Lock()
	defer mu.Unlock()

	seq := pkt.SeqNum
	expected := *expectedSeq

	switch {
	case seq == expected:
		// Pacote esperado - processa e avança janela
		go processor(pkt)
		*expectedSeq++
		if autoAck {
			SendAck(conn, addr, seq, window, roverID, logf)
		}

		// Processa pacotes bufferizados consecutivos
		for {
			if bufferedPkt, ok := buffer[*expectedSeq]; ok {
				delete(buffer, *expectedSeq)
				go processor(bufferedPkt)
				SendAck(conn, addr, *expectedSeq, window, roverID, logf)
				*expectedSeq++
			} else {
				break
			}
		}

	case seq > expected:
		// Pacote fora de ordem - bufferiza e envia ACK cumulativo
		buffer[seq] = pkt
		SendAck(conn, addr, expected, window, roverID, logf)

	case seq < expected:
		// Pacote duplicado - reenvia ACK
		SendAck(conn, addr, seq, window, roverID, logf)
	}
}
