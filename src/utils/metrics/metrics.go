package metrics

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"sync/atomic"
	"time"
)

// MLMetrics stores protocol metrics for the ML layer
type MLMetrics struct {
	mu        sync.RWMutex
	enabled   bool
	startTime time.Time

	// Packet counts
	PacketsSent     uint64
	PacketsReceived uint64
	AcksSent        uint64
	AcksReceived    uint64

	// Error metrics
	ChecksumsFailed    uint64
	Retransmissions    uint64
	PacketsLost        uint64 // Packets that exceeded max retries
	DuplicatesReceived uint64

	// Out-of-order metrics
	OutOfOrderReceived uint64
	BufferedPackets    uint64

	// Timing metrics
	TotalRTT   time.Duration
	RTTSamples uint64
	MinRTT     time.Duration
	MaxRTT     time.Duration

	// Throughput tracking
	BytesSent     uint64
	BytesReceived uint64

	// Per-packet type counters
	PacketTypesSent     map[string]uint64
	PacketTypesReceived map[string]uint64
}

// NewMetricsManager creates a new metrics manager
func NewMetricsManager(enabled bool) *MLMetrics {
	return &MLMetrics{
		enabled:             enabled,
		startTime:           time.Now(),
		MinRTT:              time.Hour, // Start high for proper min tracking
		PacketTypesSent:     make(map[string]uint64),
		PacketTypesReceived: make(map[string]uint64),
	}
}

// IsEnabled returns whether metrics collection is enabled
func (m *MLMetrics) IsEnabled() bool {
	return m.enabled
}

// Enable turns on metrics collection
func (m *MLMetrics) Enable() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.enabled = true
	m.startTime = time.Now()
}

// Disable turns off metrics collection
func (m *MLMetrics) Disable() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.enabled = false
}

// --- Increment methods (thread-safe) ---

// RecordPacketSent records a sent packet
func (m *MLMetrics) RecordPacketSent(packetType string, size int) {
	if !m.enabled {
		return
	}
	atomic.AddUint64(&m.PacketsSent, 1)
	atomic.AddUint64(&m.BytesSent, uint64(size))

	m.mu.Lock()
	m.PacketTypesSent[packetType]++
	m.mu.Unlock()
}

// RecordPacketReceived records a received packet
func (m *MLMetrics) RecordPacketReceived(packetType string, size int) {
	if !m.enabled {
		return
	}
	atomic.AddUint64(&m.PacketsReceived, 1)
	atomic.AddUint64(&m.BytesReceived, uint64(size))

	m.mu.Lock()
	m.PacketTypesReceived[packetType]++
	m.mu.Unlock()
}

// RecordAckSent records a sent ACK
func (m *MLMetrics) RecordAckSent() {
	if !m.enabled {
		return
	}
	atomic.AddUint64(&m.AcksSent, 1)
}

// RecordAckReceived records a received ACK
func (m *MLMetrics) RecordAckReceived() {
	if !m.enabled {
		return
	}
	atomic.AddUint64(&m.AcksReceived, 1)
}

// RecordChecksumFailed records a checksum verification failure
func (m *MLMetrics) RecordChecksumFailed() {
	if !m.enabled {
		return
	}
	atomic.AddUint64(&m.ChecksumsFailed, 1)
}

// RecordRetransmission records a packet retransmission
func (m *MLMetrics) RecordRetransmission() {
	if !m.enabled {
		return
	}
	atomic.AddUint64(&m.Retransmissions, 1)
}

// RecordPacketLost records a packet that was never acknowledged (max retries reached)
func (m *MLMetrics) RecordPacketLost() {
	if !m.enabled {
		return
	}
	atomic.AddUint64(&m.PacketsLost, 1)
}

// RecordDuplicateReceived records a duplicate packet received
func (m *MLMetrics) RecordDuplicateReceived() {
	if !m.enabled {
		return
	}
	atomic.AddUint64(&m.DuplicatesReceived, 1)
}

// RecordOutOfOrder records an out-of-order packet
func (m *MLMetrics) RecordOutOfOrder() {
	if !m.enabled {
		return
	}
	atomic.AddUint64(&m.OutOfOrderReceived, 1)
}

