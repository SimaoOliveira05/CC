package ml

import (
	"sync"
	"time"
	"fmt"
)

// MissionState represents the last updated state of a mission.
type MissionState struct {
	ID              uint16
	IDRover         uint16
	TaskType        uint8
	Duration        time.Duration
	UpdateFrequency time.Duration
	LastUpdate      time.Time
	CreatedAt       time.Time
	Priority        uint8
	Report          Report
	State           string // e.g, "Pending", "In Progress", "Completed"
}

// MissionManager will manage all the active missions.
type MissionManager struct {
	ActiveMissions map[uint16]*MissionState
	mu       sync.RWMutex
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


	mission := mm.GetMission(report.GetMissionID())
    if mission == nil {
        return
    }

    // Atualiza o estado genérico
	mission.Report = report
    mission.LastUpdate = time.Now()
	if report.IsLast() {
		mission.State = "Completed"
	} else {
		mission.State = "In Progress"
	}

    // Aqui podes adicionar lógica para atualizar outros campos conforme o tipo de report
    // Exemplo: mission.TaskType, mission.Priority, etc.
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

// PrintMissions imprime todas as missões e seus estados
func (mm *MissionManager) PrintMissions() {
	mm.mu.RLock()
	defer mm.mu.RUnlock()
	fmt.Println("===== Missões Ativas =====")
	for id, m := range mm.ActiveMissions {
		if (m.Report == nil) {
			fmt.Printf("ID: %d | Rover: %d | TaskType: %d | Estado: %s | Duração: %v | Última atualização: %v | Detalhes: Nenhum relatório recebido\n",
				id, m.IDRover, m.TaskType, m.State, m.Duration, m.LastUpdate)
		} else {
			fmt.Printf("ID: %d | Rover: %d | TaskType: %d | Estado: %s | Duração: %v | Última atualização: %v | Detalhes: %s\n",
				id, m.IDRover, m.TaskType, m.State, m.Duration, m.LastUpdate, m.Report.String())
		}
	}
	fmt.Println("==========================")
}
