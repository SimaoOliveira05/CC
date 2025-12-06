package core

import (
	"encoding/json"
	"fmt"
	"net"
	"os"
	"src/config"
	"src/internal/api"
	el "src/internal/eventLogger"
	"src/internal/ml"
	"src/internal/ts"
	pl "src/utils/packetsLogic"
	"sync"
)

// RoverState maintain the state of each rover connected to the mothership
type RoverState struct {
	Addr             *net.UDPAddr         // Address of the rover
	SeqNum           uint16               // Sequence number for sending packets to the rover (ML)
	ExpectedSeq      uint16               // Expected sequence number for receiving packets from the rover (ML)
	Buffer           map[uint16]ml.Packet // Buffer for out-of-order packets (ML)
	WindowLock       sync.Mutex           // Mutex for sliding window operations
	Window           *pl.Window           // Sliding window specific to this rover
	NumberOfMissions uint8                // Number of missions rover is currently handling
}

// MotherShip represents the central control system managing multiple rovers
type MotherShip struct {
	Conn           *net.UDPConn          // UDP connection for communication with rovers
	Rovers         map[uint8]*RoverState // key: rover ID
	MissionManager *ml.MissionManager    // Manages missions
	MissionQueue   chan ml.MissionState  // Queue of missions to be assigned
	Mu             sync.Mutex            // Mutex for concurrent access to Rovers map
	RoverInfo      *ts.RoverManager      // Manages rover telemetry states
	EventLogger    *el.EventLogger       // Event logger for the mothership
	APIServer      *api.APIServer        // API server for handling REST endpoints
}

// NewMotherShip creates and initializes a new MotherShip instance
func NewMotherShip() *MotherShip {
	ms := &MotherShip{
		Rovers:         make(map[uint8]*RoverState),
		MissionManager: ml.NewMissionManager(),
		MissionQueue:   make(chan ml.MissionState, config.MISSION_QUEUE_SIZE),
		Mu:             sync.Mutex{},
		RoverInfo:      ts.NewRoverManager(),
		APIServer:      api.NewAPIServer(),
	}

	// Load initial missions from JSON file
	err := loadMissionsFromJSON("missions.json", ms.MissionQueue)
	if err != nil {
		fmt.Printf("erro ao carregar missões iniciais: %v\n", err)
		return nil
	}

	// Initialize event logger
	ms.EventLogger = el.NewEventLogger(config.EVENT_LOGGER_SIZE, ms.APIServer)

	// Setup API endpoints with mothership data
	ms.setupAPIEndpoints()

	return ms
}

// loadMissionsFromJSON read the missions from a JSON file and enqueue them
func loadMissionsFromJSON(filename string, queue chan ml.MissionState) error {
	file, err := os.Open(filename)
	if err != nil {
		return fmt.Errorf("error opening file: %v", err)
	}
	defer file.Close()

	data, err := os.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("error reading file: %v", err)
	}

	var missions []ml.MissionState
	if err := json.Unmarshal(data, &missions); err != nil {
		return fmt.Errorf("error unmarshaling JSON: %v", err)
	}

	// Assign incremental IDs to missions
	for i := range missions {
		missions[i].ID = uint16(i + 1) // IDs start from 1

		queue <- missions[i]
	}

	fmt.Printf("📋 %d missions enqueued\n", len(missions))
	return nil
}

// NewRoverState cria e inicializa um novo estado de rover para a MotherShip
func NewRoverState(addr *net.UDPAddr, seqNum uint16) *RoverState {
	return &RoverState{
		Addr:             addr,
		SeqNum:           seqNum,
		ExpectedSeq:      seqNum,
		Buffer:           make(map[uint16]ml.Packet),
		WindowLock:       sync.Mutex{},
		Window:           pl.NewWindow(),
		NumberOfMissions: 0,
	}
}
