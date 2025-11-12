package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"src/config"
	"src/internal/ml"
	"sync"
)

type Window struct {
	lastAckReceived int16
	window          map[uint32](chan int8) // pacotes enviados mas ainda não ACKed
	mu              sync.Mutex
}

type RoverState struct {
	Addr        *net.UDPAddr
	SeqNum      uint16
	ExpectedSeq uint16
	Buffer      map[uint16]ml.Packet
	WindowLock  sync.Mutex
	Window      *Window // Janela deslizante específica deste rover
}

type MotherShip struct {
	conn           *net.UDPConn
	rovers         map[string]*RoverState // key: IP (ou ID do rover)
	missionManager *ml.MissionManager
	missionQueue   chan ml.MissionState
	mu             sync.Mutex
}

func initConnection(mothershipAddr string) (*MotherShip, error) {
	udpAddr, err := net.ResolveUDPAddr("udp", mothershipAddr+":9999")

	if err != nil {
		return nil, fmt.Errorf("erro ao resolver endereço UDP da nave-mãe: %v", err)
	}
	mothershipConn, err := net.ListenUDP("udp", udpAddr)

	if err != nil {
		return nil, fmt.Errorf("erro ao conectar: %v", err)
	}

	// Cria o estado da Nave-Mãe
	mothership := MotherShip{
		conn:           mothershipConn,
		rovers:         make(map[string]*RoverState),
		missionManager: ml.NewMissionManager(),
		missionQueue:   make(chan ml.MissionState, 100),
		mu:             sync.Mutex{},
	}

	// Carrega missões iniciais de um ficheiro JSON
	err = loadMissionsFromJSON("missions.json", mothership.missionQueue)
	if err != nil {
		return nil, fmt.Errorf("erro ao carregar missões iniciais: %v", err)
	}

	return &mothership, nil
}

func main() {
	config.InitConfig(false)
	config.PrintConfig()

	fmt.Println("🛰️ Nave-Mãe à escuta...")

	mothership, err := initConnection(config.GetMotherIP())
	if err != nil {
		fmt.Println("Erro ao iniciar conexão:", err)
		return
	}

	// Goroutine para ler pacotes UDP
	go mothership.receiver()

	// Loop infinito
	select {}
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

	for _, mission := range missions {
		queue <- mission
	}

	fmt.Printf("📋 %d missões enfileiradas\n", len(missions))
	return nil
}
