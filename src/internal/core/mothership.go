package core

import (
	"encoding/json"
	"net"
	"net/http"
	"strconv"

	"src/internal/api"
	"src/internal/ml"
	"src/internal/ts"
	pl "src/utils/packetsLogic"
	"sync"

	"fmt"
	"io/ioutil"
	"os"

	"github.com/gorilla/mux"
)

type RoverState struct {
	Addr             *net.UDPAddr
	SeqNum           uint16
	ExpectedSeq      uint16
	Buffer           map[uint16]ml.Packet
	WindowLock       sync.Mutex
	Window           *pl.Window // Janela deslizante específica deste rover
	NumberOfMissions uint8
}

type MotherShip struct {
	Conn           *net.UDPConn
	Rovers         map[uint8]*RoverState // key: IP (ou ID do rover)
	MissionManager *ml.MissionManager
	MissionQueue   chan ml.MissionState
	Mu             sync.Mutex
	RoverInfo      *ts.RoverManager
	APIServer      *api.APIServer // ✅ Campo para o API Server
}

// Construtor
func NewMotherShip() *MotherShip {
	ms := &MotherShip{
		Rovers:         make(map[uint8]*RoverState),
		MissionManager: ml.NewMissionManager(),
		MissionQueue:   make(chan ml.MissionState, 100),
		Mu:             sync.Mutex{},
		RoverInfo:      ts.NewRoverManager(),
	}

	err := loadMissionsFromJSON("missions.json", ms.MissionQueue)
	if err != nil {
		fmt.Printf("erro ao carregar missões iniciais: %v\n", err)
		return nil
	}

	// Inicializa o APIServer
	ms.APIServer = api.NewAPIServer()

	// Configura os endpoints com os dados da mothership
	ms.setupAPIEndpoints()

	return ms
}

// loadMissionsFromJSON lê missões de um ficheiro JSON e coloca-as na missionQueue
func loadMissionsFromJSON(filename string, queue chan ml.MissionState) error {
	file, err := os.Open(filename)
	if err != nil {
		return fmt.Errorf("erro ao abrir ficheiro: %v", err)
	}
	defer file.Close()

	data, err := ioutil.ReadAll(file)
	if err != nil {
		return fmt.Errorf("erro ao ler ficheiro: %v", err)
	}

	var missions []ml.MissionState
	if err := json.Unmarshal(data, &missions); err != nil {
		return fmt.Errorf("erro ao fazer unmarshal do JSON: %v", err)
	}

	// Assign incremental IDs to missions
	for i := range missions {
		missions[i].ID = uint16(i + 1) // IDs start from 1
		queue <- missions[i]
	}

	fmt.Printf("📋 %d missões enfileiradas\n", len(missions))
	return nil
}

// setupAPIEndpoints configura todos os endpoints REST da API
func (ms *MotherShip) setupAPIEndpoints() {
	// Endpoint: Lista todos os rovers
	ms.APIServer.RegisterEndpoint("/api/rovers", "GET", func() interface{} {
		return ms.RoverInfo.ListRovers()
	})

	// Endpoint: Lista todas as missões
	ms.APIServer.RegisterEndpoint("/api/missions", "GET", func() interface{} {
		return ms.MissionManager.ListMissions()
	})

	// Endpoint: Estatísticas gerais
	ms.APIServer.RegisterEndpoint("/api/stats", "GET", func() interface{} {
		rovers := ms.RoverInfo.ListRovers()
		missions := ms.MissionManager.ListMissions()

		return map[string]interface{}{
			"total_rovers":         len(rovers),
			"total_missions":       len(missions),
			"active_rovers":        ms.countActiveRovers(rovers),
			"completed_missions":   ms.countCompletedMissions(missions),
			"pending_missions":     ms.countPendingMissions(missions),
			"missions_in_progress": ms.countInProgressMissions(missions),
		}
	})

	// Endpoint: Obter rover específico por ID (com parâmetros)
	ms.APIServer.RegisterEndpointWithParams("/api/rovers/{id}", "GET", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		vars := mux.Vars(r)
		id, err := strconv.Atoi(vars["id"])
		if err != nil {
			http.Error(w, "ID inválido", http.StatusBadRequest)
			return
		}

		rover := ms.RoverInfo.GetRover(uint8(id))
		if rover == nil {
			http.Error(w, "Rover não encontrado", http.StatusNotFound)
			return
		}
		json.NewEncoder(w).Encode(rover)
	})

	// Endpoint: Obter missão específica por ID (com parâmetros)
	ms.APIServer.RegisterEndpointWithParams("/api/missions/{id}", "GET", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		vars := mux.Vars(r)
		id, err := strconv.Atoi(vars["id"])
		if err != nil {
			http.Error(w, "ID inválido", http.StatusBadRequest)
			return
		}

		mission := ms.MissionManager.GetMission(uint16(id))
		if mission == nil {
			http.Error(w, "Missão não encontrada", http.StatusNotFound)
			return
		}
		json.NewEncoder(w).Encode(mission)
	})
}

// Funções auxiliares para estatísticas
func (ms *MotherShip) countActiveRovers(rovers []*ts.RoverTSState) int {
	count := 0
	for _, r := range rovers {
		if r.State == "Active" || r.State == "InMission" {
			count++
		}
	}
	return count
}

func (ms *MotherShip) countCompletedMissions(missions []*ml.MissionState) int {
	count := 0
	for _, m := range missions {
		if m.State == "Completed" {
			count++
		}
	}
	return count
}

func (ms *MotherShip) countPendingMissions(missions []*ml.MissionState) int {
	count := 0
	for _, m := range missions {
		if m.State == "Pending" {
			count++
		}
	}
	return count
}

func (ms *MotherShip) countInProgressMissions(missions []*ml.MissionState) int {
	count := 0
	for _, m := range missions {
		if m.State == "InProgress" {
			count++
		}
	}
	return count
}