// RecordBufferedPacket records a packet being buffered
func (m *MLMetrics) RecordBufferedPacket() {
	if !m.enabled {
		return
	}
	atomic.AddUint64(&m.BufferedPackets, 1)
}

// RecordRTT records a Round-Trip Time sample
func (m *MLMetrics) RecordRTT(rtt time.Duration) {
	if !m.enabled {
		return
	}
	m.mu.Lock()
	defer m.mu.Unlock()

	m.TotalRTT += rtt
	m.RTTSamples++

	if rtt < m.MinRTT {
		m.MinRTT = rtt
	}
	if rtt > m.MaxRTT {
		m.MaxRTT = rtt
	}
}

// --- Computed metrics ---

// GetAverageRTT returns the average RTT
func (m *MLMetrics) GetAverageRTT() time.Duration {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.RTTSamples == 0 {
		return 0
	}
	return m.TotalRTT / time.Duration(m.RTTSamples)
}

// GetPacketLossRate returns the packet loss rate as a percentage
func (m *MLMetrics) GetPacketLossRate() float64 {
	sent := atomic.LoadUint64(&m.PacketsSent)
	lost := atomic.LoadUint64(&m.PacketsLost)

	if sent == 0 {
		return 0
	}
	return float64(lost) / float64(sent) * 100
}

// GetRetransmissionRate returns the retransmission rate as a percentage
func (m *MLMetrics) GetRetransmissionRate() float64 {
	sent := atomic.LoadUint64(&m.PacketsSent)
	retrans := atomic.LoadUint64(&m.Retransmissions)

	if sent == 0 {
		return 0
	}
	return float64(retrans) / float64(sent) * 100
}

// GetDuplicateRate returns the duplicate packet rate as a percentage
func (m *MLMetrics) GetDuplicateRate() float64 {
	received := atomic.LoadUint64(&m.PacketsReceived)
	duplicates := atomic.LoadUint64(&m.DuplicatesReceived)

	if received == 0 {
		return 0
	}
	return float64(duplicates) / float64(received+duplicates) * 100
}

// GetThroughput returns the throughput in bytes per second
func (m *MLMetrics) GetThroughput() (sentBps, receivedBps float64) {
	elapsed := time.Since(m.startTime).Seconds()
	if elapsed == 0 {
		return 0, 0
	}

	sentBps = float64(atomic.LoadUint64(&m.BytesSent)) / elapsed
	receivedBps = float64(atomic.LoadUint64(&m.BytesReceived)) / elapsed
	return
}

// GetUptime returns the duration since metrics collection started
func (m *MLMetrics) GetUptime() time.Duration {
	return time.Since(m.startTime)
}

// --- Export methods ---

// Summary returns a summary of all metrics
type MetricsSummary struct {
	Uptime              string            `json:"uptime"`
	PacketsSent         uint64            `json:"packets_sent"`
	PacketsReceived     uint64            `json:"packets_received"`
	AcksSent            uint64            `json:"acks_sent"`
	AcksReceived        uint64            `json:"acks_received"`
	ChecksumsFailed     uint64            `json:"checksums_failed"`
	Retransmissions     uint64            `json:"retransmissions"`
	PacketsLost         uint64            `json:"packets_lost"`
	DuplicatesReceived  uint64            `json:"duplicates_received"`
	OutOfOrderReceived  uint64            `json:"out_of_order_received"`
	BytesSent           uint64            `json:"bytes_sent"`
	BytesReceived       uint64            `json:"bytes_received"`
	AvgRTT              string            `json:"avg_rtt"`
	MinRTT              string            `json:"min_rtt"`
	MaxRTT              string            `json:"max_rtt"`
	PacketLossRate      float64           `json:"packet_loss_rate_percent"`
	RetransmissionRate  float64           `json:"retransmission_rate_percent"`
	DuplicateRate       float64           `json:"duplicate_rate_percent"`
	ThroughputSentBps   float64           `json:"throughput_sent_bps"`
	ThroughputRecvBps   float64           `json:"throughput_recv_bps"`
	PacketTypesSent     map[string]uint64 `json:"packet_types_sent"`
	PacketTypesReceived map[string]uint64 `json:"packet_types_received"`
}

