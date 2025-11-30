package main

import (
	"fmt"
	"src/config"
	"src/internal/core"
	"src/internal/ml"
	"time"
)

type Rover struct {
	*core.RoverSystem
}



func main() {
	config.InitConfig(true) // Lê flag -ms-ip
	config.PrintConfig()

	// 1. Obter endereço da mãe (IP da flag + Porta Fixa 9999)
	motherUDPAddr := config.GetMotherUDPAddr()
	motherIP := config.GlobalConfig.MotherIP

	rover := Rover{
		RoverSystem: core.NewRoverSystem(
			motherUDPAddr,
			motherIP,
		),
	}

	go rover.receiver()
	go rover.telemetrySender(motherIP) // Telemetria usa porta fixa 9998 internamente?

	for {
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
