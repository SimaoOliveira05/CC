package ml

import (
	"fmt"
	"src/utils"
	"sync"
	"time"
)

// MissionState represents the last updated state of a mission.
type MissionState struct {
	ID              uint16           `json:"id"`
	IDRover         uint8            `json:"idRover"`
	TaskType        uint8            `json:"taskType"`
	Duration        time.Duration    `json:"duration"`
	UpdateFrequency time.Duration    `json:"updateFrequency"`
	LastUpdate      time.Time        `json:"lastUpdate"`
	CreatedAt       time.Time        `json:"createdAt"`
	Priority        uint8            `json:"priority"`
	Report          []Report         `json:"reports"`
	State           string           `json:"state"` // e.g, "Pending", "Moving to", "In Progress", "Completed"
	Coordinate      utils.Coordinate `json:"coordinate"`
}

// MissionManager will manage all the active missions.
type MissionManager struct {
	ActiveMissions map[uint16]*MissionState
	mu             sync.RWMutex
}

func NewMissionManager() *MissionManager {
	return &MissionManager{
		ActiveMissions: make(map[uint16]*MissionState),
	}
}

// AddMission adds a new mission to the manager
func (mm *MissionManager) AddMission(mission *MissionState) {
	mm.mu.Lock()
	defer mm.mu.Unlock()
	mm.ActiveMissions[mission.ID] = mission
}

func UpdateMission(mm *MissionManager, report Report) {
	mm.mu.Lock()
	defer mm.mu.Unlock()

	mission := mm.ActiveMissions[report.GetMissionID()]
	if mission == nil {
		return
	}

	// Atualiza o estado genérico
	mission.Report = append(mission.Report, report)
	mission.LastUpdate = time.Now()

	// Atualizar estado baseado no report
	if report.IsLast() {
		mission.State = "Completed"
	} else {
		mission.State = "In Progress"
	}

	// Aqui podes adicionar lógica para atualizar outros campos conforme o tipo de report
	// Exemplo: mission.TaskType, mission.Priority, etc.
} // UpdateMissionState atualiza apenas o estado de uma missão
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
	defer mm.mu.Unlock() // When the function ends, unlock the mutex even in case of panic
	delete(mm.ActiveMissions, id)
}

// GetMission gets a mission by ID
func (mm *MissionManager) GetMission(id uint16) *MissionState {
	mm.mu.RLock()
	defer mm.mu.RUnlock() // When the function ends, unlock the mutex even in case of panic
	return mm.ActiveMissions[id]
}

// ListMissions returns a list of all missions
func (mm *MissionManager) ListMissions() []*MissionState {
	mm.mu.RLock()
	defer mm.mu.RUnlock()
	list := make([]*MissionState, 0, len(mm.ActiveMissions))
	for _, mission := range mm.ActiveMissions {
		list = append(list, mission)
	}
	return list
}

// PrintMissions imprime todas as missões e seus estados
func (mm *MissionManager) PrintMissions() {
	mm.mu.RLock()
	defer mm.mu.RUnlock()
	fmt.Println("===== Missões Ativas =====")
	for id, m := range mm.ActiveMissions {
		fmt.Printf("ID: %d | Rover: %d | TaskType: %d | Estado: %s | Duração: %v | Última atualização: %v\n",
			id, m.IDRover, m.TaskType, m.State, m.Duration, m.LastUpdate)
		if len(m.Report) == 0 {
			fmt.Println("  Nenhum relatório recebido")
		} else {
			fmt.Println("  Relatórios recebidos:")
			for i, rep := range m.Report {
				fmt.Printf("    [%d] %s\n", i+1, rep.String())
			}
		}
	}
	fmt.Println("==========================")
}
