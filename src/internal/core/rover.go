package core


import (
	"src/internal/devices"
	"src/internal/ml"
	"src/internal/ts"
	"src/utils"
	pl "src/utils/packetsLogic"
	"sync"
	"net"
	"fmt"
	"time"
)

// Base compartilhada
type RoverBase struct {
	ID          uint8
	Coordinates utils.Coordinate // Mover de utils para core
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


type RoverSystem struct {
	*RoverBase
	ML         *RoverMLState
	TS         *ts.RoverTSState
	MLConn     *RoverMLConnection
	Devices    *Devices
	CurrentPos utils.Coordinate
}

func requestID(mothershipAddr string) (uint8, uint, error) {
    conn, err := net.Dial("tcp", mothershipAddr+":9997")
    if err != nil {
        return 0, 0, fmt.Errorf("erro ao conectar ao servidor de IDs: %v", err)
    }
    defer conn.Close()

    buf := make([]byte, 2)
    conn.SetReadDeadline(time.Now().Add(3 * time.Second))
    _, err = conn.Read(buf)
    if err != nil {
        return 0, 0, fmt.Errorf("timeout ou erro ao receber ID: %v", err)
    }

    id := buf[0]
    updateFrequency := uint(buf[1])
    fmt.Printf("✅ ID recebido da nave-mãe: %d (updateFrequency=%d)\n", id, updateFrequency)
    return id, updateFrequency, nil
}


// initConnection agora aceita o endereço completo "IP:PORT"
func initConnection(targetAddr string) (*RoverMLConnection, error) {
	motherAddr, err := net.ResolveUDPAddr("udp", targetAddr)
	if err != nil {
		return nil, err
	}

	roverConn, err := net.ListenUDP("udp", &net.UDPAddr{Port: 0})
	if err != nil {
		return nil, err
	}

	fmt.Printf("✅ Conexão UDP local na porta %d -> Alvo %s\n", roverConn.LocalAddr().(*net.UDPAddr).Port, motherAddr)

	return &RoverMLConnection{
		Conn: roverConn,
		Addr: motherAddr,
	}, nil
}

func NewRoverSystem(motherUDP string, motherIP string) *RoverSystem{


	// 2. Pedir ID (TCP Porta 9997 fixa)
	// Nota: Garante que a tua função requestID usa a porta 9997 internamente ou concatena aqui
	roverID, updateFrequency, err := requestID(motherIP) 
	if err != nil {
		fmt.Println("❌ Erro ao obter ID:", err)
		return nil
	}

	// 3. Conectar UDP
	roverConn, err := initConnection(motherUDP)
	if err != nil {
		return nil
	}
	
	return &RoverSystem{
		RoverBase: &RoverBase{
			ID: roverID,
		},
		ML: &RoverMLState{
			ActiveMissions:      0,
			Cond:                sync.NewCond(&sync.Mutex{}),
			CondMu:              sync.Mutex{},
			ExpectedSeq:         0,
			Waiting:             false,
			MissionReceivedChan: make(chan bool, 1),
			Buffer:              make(map[uint16]ml.Packet),
			BufferMu:            sync.Mutex{},
			Window:              pl.NewWindow(),
		},
		TS: &ts.RoverTSState{
			State:           "Idle",
			Battery:         100,
			Speed:           0.0,
			UpdateFrequency: updateFrequency,
		},

		MLConn: roverConn,
		CurrentPos: utils.Coordinate{
			Latitude:  1.000 + float64(roverID)*0.001,
			Longitude: -1.000 + float64(roverID)*0.001,
		},

		Devices: &Devices{
			GPS: devices.NewMockGPS(utils.Coordinate{
				Latitude:  1.000 + float64(roverID)*0.001,
				Longitude: -1.000 + float64(roverID)*0.001,
			}),
			Thermometer:      devices.NewMockThermometer(),
			Battery:          devices.NewMockBattery(100),
			Camera:           devices.NewMockCamera(),
			ChemicalAnalyzer: devices.NewMockChemicalAnalyzer(),
		},
	}
}

