package ts

import (
	"fmt"
	"src/utils"
	"sync"
)

// RoverTSState holds the telemetry state of a rover.
type RoverTSState struct {
	ID              uint8            `json:"id"` // Rover ID
	State           string           `json:"state"` // e.g., "Idle", "Moving", "Error"
	Battery         uint8            `json:"battery"` // Battery level percentage
	Speed           float32          `json:"speed"` // Speed in m/s
	Position        utils.Coordinate `json:"position"` // Current position
	UpdateFrequency uint             `json:"updateFrequency"` // Update frequency in seconds
	MissedTelemetry int              `json:"missedTelemetry"` // Consecutive telemetry failures count
}

// String returns a human-readable representation of the RoverTSState.
func (r *RoverTSState) String() string {
	return fmt.Sprintf("Rover %d | State: %s | Battery: %d%% | Speed: %.2f m/s", r.ID, r.State, r.Battery, r.Speed)
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
func (rm *RoverManager) UpdateRover(id uint8, state string, battery uint8, speed float32, position utils.Coordinate, missedTelemetry int) {
	rm.mu.Lock()
	defer rm.mu.Unlock()
	if rover, ok := rm.rovers[id]; ok {
		rover.State = state
		rover.Battery = battery
		rover.Speed = speed
		rover.Position = position
		rover.MissedTelemetry = missedTelemetry
	} else {
		// Create rover if it doesn't exist
		rm.rovers[id] = &RoverTSState{
			ID:              id,
			State:           state,
			Battery:         battery,
			Speed:           speed,
			Position:        position,
			MissedTelemetry: missedTelemetry,
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

// String returns a human-readable representation of all rovers managed.
func (rm *RoverManager) String() string {
	rm.mu.Lock()
	defer rm.mu.Unlock()
	if len(rm.rovers) == 0 {
		return "No rovers registered."
	}
	result := "--- Rovers State ---\n"
	for _, rover := range rm.rovers {
		result += rover.String() + "\n"
	}
	return result
}
