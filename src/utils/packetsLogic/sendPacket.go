package packetslogic

import (
	"fmt"
	"net"
	"src/internal/ml"
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

// Constants for RTO calculation and retransmission management
const (
	initialRTO     = 1200 * time.Millisecond
	minRTO         = 200 * time.Millisecond
	maxRTO         = 5 * time.Second
	maxRetries     = 5
	chanBufferSize = 1
)

// NewWindow creates and initializes a new Window instance
func NewWindow() *Window {
	return &Window{
		LastAckReceived: -1,
		Window:          make(map[uint32](chan int8)),
		Mu:              sync.Mutex{},
		SRTT:            0,
		RTTVAR:          0,
		RTO:             initialRTO, // initial fallback
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

	// Safety limits
	if w.RTO < minRTO {
		w.RTO = minRTO
	}
	if w.RTO > maxRTO {
		w.RTO = maxRTO
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

	pktType := ml.PacketType(packet.MsgType).String()

	fmt.Printf("[PACKET-LOGIC] Sent packet of type %s with seq: %d to %s\n", pktType, packet.SeqNum, addr.String())
	return error
}

// CreateAndSendPacket creates a packet with auto-incremented SeqNum and sends it
// This is a generic function that handles both rover and mothership packet sending
// windowLock can be nil if no locking is needed (e.g., for rover which has its own locking strategy)
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

	// Increment SeqNum for next packet
	*seqNum++

	// Unlock if mutex was provided
	if windowLock != nil {
		windowLock.Unlock()
	}

	// Send packet using PacketManager
	PacketManager(conn, addr, pkt, window, logf)
}

// PacketManager manages the sending and retransmission of a packet until an ACK is received
func PacketManager(conn *net.UDPConn, addr *net.UDPAddr, pkt ml.Packet, window *Window, logf Logger) {
	// ACK packets are sent immediately without retransmission
	if pkt.MsgType == ml.MSG_ACK {
		sendAckPacket(conn, addr, pkt, logf)
		return
	}

	// For other packets, manage retransmissions with window
	manageRetransmission(conn, addr, pkt, window, logf)
}

// sendAckPacket sends an ACK packet without waiting for acknowledgment
func sendAckPacket(conn *net.UDPConn, addr *net.UDPAddr, pkt ml.Packet, logf Logger) {
	fmt.Printf("📤 ACK sent, AckNum: %d\n", pkt.AckNum)
	if err := SendPacketUDP(conn, addr, pkt); err != nil {
		fmt.Println("❌ Error sending ACK:", err)
		logf("ERROR", "Failed to send ACK", err)
	}
}

// manageRetransmission handles sending and retransmitting packets until ACK is received
func manageRetransmission(conn *net.UDPConn, addr *net.UDPAddr, pkt ml.Packet, window *Window, logf Logger) {
	// Register packet in window
	ch := registerPacket(window, pkt.SeqNum)
	defer unregisterPacket(window, pkt.SeqNum)

	for retries := 0; retries <= maxRetries; retries++ {
		sendTime := time.Now()
		// Send the packet
		if err := SendPacketUDP(conn, addr, pkt); err != nil {
			fmt.Println("❌ Error sending packet:", err)
			logf("ERROR", "Failed to send packet", err)
			return
		}

		// get current RTO value
		rto := getRTO(window)

		select {
		case <-ch:
			// ACK received - update RTO and exit
			handleAckReceived(window, pkt.SeqNum, time.Since(sendTime))
			return

		case <-time.After(rto):
			// Timeout - prepare for retransmission
			if retries == maxRetries {
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
func handleAckReceived(window *Window, seqNum uint16, rtt time.Duration) {
	window.Mu.Lock()
	window.UpdateRTO(rtt)
	newRTO := window.RTO
	window.Mu.Unlock()

	fmt.Printf("✅ ACK to %d | RTT: %v | New RTO: %v\n", seqNum, rtt, newRTO)
}

// handleTimeout logs timeout and prepares for retransmission
func handleTimeout(seqNum uint16, retries int, rto time.Duration, logf func(level string, msg string, meta any)) {
	fmt.Printf("⏱️ Timeout waiting for ACK for SeqNum %d. Retransmitting (attempt %d)...\n", seqNum, retries+1)
	logf("WARN", "Timeout, retransmission", map[string]any{
		"seq":   seqNum,
		"retry": retries + 1,
		"rto":   rto,
	})
}

// handleMaxRetriesReached logs failure after max retries
func handleMaxRetriesReached(seqNum uint16, logf func(level string, msg string, meta any)) {
	fmt.Printf("❌ Failure to receive ACK for SeqNum %d after %d attempts. Aborting...\n", seqNum, maxRetries)
	logf("ERROR", "Failure after all attempts", map[string]any{
		"seq":        seqNum,
		"maxRetries": maxRetries,
	})
}

// SendAck sends an ACK packet for the given ackNum
func SendAck(conn *net.UDPConn, addr *net.UDPAddr, ackNum uint16, window *Window, roverId uint8, logf func(level string, msg string, meta any)) {
	ackPacket := ml.Packet{
		RoverId: roverId,
		MsgType: ml.MSG_ACK,
		SeqNum:  0,
		AckNum:  ackNum + 1,
		Payload: []byte{},
	}

	// Use PacketManager to handle sending the ACK packet
	PacketManager(conn, addr, ackPacket, window, logf)
}
