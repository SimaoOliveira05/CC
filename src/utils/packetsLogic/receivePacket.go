package packetslogic

import (
	"net"
	"src/internal/ml"
	"src/utils/metrics"
	"sync"
)

// seqLessThan compares sequence numbers considering wraparound (RFC 1982)
// Returns true if seq1 is "less than" seq2 in circular arithmetic
func seqLessThan(seq1, seq2 uint16) bool {
	return int16(seq1-seq2) < 0
}

// seqGreaterThan compares sequence numbers considering wraparound
// Returns true if seq1 is "greater than" seq2 in circular arithmetic
func seqGreaterThan(seq1, seq2 uint16) bool {
	return int16(seq1-seq2) > 0
}

// handleACK processes delivery acknowledgments and updates the sliding window
// ACK numbers represent bytes acknowledged (TCP-style)
func HandleAck(p ml.Packet, window *Window) {
	ProcessAckNum(p.AckNum, window)
}

// ProcessAckNum processes an AckNum and signals waiting goroutines
// This handles both explicit ACKs and implicit ACKs (e.g., MSG_MISSION responding to MSG_REQUEST)
func ProcessAckNum(ackNum uint16, window *Window) {
	if ackNum == 0 {
		return // No ACK to process
	}

	window.Mu.Lock()
	defer window.Mu.Unlock()

	// Mark all packets with SeqNum < AckNum as acknowledged (considering wraparound)
	// In TCP-style, AckNum represents the next byte expected
	for seqKey, ch := range window.Window {
		if seqLessThan(uint16(seqKey), ackNum) {
			select {
			case ch <- 1: // Signal ACK received
			default:
				// Channel might be full or closed
			}
			delete(window.Window, seqKey)
		}
	}

	// Update LastAckReceived considering wraparound
	if seqGreaterThan(ackNum-1, uint16(window.LastAckReceived)) {
		window.LastAckReceived = int16(ackNum - 1)
	}
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
		// Record metric
		if m := metrics.GlobalMetrics; m != nil {
			m.RecordChecksumFailed()
		}
		return
	}

	// Record valid packet received
	if m := metrics.GlobalMetrics; m != nil {
		packetSize := ml.PacketHeaderSize + len(pkt.Payload)
		m.RecordPacketReceived(pkt.MsgType.String(), packetSize)
	}

	// ALWAYS process implicit ACK if AckNum > 0
	// This handles MSG_MISSION/MSG_NO_MISSION acting as ACK for MSG_REQUEST
	if pkt.AckNum > 0 {
		ProcessAckNum(pkt.AckNum, window)
	}

	// If processing without ordering (pure ACKs), we're done after processing AckNum
	if skipOrdering {
		go processor(pkt)
		return
	}

	// Sliding window ordering logic
	mu.Lock()
	defer mu.Unlock()

	seq := pkt.SeqNum
	expected := *expectedSeq

	// Calculate packet size for next expected SeqNum
	// Use payload size, but minimum 1 to ensure SeqNum always advances
	payloadSize := len(pkt.Payload)
	if payloadSize == 0 {
		payloadSize = 1 // Minimum increment to avoid deadlock with empty payloads
	}
	nextExpected := uint16(uint32(seq) + uint32(payloadSize)) // Wraparound via cast

	switch {
	case seq == expected:
		// Expected packet - process and advance window
		go processor(pkt)
		*expectedSeq = nextExpected

		// Process consecutive buffered packets WITHOUT sending individual ACKs
		bufferedCount := 0
		for {
			if bufferedPkt, ok := buffer[*expectedSeq]; ok {
				delete(buffer, *expectedSeq)
				go processor(bufferedPkt)
				// Calculate buffered packet size (minimum 1 for empty payloads)
				bufferedPayloadSize := len(bufferedPkt.Payload)
				if bufferedPayloadSize == 0 {
					bufferedPayloadSize = 1
				}
				*expectedSeq = uint16(uint32(*expectedSeq) + uint32(bufferedPayloadSize))
				bufferedCount++
			} else {
				break
			}
		}

		// Send ONE cumulative ACK at the end (covers all processed packets)
		if autoAck {
			if bufferedCount > 0 {
				logf("INFO", "CUMULATIVE ACK - processed buffered packets", map[string]any{
					"bufferedCount": bufferedCount,
					"cumulativeAck": *expectedSeq,
					"originalSeq":   seq,
				})
			}
			SendAck(conn, addr, *expectedSeq, window, roverID, logf)
		}

	case seqGreaterThan(seq, expected):
		// Out-of-order packet - buffer and send cumulative ACK
		buffer[seq] = pkt
		SendAck(conn, addr, expected, window, roverID, logf)
		// Record out-of-order metric
		if m := metrics.GlobalMetrics; m != nil {
			m.RecordOutOfOrder()
		}

	case seqLessThan(seq, expected):
		// Duplicate packet - resend ACK
		SendAck(conn, addr, nextExpected, window, roverID, logf)
		// Record duplicate metric
		if m := metrics.GlobalMetrics; m != nil {
			m.RecordDuplicateReceived()
		}
	}
}
