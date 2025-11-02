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

// UpdateMission updates a mission stat and last update time
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
 