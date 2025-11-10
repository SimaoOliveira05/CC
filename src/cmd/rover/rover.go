package main

import (
	"fmt"
	"net"
	"src/config"
	"os"
	"sync"
	"time"
	"src/internal/ml"
	//"strconv"
)

type RoverMlConection struct {
	conn         *net.UDPConn // Conexão UDP com a nave-mãe
	seqNum       uint32 // Número de sequência esperado para envio
}

type Window struct {
	lastAckReceived int16
	window          map[uint32](chan int8) // pacotes enviados mas ainda não ACKed
	mu        sync.Mutex
}

type Rover struct {
    id                uint8
	activeMissions    uint8
	mu                sync.Mutex
	cond              *sync.Cond
	waiting           bool
	missionReceivedChan chan bool
}

func initConnection(mothershipAddr string) (*RoverMlConection, error) {
	udpAddr, err := net.ResolveUDPAddr("udp", mothershipAddr+":9999")

	if err != nil {
		return nil, fmt.Errorf("erro ao resolver endereço UDP da nave-mãe: %v", err)
	}

	roverConn, err := net.DialUDP("udp", nil, udpAddr)

	if err != nil {
		return nil, fmt.Errorf("erro ao conectar: %v", err)
	}

	RoverMlConection := RoverMlConection{
		conn:   roverConn,
		seqNum: 0,
	}

	return &RoverMlConection, nil
}

func main() {

	// Verifica se o argumento do id foi passado
    if len(os.Args) < 2 {
        fmt.Println("Use: ./rover1 <id_do_rover>")
        return
    }
    //idInt, err := strconv.Atoi(os.Args[1])
	//if err != nil {
	//	fmt.Println("ID do rover inválido:", err)
	//	return
	//}
	//roverID := uint8(idInt)

	// Inicializa configuração (isRover = true)
	config.InitConfig(true)
	config.PrintConfig()

	// Inicia conexão com a nave-mãe
	mothershipAddr := config.GetMotherIP()
	roverConn, err := initConnection(mothershipAddr)
	if err != nil {
		fmt.Println("❌ Erro ao inicializar conexão:", err)
		return
	}
	defer roverConn.conn.Close()

	// Cria o Rover
	rover := Rover{
		id:         0,
		activeMissions: 0,
		mu:         sync.Mutex{},
		cond:       sync.NewCond(&sync.Mutex{}),
		waiting:   false,
		missionReceivedChan: make(chan bool, 1), //Channel para saber se a nave mãe enviou missões
	}

	// Cria a Window
	window := Window{
		window:     make(map[uint32](chan int8)), // Janela de tamanho máximo para SeqNum uint16
		mu:   sync.Mutex{},
	}

	// Inicia o receiver de pacotes
	go receiver(&rover, roverConn, &window)

	// Loop principal
	for {
		// Gerencia missões
		manageMissions(&rover, roverConn, &window)
	}
}




func generate(mission ml.MissionData, rover *Rover, c *RoverMlConection, window *Window){

	rover.IncrementActiveMission()
	defer rover.DecrementActiveMission()

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
				sendReport(mission,true, c, window, rover)
				return
			case <-ticker.C:
				// Enviar report periódico
				sendReport(mission,false, c, window, rover)
			}
		}
	} else {
		// Modo sem updates: apenas espera Duration e envia um report final
		<-deadline.C
		// Termina quando Duration expirar
		sendReport(mission,true, c, window, rover)
		return
	}
}

