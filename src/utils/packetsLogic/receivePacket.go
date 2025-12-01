package packetslogic

import (
	"net"
	"src/internal/ml"
	"sync"
)

// handleACK processes delivery acknowledgments and updates the sliding window
func HandleAck(p ml.Packet, window *Window) {
	window.Mu.Lock()
	for i := window.LastAckReceived + 1; i < int16(p.AckNum); i++ {
		if ch, exists := window.Window[uint32(i)]; exists {
			ch <- 1 // Signal ACK received
			delete(window.Window, uint32(i))
		}
	}
	window.LastAckReceived = int16(p.AckNum - 1)
	window.Mu.Unlock()
}

// PacketProcessor is the callback function to process a packet after ordering
type PacketProcessor func(pkt ml.Packet)

// HandleOrderedPacket processes packets with ordering and checksum verification
// Parameters:
//   - pkt: received packet
//   - expectedSeq: pointer to the expected sequence number
//   - buffer: buffer for out-of-order packets
//   - mu: mutex to protect state access
//   - conn: UDP connection
//   - addr: sender's address
//   - window: flow control window
//   - roverID: rover ID (0 for MotherShip)
//   - processor: callback function to process the packet
//   - skipOrdering: if true, process without ordering (e.g., ACKs)
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
	// Verify checksum before any processing
	expectedChecksum := ml.Checksum(pkt.Payload)
	if pkt.Checksum != expectedChecksum {
		logf("ERROR", "Invalid checksum, packet discarded", map[string]any{
			"addr":     addr.String(),
			"expected": expectedChecksum,
			"received": pkt.Checksum,
		})
		return
	}

	// If processing without ordering (ACKs), process directly
	if skipOrdering {
		go processor(pkt)
		return
	}

	// Sliding window ordering logic
	mu.Lock()
	defer mu.Unlock()

	seq := pkt.SeqNum
	expected := *expectedSeq

	switch {
	case seq == expected:
		// Expected packet - process and advance window
		go processor(pkt)
		*expectedSeq++
		if autoAck {
			SendAck(conn, addr, seq, window, roverID, logf)
		}

		// Process consecutive buffered packets
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
		// Out-of-order packet - buffer and send cumulative ACK
		buffer[seq] = pkt
		SendAck(conn, addr, expected, window, roverID, logf)

	case seq < expected:
		// Duplicate packet - resend ACK
		SendAck(conn, addr, seq, window, roverID, logf)
	}
}