// GetSummary returns a complete summary of metrics
func (m *MLMetrics) GetSummary() MetricsSummary {
	m.mu.RLock()
	defer m.mu.RUnlock()

	sentBps, recvBps := m.GetThroughput()

	minRTT := m.MinRTT
	if minRTT == time.Hour {
		minRTT = 0
	}

	// Copy maps to avoid race conditions
	typesSent := make(map[string]uint64)
	for k, v := range m.PacketTypesSent {
		typesSent[k] = v
	}
	typesReceived := make(map[string]uint64)
	for k, v := range m.PacketTypesReceived {
		typesReceived[k] = v
	}

	return MetricsSummary{
		Uptime:              m.GetUptime().Round(time.Second).String(),
		PacketsSent:         atomic.LoadUint64(&m.PacketsSent),
		PacketsReceived:     atomic.LoadUint64(&m.PacketsReceived),
		AcksSent:            atomic.LoadUint64(&m.AcksSent),
		AcksReceived:        atomic.LoadUint64(&m.AcksReceived),
		ChecksumsFailed:     atomic.LoadUint64(&m.ChecksumsFailed),
		Retransmissions:     atomic.LoadUint64(&m.Retransmissions),
		PacketsLost:         atomic.LoadUint64(&m.PacketsLost),
		DuplicatesReceived:  atomic.LoadUint64(&m.DuplicatesReceived),
		OutOfOrderReceived:  atomic.LoadUint64(&m.OutOfOrderReceived),
		BytesSent:           atomic.LoadUint64(&m.BytesSent),
		BytesReceived:       atomic.LoadUint64(&m.BytesReceived),
		AvgRTT:              m.GetAverageRTT().Round(time.Microsecond).String(),
		MinRTT:              minRTT.Round(time.Microsecond).String(),
		MaxRTT:              m.MaxRTT.Round(time.Microsecond).String(),
		PacketLossRate:      m.GetPacketLossRate(),
		RetransmissionRate:  m.GetRetransmissionRate(),
		DuplicateRate:       m.GetDuplicateRate(),
		ThroughputSentBps:   sentBps,
		ThroughputRecvBps:   recvBps,
		PacketTypesSent:     typesSent,
		PacketTypesReceived: typesReceived,
	}
}

// ExportToJSON exports metrics to a JSON file
func (m *MLMetrics) ExportToJSON(filename string) error {
	summary := m.GetSummary()

	data, err := json.MarshalIndent(summary, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal metrics: %w", err)
	}

	err = os.WriteFile(filename, data, 0644)
	if err != nil {
		return fmt.Errorf("failed to write metrics file: %w", err)
	}

	return nil
}

// Reset resets all metrics
func (m *MLMetrics) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()

	atomic.StoreUint64(&m.PacketsSent, 0)
	atomic.StoreUint64(&m.PacketsReceived, 0)
	atomic.StoreUint64(&m.AcksSent, 0)
	atomic.StoreUint64(&m.AcksReceived, 0)
	atomic.StoreUint64(&m.ChecksumsFailed, 0)
	atomic.StoreUint64(&m.Retransmissions, 0)
	atomic.StoreUint64(&m.PacketsLost, 0)
	atomic.StoreUint64(&m.DuplicatesReceived, 0)
	atomic.StoreUint64(&m.OutOfOrderReceived, 0)
	atomic.StoreUint64(&m.BufferedPackets, 0)
	atomic.StoreUint64(&m.BytesSent, 0)
	atomic.StoreUint64(&m.BytesReceived, 0)

	m.TotalRTT = 0
	m.RTTSamples = 0
	m.MinRTT = time.Hour
	m.MaxRTT = 0
	m.startTime = time.Now()
	m.PacketTypesSent = make(map[string]uint64)
	m.PacketTypesReceived = make(map[string]uint64)
}

// Global metrics instance (can be nil if not in test mode)
var GlobalMetrics *MLMetrics

// InitGlobalMetrics initializes the global metrics instance
func InitGlobalMetrics(enabled bool) {
	GlobalMetrics = NewMetricsManager(enabled)
}

// GetGlobalMetrics returns the global metrics instance
func GetGlobalMetrics() *MLMetrics {
	return GlobalMetrics
}
