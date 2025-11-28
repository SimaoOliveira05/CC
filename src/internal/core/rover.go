package core

import (
	"net"
	"src/internal/devices"
	"src/internal/ml"
	"src/utils"
	pl "src/utils/packetsLogic"
	"sync"
)

// Base compartilhada
type RoverBase struct {
	ID          uint8
	Coordinates utils.Coordinate // Mover de utils para core
	mu          sync.RWMutex
}

type RoverMLState struct {
	// Gerenciar missões
	ActiveMissions      uint8      // Número de missões ativas
	Cond                *sync.Cond // Condição para sincronização de missões
	CondMu              sync.Mutex // Mutex para a condição
	Waiting             bool       // Indica se o rover está esperando por uma missão
	MissionReceivedChan chan bool
	SeqNum              uint32 // Número de sequência para envio

	// Gerenciamento de pacotes e seqnums
	ExpectedSeq uint16
	Buffer      map[uint16]ml.Packet
	BufferMu    sync.Mutex

	// Janela deslizante para controle de ACKS e retransmissões
	Window *pl.Window // Janela deslizante específica deste rover
}

type RoverMLConnection struct {
	Conn *net.UDPConn // Conexão UDP com a nave-mãe
	Addr *net.UDPAddr // Endereço da nave-mãe
}

type RoverInfo struct {
	State   string
	Battery uint8
	Speed   float32
}

type Devices struct {
	GPS              devices.GPS
	Thermometer      devices.Thermometer
	Battery          devices.Battery
	Camera           devices.Camera
	ChemicalAnalyzer devices.ChemicalAnalyzer
}
