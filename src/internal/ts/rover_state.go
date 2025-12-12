package ts

import (
	"src/utils"
	"sync"
)

// RoverTSState holds the telemetry state of a rover.
type RoverTSState struct {
	ID              uint8            `json:"id"`              // Rover ID
	State           string           `json:"state"`           // e.g., "Idle", "Moving", "Error"
	Battery         uint8            `json:"battery"`         // Battery level percentage
	Speed           float32          `json:"speed"`           // Speed in m/s
	Position        utils.Coordinate `json:"position"`        // Current position
	UpdateFrequency uint             `json:"updateFrequency"` // Update frequency in seconds
	MissedTelemetry int              `json:"missedTelemetry"` // Consecutive telemetry failures count
	QueuedMissions  QueueInfo        `json:"queuedMissions"`  // Mission queue status
}

// QueueInfo holds information about the mission queue
type QueueInfo struct {
	Priority1Count uint8    `json:"priority1Count"` // Number of priority 1 missions
	Priority2Count uint8    `json:"priority2Count"` // Number of priority 2 missions
	Priority3Count uint8    `json:"priority3Count"` // Number of priority 3 missions
	Priority1IDs   []uint16 `json:"priority1Ids"`   // Mission IDs in priority 1 queue
	Priority2IDs   []uint16 `json:"priority2Ids"`   // Mission IDs in priority 2 queue
	Priority3IDs   []uint16 `json:"priority3Ids"`   // Mission IDs in priority 3 queue
}


// RoverManager manages multiple rovers' telemetry states.
type RoverManager struct {
	mu     sync.Mutex
	rovers map[uint8]*RoverTSState
}

// NewRoverManager creates a new RoverManager.
func NewRoverManager() *RoverManager {
	return &RoverManager{
		rovers: make(map[uint8]*RoverTSState),
	}
}

// AddRover adds a new rover to the manager.
func (rm *RoverManager) AddRover(rover *RoverTSState) {
	rm.mu.Lock()
	defer rm.mu.Unlock()
	if _, exists := rm.rovers[rover.ID]; !exists {
		rm.rovers[rover.ID] = rover
	}
}

// UpdateRover updates the telemetry state of an existing rover.
func (rm *RoverManager) UpdateRover(id uint8, state string, battery uint8, speed float32, position utils.Coordinate, missedTelemetry int, queueInfo QueueInfo) {
	rm.mu.Lock()
	defer rm.mu.Unlock()
	if rover, ok := rm.rovers[id]; ok {
		rover.State = state
		rover.Battery = battery
		rover.Speed = speed
		rover.Position = position
		rover.MissedTelemetry = missedTelemetry
		rover.QueuedMissions = queueInfo
	} else {
		// Create rover if it doesn't exist
		rm.rovers[id] = &RoverTSState{
			ID:              id,
			State:           state,
			Battery:         battery,
			Speed:           speed,
			Position:        position,
			MissedTelemetry: missedTelemetry,
			QueuedMissions:  queueInfo,
		}
	}
}

// RemoveRover removes a rover from the manager by its ID.
func (rm *RoverManager) RemoveRover(id uint8) {
	rm.mu.Lock()
	defer rm.mu.Unlock()
	delete(rm.rovers, id)
}

// GetRover retrieves a rover's telemetry state by its ID.
func (rm *RoverManager) GetRover(id uint8) *RoverTSState {
	rm.mu.Lock()
	defer rm.mu.Unlock()
	return rm.rovers[id]
}

// ListRovers returns a list of all registered rovers.
func (rm *RoverManager) ListRovers() []*RoverTSState {
	rm.mu.Lock()
	defer rm.mu.Unlock()
	list := make([]*RoverTSState, 0, len(rm.rovers))
	for _, rover := range rm.rovers {
		list = append(list, rover)
	}
	return list
}
