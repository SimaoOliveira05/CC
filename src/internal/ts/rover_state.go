package ts

import (
	"fmt"
	"src/utils"
	"sync"
)

type RoverTSState struct {
	ID              uint8            `json:"id"`
	State           string           `json:"state"`
	Battery         uint8            `json:"battery"`
	Speed           float32          `json:"speed"`
	Position        utils.Coordinate `json:"position"`
	UpdateFrequency uint
	MissedTelemetry int // contador de falhas consecutivas
}

func (r *RoverTSState) String() string {
	return fmt.Sprintf("Rover %d | Estado: %s | Bateria: %d%% | Velocidade: %.2f m/s", r.ID, r.State, r.Battery, r.Speed)
}

type RoverManager struct {
	mu     sync.Mutex
	rovers map[uint8]*RoverTSState
}

func NewRoverManager() *RoverManager {
	return &RoverManager{
		rovers: make(map[uint8]*RoverTSState),
	}
}

func (rm *RoverManager) AddRover(rover *RoverTSState) {
	rm.mu.Lock()
	defer rm.mu.Unlock()
	if _, exists := rm.rovers[rover.ID]; !exists {
		rm.rovers[rover.ID] = rover
	}
}

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

func (rm *RoverManager) RemoveRover(id uint8) {
	rm.mu.Lock()
	defer rm.mu.Unlock()
	delete(rm.rovers, id)
}

func (rm *RoverManager) GetRover(id uint8) *RoverTSState {
	rm.mu.Lock()
	defer rm.mu.Unlock()
	return rm.rovers[id]
}

func (rm *RoverManager) ListRovers() []*RoverTSState {
	rm.mu.Lock()
	defer rm.mu.Unlock()
	list := make([]*RoverTSState, 0, len(rm.rovers))
	for _, rover := range rm.rovers {
		list = append(list, rover)
	}
	return list
}

func (rm *RoverManager) String() string {
	rm.mu.Lock()
	defer rm.mu.Unlock()
	if len(rm.rovers) == 0 {
		return "Nenhum rover registado."
	}
	result := "--- Estado dos Rovers ---\n"
	for _, rover := range rm.rovers {
		result += rover.String() + "\n"
	}
	return result
}
