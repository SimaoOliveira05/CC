package main

import (
	"net"
	"src/config"
	"src/internal/ts"
	"src/utils"
	"sync"
)

// IDManager manages the assignment of unique IDs to rovers
type IDManager struct {
	nextID uint8      // Next available ID to assign
	mu     sync.Mutex // Mutex for concurrent access
}

// NewIDManager creates and initializes a new IDManager instance
func NewIDManager() *IDManager {
	return &IDManager{nextID: 1}
}

// GetNextID returns the next available unique ID and increments the counter
func (idm *IDManager) GetNextID() uint8 {
	idm.mu.Lock()
	defer idm.mu.Unlock()
	id := idm.nextID
	idm.nextID++
	return id
}

// idAssignmentServer starts a TCP server to handle ID assignment requests from rovers
func (ms *MotherShip) idAssignmentServer(port string) {

	idManager := NewIDManager()

	// Start TCP server
	listener, err := net.Listen("tcp", "0.0.0.0:"+port)
	if err != nil {
		ms.Logger.Errorf("IDHandler", "Erro ao iniciar servidor de IDs: %v", err)
		return
	}
	defer listener.Close()

	// Log server start
	ms.Logger.Infof("IDHandler", "Servidor de atribuição de IDs à escuta na porta %s", port)

	// Accept incoming connections
	for {
		conn, err := listener.Accept()
		if err != nil {
			ms.Logger.Errorf("IDHandler", "Erro ao aceitar conexão: %v", err)
			continue
		}
		go ms.handleIDRequest(conn, idManager)
	}
}

// handleIDRequest processes a single ID assignment request
func (ms *MotherShip) handleIDRequest(conn net.Conn, idManager *IDManager) {
	defer conn.Close()

	id := idManager.GetNextID()
	// Get update frequency from config (convert from Duration to seconds)
	updateFrequency := uint(config.DEFAULT_TELEMETRY_FREQ.Seconds())

	// Send assigned ID and update frequency to rover
	buf := []byte{id, byte(updateFrequency)}
	_, err := conn.Write(buf)
	if err != nil {
		ms.Logger.Errorf("IDHandler", "Error sending ID/updateFrequency: %v", err)
		return
	}

	// Log assignment and register rover in RoverInfo manager
	ms.Logger.Infof("IDHandler", "ID %d assigned to new rover (updateFrequency=%d)", id, updateFrequency)
	ms.RoverInfo.AddRover(&ts.RoverTSState{
		ID:              id,
		State:           "Unknown",
		Battery:         config.INITIAL_BATTERY,
		Speed:           0.0,
		Position:        utils.Coordinate{Latitude: 0, Longitude: 0},
		UpdateFrequency: updateFrequency,
		QueuedMissions: ts.QueueInfo{
			Priority1Count: 0,
			Priority2Count: 0,
			Priority3Count: 0,
			Priority1IDs:   []uint16{},
			Priority2IDs:   []uint16{},
			Priority3IDs:   []uint16{},
		},
	})
}
