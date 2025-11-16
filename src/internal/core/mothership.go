package core

import (
    "net"
    "src/internal/ml"
	"src/internal/ts"
	pl "src/utils/packetsLogic"
    "sync"
)

type RoverState struct {
	Addr        *net.UDPAddr
	SeqNum      uint16
	ExpectedSeq uint16
	Buffer      map[uint16]ml.Packet
	WindowLock  sync.Mutex
	Window      *pl.Window // Janela deslizante específica deste rover
}

type MotherShip struct {
	Conn           *net.UDPConn
	Rovers         map[uint8]*RoverState // key: IP (ou ID do rover)
	MissionManager *ml.MissionManager
	MissionQueue   chan ml.MissionState
	Mu             sync.Mutex
	RoverInfo      *ts.RoverManager
}

// Construtor
func NewMotherShip(conn *net.UDPConn) *MotherShip {
    return &MotherShip{
        Conn:           conn,
		Rovers:         make(map[uint8]*RoverState),
        MissionManager: ml.NewMissionManager(),
        MissionQueue:   make(chan ml.MissionState, 100),
		Mu:             sync.Mutex{},
		RoverInfo:      ts.NewRoverManager(),
    }
}
