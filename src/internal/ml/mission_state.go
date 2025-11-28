package ml

import (
	"fmt"
	"sort"
	"src/utils"
	"sync"
	"time"
)

// MissionState represents the last updated state of a mission.
type MissionState struct {
	ID              uint16           `json:"id"`              // Unique mission ID
	IDRover         uint8            `json:"idRover"`         // ID of the rover assigned to the mission
	TaskType        uint8            `json:"taskType"`        // e.g., 1 = MoveTo, 2 = SampleCollection, etc.
	Duration        time.Duration    `json:"duration"`        // Duration since mission start
	UpdateFrequency time.Duration    `json:"updateFrequency"` // Frequency of updates
	LastUpdate      time.Time        `json:"lastUpdate"`      // Time of the last update
	CreatedAt       time.Time        `json:"createdAt"`       // Time when the mission was created
	Priority        uint8            `json:"priority"`        // Priority level of the mission
	Report          []Report         `json:"reports"`         // Reports related to the mission
	State           string           `json:"state"`           // e.g, "Pending", "Moving to", "In Progress", "Completed"
	Coordinate      utils.Coordinate `json:"coordinate"`      // Target coordinate for the mission
}

// MissionManager will manage all the active missions.
type MissionManager struct {
	ActiveMissions map[uint16]*MissionState
	mu             sync.RWMutex
}

// NewMissionManager creates a new MissionManager.
func NewMissionManager() *MissionManager {
	return &MissionManager{
		ActiveMissions: make(map[uint16]*MissionState),
	}
}

// AddMission adds a new mission to the manager.
func (mm *MissionManager) AddMission(mission *MissionState) {
	mm.mu.Lock()
	defer mm.mu.Unlock()
	mm.ActiveMissions[mission.ID] = mission
}

// UpdateMission updates the mission state based on a report.
func UpdateMission(mm *MissionManager, report Report) {
	mm.mu.Lock()
	defer mm.mu.Unlock()

	mission := mm.ActiveMissions[report.GetMissionID()]
	if mission == nil {
		return
	}

	// Actualize generic state
	mission.Report = append(mission.Report, report)
	mission.LastUpdate = time.Now()

	// Actualize state based on the report
	if report.IsLast() {
		mission.State = "Completed"
	} else {
		mission.State = "In Progress"
	}
}

// UpdateMissionState actualize the state of a mission.
func (mm *MissionManager) UpdateMissionState(missionID uint16, newState string) {
	mm.mu.Lock()
	defer mm.mu.Unlock()

	mission := mm.ActiveMissions[missionID]
	if mission == nil {
		return
	}

	mission.State = newState
	mission.LastUpdate = time.Now()
}

// DeleteMission removes a mission from the manager
func (mm *MissionManager) DeleteMission(id uint16) {
	mm.mu.Lock()
	defer mm.mu.Unlock()
	delete(mm.ActiveMissions, id)
}

// GetMission gets a mission by ID
func (mm *MissionManager) GetMission(id uint16) *MissionState {
	mm.mu.RLock()
	defer mm.mu.RUnlock()
	return mm.ActiveMissions[id]
}

// ListMissions returns a list of all missions.
func (mm *MissionManager) ListMissions() []*MissionState {
	mm.mu.RLock()
	defer mm.mu.RUnlock()
	list := make([]*MissionState, 0, len(mm.ActiveMissions))
	for _, mission := range mm.ActiveMissions {
		list = append(list, mission)
	}
	return list
}

// PrintMissions prints all missions and their states.
func (mm *MissionManager) PrintMissions() {
	mm.mu.RLock()
	defer mm.mu.RUnlock()
	fmt.Println("===== Active Missions =====")
	for id, m := range mm.ActiveMissions {
		fmt.Printf("ID: %d | Rover: %d | TaskType: %d | State: %s | Duration: %v | Last Update: %v\n",
			id, m.IDRover, m.TaskType, m.State, m.Duration, m.LastUpdate)
		if len(m.Report) == 0 {
			fmt.Println("  No reports received")
		} else {
			fmt.Println("  Reports received:")
			for i, rep := range m.Report {
				fmt.Printf("    [%d] %s\n", i+1, rep.String())
			}
		}
	}
	fmt.Println("==========================")
}

// AssembleImage concatenates all image chunks (by ChunkID order) for the mission into a single byte slice.
func (m *MissionState) AssembleImage() []byte {
	// Collect chunks by id
	chunks := make(map[uint16][]byte)
	var ids []int
	for _, rep := range m.Report {
		if rep.Header.TaskType == TASK_IMAGE_CAPTURE {
			var img ImageReportData
			img.DecodePayload(rep.Payload)
			// copy data to avoid referencing underlying slices
			dataCopy := make([]byte, len(img.Data))
			copy(dataCopy, img.Data)
			chunks[img.ChunkID] = dataCopy
			ids = append(ids, int(img.ChunkID))
		}
	}
	if len(ids) == 0 {
		return nil
	}
	sort.Ints(ids)
	// Concatenate in order
	var result []byte
	for _, id := range ids {
		c := chunks[uint16(id)]
		result = append(result, c...)
	}
	return result
}
