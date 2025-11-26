package main

import (
	"fmt"
	"net"
	"src/config"
	"src/internal/core"
	"src/internal/devices"
	"src/internal/ml"
	"src/internal/ts"
	"src/utils"
	pl "src/utils/packetsLogic"
	"sync"
	"time"
)

type Rover struct {
	*core.RoverBase
	ML         *core.RoverMLState
	TS         *ts.RoverTSState
	MLConn     *core.RoverMLConnection
	Devices    *core.Devices
	CurrentPos utils.Coordinate
}

func NewRover(id uint8, mlConn *core.RoverMLConnection, updateFrequency uint) *Rover {
	return &Rover{
		RoverBase: &core.RoverBase{
			ID: id,
		},
		ML: &core.RoverMLState{
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
			State:   "Idle",
			Battery: 100,
			Speed:   0.0,
			UpdateFrequency: updateFrequency,
		},

		MLConn: mlConn,
		CurrentPos: utils.Coordinate{
			Latitude:  1.000 + float64(id)*0.001,
			Longitude: -1.000 + float64(id)*0.001,
		},

		Devices: &core.Devices{
			GPS: devices.NewMockGPS(utils.Coordinate{
				Latitude:  1.000 + float64(id)*0.001,
				Longitude: -1.000 + float64(id)*0.001,
			}),
			Thermometer: devices.NewMockThermometer(),
			Battery:     devices.NewMockBattery(100),
		},
	}
}

func initConnection(mothershipAddr string) (*core.RoverMLConnection, error) {
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

	RoverMlConection := core.RoverMLConnection{
		Conn: roverConn,
		Addr: motherAddr,
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
	roverID, updateFrequency, err := requestID(mothershipAddr)
	if err != nil {
		fmt.Println("❌ Erro ao obter ID:", err)
		return
	}

	// Inicia conexão UDP com a nave-mãe
	roverConn, err := initConnection(mothershipAddr)
	if err != nil {
		fmt.Println("❌ Erro ao inicializar conexão:", err)
		return
	}
	defer roverConn.Conn.Close()

	// Cria o Rover
	rover := NewRover(roverID, roverConn, updateFrequency)

	// Inicia o receiver de pacotes
	go rover.receiver()

	go rover.telemetrySender(config.GetMotherIP())

	// Loop principal
	for {
		// Gerencia missões
		rover.manageMissions()
	}
}

func (rover *Rover) generate(mission ml.MissionData) {

	rover.IncrementActiveMission()
	defer rover.DecrementActiveMission()

	fmt.Printf("🎯 Missão %d recebida: TaskType=%d\n", mission.MsgID, mission.TaskType)

	// 1. Move para a localização da missão
	fmt.Printf("🚀 Movendo para coordenadas (%.4f, %.4f)\n", mission.Coordinate.Latitude, mission.Coordinate.Longitude)
	if err := core.MoveTo(
		&rover.CurrentPos,
		mission.Coordinate,
		rover.Devices.GPS,
		rover.Devices.Battery,
	); err != nil {
		fmt.Printf("❌ Erro ao mover: %v\n", err)
		return
	}
	fmt.Printf("✅ Chegou ao destino. Iniciando tarefa...\n")

	// 2. Executa a tarefa com timer
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

				rover.sendReport(mission, true)
				return
			case <-ticker.C:
				// Enviar report periódico
				rover.sendReport(mission, false)
			}
		}
	} else {
		// Modo sem updates: apenas espera Duration e envia um report final
		<-deadline.C
		// Termina quando Duration expirar
		rover.sendReport(mission, true)
	}

	// 3. Consome bateria da execução da tarefa
	core.ConsumeBattery(rover.Devices.Battery, uint8(core.TaskBatteryRate))
}
