package main

import (
	"fmt"
	"net"
	"src/internal/ts"
	"src/utils"
	"sync"
)

// IDManager manages the assignment of unique IDs to rovers
type IDManager struct {
	nextID uint8		// Next available ID to assign
	mu     sync.Mutex	// Mutex for concurrent access
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
		ms.EventLogger.Log("ERROR", "IDHandler", fmt.Sprintf("Erro ao iniciar servidor de IDs: %v", err), nil)
		return
	}
	defer listener.Close()

	// Log server start
	ms.EventLogger.Log("INFO", "IDHandler", fmt.Sprintf("Servidor de atribuição de IDs à escuta na porta %s", port), nil)

	// Accept incoming connections
	for {
		conn, err := listener.Accept()
		if err != nil {
			ms.EventLogger.Log("ERROR", "IDHandler", fmt.Sprintf("Erro ao aceitar conexão: %v", err), nil)
			continue
		}
		go ms.handleIDRequest(conn, idManager)
	}
}

// handleIDRequest processes a single ID assignment request
func (ms *MotherShip) handleIDRequest(conn net.Conn, idManager *IDManager) {
    defer conn.Close()

    id := idManager.GetNextID()
    var updateFrequency uint = 2 // Default update frequency in seconds

	// Send assigned ID and update frequency to rover
    buf := []byte{id, byte(updateFrequency)}
    _, err := conn.Write(buf)
    if err != nil {
        ms.EventLogger.Log("ERROR", "IDHandler", fmt.Sprintf("Error sending ID/updateFrequency: %v", err), nil)
        return
    }

	// Log assignment and register rover in RoverInfo manager
    ms.EventLogger.Log("INFO", "IDHandler", fmt.Sprintf("ID %d assigned to new rover (updateFrequency=%d)", id, updateFrequency), nil)
    ms.RoverInfo.AddRover(&ts.RoverTSState{
        ID:       id,
        State:    "Unknown",
        Battery:  100,
        Speed:    0.0,
        Position: utils.Coordinate{Latitude: 0, Longitude: 0},
        UpdateFrequency: updateFrequency,
    })
}
