package packetslogic

import (
	"net"
	"src/config"
	"src/internal/ml"
	"src/utils/metrics"
	"sync"
	"time"
)

// Logger is a function type for logging messages
type Logger func(level, msg string, meta any)

// Window is the sliding window structure to manage sent packets and RTO calculation
type Window struct {
	LastAckReceived int16                  // Last ACK received number
	Window          map[uint32](chan int8) // Sent packets not yet ACKed
	Mu              sync.Mutex             // Mutex for concurrent access
	// Fields for dynamic RTO calculation
	SRTT   time.Duration
	RTTVAR time.Duration
	RTO    time.Duration
}

const chanBufferSize = 1

// NewWindow creates and initializes a new Window instance
func NewWindow() *Window {
	return &Window{
		LastAckReceived: -1,
		Window:          make(map[uint32](chan int8)),
		Mu:              sync.Mutex{},
		SRTT:            0,
		RTTVAR:          0,
		RTO:             config.INITIAL_RTO, // initial fallback from config
	}
}

// GetPendingCount returns the number of packets waiting for ACK
func (w *Window) GetPendingCount() int {
	w.Mu.Lock()
	defer w.Mu.Unlock()
	return len(w.Window)
}

// WaitForWindowSlot blocks until there's room in the sliding window
// This implements flow control to prevent overwhelming the receiver
func (w *Window) WaitForWindowSlot() {
	for {
		if w.GetPendingCount() < config.MAX_PACKETS_IN_FLIGHT {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
}

// UpdateRTO updates the RTO based on a new RTT sample using TCP's algorithm technique
func (w *Window) UpdateRTO(sampleRTT time.Duration) {
	if w.SRTT == 0 {
		// First measured RTT
		w.SRTT = sampleRTT
		w.RTTVAR = sampleRTT / 2
	} else {
		w.RTTVAR = (3*w.RTTVAR + absDuration(w.SRTT-sampleRTT)) / 4
		w.SRTT = (7*w.SRTT + sampleRTT) / 8
	}

	w.RTO = w.SRTT + 4*w.RTTVAR

	// Safety limits from config
	if w.RTO < config.MIN_RTO {
		w.RTO = config.MIN_RTO
	}
	if w.RTO > config.MAX_RTO {
		w.RTO = config.MAX_RTO
	}
}

// absDuration returns the absolute value of a time.Duration
func absDuration(d time.Duration) time.Duration {
	if d < 0 {
		return -d
	}
	return d
}

// SendPacketUDP encodes and sends a packet through UDP connection
func SendPacketUDP(conn *net.UDPConn, addr *net.UDPAddr, packet ml.Packet) error {
	// Encodes the packet
	packet.Checksum = ml.Checksum(packet.Payload)
	encodedPacket := packet.Encode()

	// Sends the encoded data to the specified address port
	_, error := conn.WriteToUDP(encodedPacket, addr)
	return error
}

// CreateAndSendPacket creates a packet with auto-incremented SeqNum and sends it
// This is a generic function that handles both rover and mothership packet sending
// SeqNum is incremented by the total payload
func CreateAndSendPacket(
	conn *net.UDPConn,
	addr *net.UDPAddr,
	roverID uint8,
	msgType ml.PacketType,
	seqNum *uint16,
	ackNum uint16,
	payload []byte,
	window *Window,
	windowLock *sync.Mutex,
	logf Logger,
) {
	// Flow control: wait if too many packets are in flight (except for ACKs)
	if msgType != ml.MSG_ACK {
		window.WaitForWindowSlot()
	}

	// Lock if mutex is provided (mothership case)
	if windowLock != nil {
		windowLock.Lock()
	}

	// Create packet with current SeqNum
	pkt := ml.Packet{
		RoverId:  roverID,
		MsgType:  msgType,
		SeqNum:   *seqNum,
		AckNum:   ackNum,
		Checksum: 0,
		Payload:  payload,
	}

	// Calculate payload size for SeqNum increment (minimum 1 for empty payloads)
	payloadSize := len(payload)
	if payloadSize == 0 {
		payloadSize = 1 // Minimum increment to ensure SeqNum always advances
	}

	// Increment SeqNum by payload size
	// Use uint32 for calculation to avoid overflow, then cast back
	newSeq := uint32(*seqNum) + uint32(payloadSize)
	*seqNum = uint16(newSeq) // Automatic wraparound when casting

	// Unlock if mutex was provided
	if windowLock != nil {
		windowLock.Unlock()
	}

	// Send packet using PacketManager
	go PacketManager(conn, addr, pkt, window, logf)
}

// PacketManager manages the sending and retransmission of a packet until an ACK is received
func PacketManager(conn *net.UDPConn, addr *net.UDPAddr, pkt ml.Packet, window *Window, logf Logger) {
	// ACK packets are sent immediately without retransmission
	if pkt.MsgType == ml.MSG_ACK {
		sendAckPacket(conn, addr, pkt, logf)
		return
	}

	// For other packets, manage retransmissions with window (non-blocking)
	manageRetransmission(conn, addr, pkt, window, logf)
}

// sendAckPacket sends an ACK packet without waiting for acknowledgment
func sendAckPacket(conn *net.UDPConn, addr *net.UDPAddr, pkt ml.Packet, logf Logger) {
	if err := SendPacketUDP(conn, addr, pkt); err != nil {
		logf("ERROR", "Failed to send ACK", map[string]any{
			"ackNum": pkt.AckNum,
			"error":  err,
		})
	} else {
		// Record ACK sent metric
		if m := metrics.GetGlobalMetrics(); m != nil {
			m.RecordAckSent()
		}
		logf("INFO", "ACK sent", map[string]any{
			"ackNum": pkt.AckNum,
		})
	}
}

// manageRetransmission handles sending and retransmitting packets until ACK is received
// Implements Karn's Algorithm: RTT is only measured for packets ACKed on first attempt
// to avoid ambiguity with retransmissions (we can't know if ACK is for original or retransmit)
func manageRetransmission(conn *net.UDPConn, addr *net.UDPAddr, pkt ml.Packet, window *Window, logf Logger) {
	// Register packet in window
	ch := registerPacket(window, pkt.SeqNum)
	defer unregisterPacket(window, pkt.SeqNum)

	// Record send time only once for RTT measurement (Karn's Algorithm)
	firstSendTime := time.Now()

	for retries := 0; retries <= config.MAX_RETRIES; retries++ {
		// Send the packet
		if err := SendPacketUDP(conn, addr, pkt); err != nil {
			logf("ERROR", "Failed to send packet", map[string]any{
				"seqNum": pkt.SeqNum,
				"error":  err,
			})
			return
		}

		// Record packet sent metric
		if m := metrics.GetGlobalMetrics(); m != nil {
			pktType := ml.PacketType(pkt.MsgType).String()
			packetSize := ml.PacketHeaderSize + len(pkt.Payload)
			m.RecordPacketSent(pktType, packetSize)

			// Record retransmission if not first attempt
			if retries > 0 {
				m.RecordRetransmission()
			}
		}

		if retries == 0 {
			// Log only on first send, not retries
			pktType := ml.PacketType(pkt.MsgType).String()
			logf("INFO", "Packet sent", map[string]any{
				"type":   pktType,
				"seqNum": pkt.SeqNum,
			})
		}

		// get current RTO value
		rto := getRTO(window)

		select {
		case <-ch:
			// ACK received
			// Karn's Algorithm: Only update RTO on first attempt (no retransmissions)
			// After retransmission, we can't know if ACK is for original or retransmit
			if retries == 0 {
				rtt := time.Since(firstSendTime)
				handleAckReceived(window, rtt)

				// Record RTT metric only for clean samples
				if m := metrics.GetGlobalMetrics(); m != nil {
					m.RecordRTT(rtt)
				}
			}

			// Always record ACK received metric
			if m := metrics.GetGlobalMetrics(); m != nil {
				m.RecordAckReceived()
			}
			return

		case <-time.After(rto):
			// Timeout - prepare for retransmission
			if retries == config.MAX_RETRIES {
				// Record packet lost metric
				if m := metrics.GetGlobalMetrics(); m != nil {
					m.RecordPacketLost()
				}
				handleMaxRetriesReached(pkt.SeqNum, logf)
				return
			}
			handleTimeout(pkt.SeqNum, retries, rto, logf)
		}
	}
}

// registerPacket adds a packet to the window
func registerPacket(window *Window, seqNum uint16) chan int8 {
	window.Mu.Lock()
	defer window.Mu.Unlock()

	ch := make(chan int8, chanBufferSize)
	window.Window[uint32(seqNum)] = ch
	return ch
}

// unregisterPacket removes a packet from the window
func unregisterPacket(window *Window, seqNum uint16) {
	window.Mu.Lock()
	defer window.Mu.Unlock()
	delete(window.Window, uint32(seqNum))
}

// getRTO safely retrieves the current RTO value
func getRTO(window *Window) time.Duration {
	window.Mu.Lock()
	defer window.Mu.Unlock()
	return window.RTO
}

// handleAckReceived processes a received ACK and updates RTO
func handleAckReceived(window *Window, rtt time.Duration) {
	window.Mu.Lock()
	window.UpdateRTO(rtt)
	newRTO := window.RTO
	window.Mu.Unlock()
	// Note: logging handled by caller if needed
	_ = newRTO // Keep RTO calculation for future use
}

// handleTimeout logs timeout and prepares for retransmission
func handleTimeout(seqNum uint16, retries int, rto time.Duration, logf func(level string, msg string, meta any)) {
	logf("WARN", "Timeout, retransmitting", map[string]any{
		"seq":   seqNum,
		"retry": retries + 1,
		"rto":   rto,
	})
}

// handleMaxRetriesReached logs failure after max retries
func handleMaxRetriesReached(seqNum uint16, logf func(level string, msg string, meta any)) {
	logf("ERROR", "Failed to receive ACK after all attempts", map[string]any{
		"seq":        seqNum,
		"maxRetries": config.MAX_RETRIES,
	})
}

// SendAck sends an ACK packet for the given ackNum
// ackNum should be the next expected byte (currentSeqNum + packetSize)
func SendAck(conn *net.UDPConn, addr *net.UDPAddr, ackNum uint16, window *Window, roverId uint8, logf func(level string, msg string, meta any)) {
	ackPacket := ml.Packet{
		RoverId: roverId,
		MsgType: ml.MSG_ACK,
		SeqNum:  0,
		AckNum:  ackNum,
		Payload: []byte{},
	}

	// Use PacketManager to handle sending the ACK packet
	PacketManager(conn, addr, ackPacket, window, logf)
}

// CalculateAckNum calculates the proper AckNum for a received packet
// This follows the protocol rule: AckNum = SeqNum + max(PayloadSize, 1)
func CalculateAckNum(pkt ml.Packet) uint16 {
	payloadSize := len(pkt.Payload)
	if payloadSize == 0 {
		payloadSize = 1 // Minimum increment for empty payloads
	}
	return pkt.SeqNum + uint16(payloadSize)
}
