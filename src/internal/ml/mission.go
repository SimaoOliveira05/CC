package ml

import (
	"sync"
	"time"
)

// MissionState represents the last updated state of a mission.
type MissionState struct {
	ID              uint32
	IDRover         uint16
	TaskType        uint8
	Duration        time.Duration
	UpdateFrequency time.Duration
	LastUpdate      time.Time
	CreatedAt       time.Time
	Priority        uint8
	State           string // e.g, "Pending", "In Progress", "Completed"
}

// MissionManager will manage all the active missions.
type MissionManager struct {
	missions map[uint32]*MissionState
	mu       sync.RWMutex
}

func NewMissionManager() *MissionManager {
	return &MissionManager{
		missions: make(map[uint32]*MissionState),
	}
}

// AddMission adds a new mission to the manager
func (mm *MissionManager) AddMission(mission *MissionState) {
	mm.mu.Lock()
	defer mm.mu.Unlock()
	mm.missions[mission.ID] = mission
}

// UpdateMission updates a mission state and last update time
func (mm *MissionManager) UpdateMission(id uint32, state string) {
	mm.mu.Lock()
	defer mm.mu.Unlock()
	if m, exists := mm.missions[id]; exists {
		m.State = state
		m.LastUpdate = time.Now()
	}
}

// DeleteMission removes a mission from the manager
func (mm *MissionManager) DeleteMission(id uint32) {
	mm.mu.Lock()
	defer mm.mu.Unlock() // When the function ends, unlock the mutex even in case of panic
	delete(mm.missions, id)
}

// GetMission gets a mission by ID
func (mm *MissionManager) GetMission(id uint32) *MissionState {
	mm.mu.RLock()
	defer mm.mu.RUnlock() // When the function ends, unlock the mutex even in case of panic
	return mm.missions[id]
}

// PopulateWithDemoMissions seeds the MissionManager with a few demo missions
// using different task types, durations and priorities. It returns the created
// mission IDs so callers can reference them in tests.
func PopulateWithDemoMissions(mm *MissionManager) []uint32 {
	now := time.Now()
	ids := make([]uint32, 0, 5)

	// small helper to add one mission
	add := func(offsetSec int, roverID uint16, taskType uint8, durSec, freqSec int, priority uint8, state string) {
		id := uint32(now.Unix()) + uint32(offsetSec)
		m := &MissionState{
			ID:              id,
			IDRover:         roverID,
			TaskType:        taskType,
			Duration:        time.Duration(durSec) * time.Second,
			UpdateFrequency: time.Duration(freqSec) * time.Second,
			LastUpdate:      now,
			CreatedAt:       now,
			Priority:        priority,
			State:           state,
		}
		mm.AddMission(m)
		ids = append(ids, id)
	}

	// Add a few varied demo missions (task types from data.go)
	add(1, 101, TASK_IMAGE_CAPTURE, 180, 10, 2, "Pending")
	add(2, 101, TASK_SAMPLE_COLLECTION, 240, 15, 1, "Pending")
	//add(3, 101, TASK_ENV_ANALYSIS, 300, 20, 3, "In Progress")
	add(4, 101, TASK_REPAIR_RESCUE, 120, 5, 0, "Pending")
	add(5, 101, TASK_TOPO_MAPPING, 600, 30, 2, "Pending")
	add(6, 101, TASK_INSTALLATION, 420, 20, 1, "Pending")

	return ids
}
