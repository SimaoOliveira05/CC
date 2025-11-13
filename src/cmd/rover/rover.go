package main

import (
	"fmt"
	"net"
	"src/config"
	"src/internal/ml"
	"src/utils/packetsLogic"
	"sync"
	"time"
)

type RoverMlConection struct {
	conn   *net.UDPConn // Conexão UDP com a nave-mãe
	addr   *net.UDPAddr // Endereço da nave-mãe
	seqNum uint32       // Número de sequência esperado para envio
}

type Rover struct {
	id                  uint8
	activeMissions      uint8
	mu                  sync.Mutex
	cond                *sync.Cond
	waiting             bool
	missionReceivedChan chan bool
	conn                *RoverMlConection
	window              *packetslogic.Window
	expectedSeq         uint16
	buffer              map[uint16]ml.Packet
	bufferMu            sync.Mutex
}

func initConnection(mothershipAddr string) (*RoverMlConection, error) {
	// Resolve o endereço da nave-mãe
	motherAddr, err := net.ResolveUDPAddr("udp", mothershipAddr+":9999")
	if err != nil {
		return nil, fmt.Errorf("erro ao resolver endereço UDP da nave-mãe: %v", err)
	}

	// Abre uma porta UDP local (porta 0 = qualquer porta livre)
	roverConn, err := net.ListenUDP("udp", &net.UDPAddr{Port: 0})
	if err != nil {
		return nil, fmt.Errorf("erro ao criar conexão UDP: %v", err)
	}

	fmt.Printf("✅ Conexão UDP aberta na porta %d\n", roverConn.LocalAddr().(*net.UDPAddr).Port)

	RoverMlConection := RoverMlConection{
		conn:   roverConn,
		addr:   motherAddr,
		seqNum: 0,
	}

	return &RoverMlConection, nil
}

func main() {

	// Inicializa configuração (isRover = true)
	config.InitConfig(true)
	config.PrintConfig()

	// Inicia conexão com a nave-mãe
	mothershipAddr := config.GetMotherIP()

	// 🆔 Solicita ID à nave-mãe via TCP
	roverID, err := requestID(mothershipAddr)
	if err != nil {
		fmt.Println("❌ Erro ao obter ID:", err)
		return
	}

	roverConn, err := initConnection(mothershipAddr)
	if err != nil {
		fmt.Println("❌ Erro ao inicializar conexão:", err)
		return
	}
	defer roverConn.conn.Close()

	// Cria o Rover
	rover := Rover{
		id:                  roverID,
		activeMissions:      0,
		mu:                  sync.Mutex{},
		cond:                sync.NewCond(&sync.Mutex{}),
		waiting:             false,
		missionReceivedChan: make(chan bool, 1), //Channel para saber se a nave mãe enviou missões
		conn:                roverConn,
		window: &packetslogic.Window{
			LastAckReceived: -1,
			Window:          make(map[uint32](chan int8)),
			Mu:              sync.Mutex{},
		},
		expectedSeq: 0,
		buffer:      make(map[uint16]ml.Packet),
		bufferMu:    sync.Mutex{},
	}

	// Inicia o receiver de pacotes
	go rover.receiver()

	go rover.telemetrySender(config.GetMotherIP())

	// Loop principal
	for {
		// Gerencia missões
		rover.manageMissions()
	}
}

func (rv *Rover) generate(mission ml.MissionData) {

	rv.IncrementActiveMission()
	defer rv.DecrementActiveMission()

	deadline := time.NewTimer(time.Duration(mission.Duration) * time.Second)
	defer deadline.Stop()

	if mission.UpdateFrequency > 0 {
		// Modo periódico: enviar reports a cada UpdateFrequency
		ticker := time.NewTicker(time.Duration(mission.UpdateFrequency) * time.Second)
		defer ticker.Stop()

		for {
			select {

			case <-deadline.C:
				// Termina quando Duration expirar
				rv.sendReport(mission, true)
				return
			case <-ticker.C:
				// Enviar report periódico
				rv.sendReport(mission, false)
			}
		}
	} else {
		// Modo sem updates: apenas espera Duration e envia um report final
		<-deadline.C
		// Termina quando Duration expirar
		rv.sendReport(mission, true)
		return
	}
}
