package main

import (
	"fmt"
	"time"
	"src/internal/ml"
	"src/internal/core"
)


func (rover *Rover) ExecuteMission(mission ml.MissionData) {

	rover.IncrementActiveMission()
	defer rover.DecrementActiveMission()

	fmt.Printf("🎯 Missão %d recebida: TaskType=%d\n", mission.MsgID, mission.TaskType)

	// 1. Move para a localização da missão
	fmt.Printf("🚀 Movendo para coordenadas (%.4f, %.4f)\n", mission.Coordinate.Latitude, mission.Coordinate.Longitude)
	if err := core.MoveTo(
		&rover.RoverBase.CurrentPos,
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
	core.ConsumeBattery(rover.Devices.Battery, core.TaskBatteryRate)
}


func (rover *Rover) manageMissions() {
	for {
		// Espera até que não haja missões ativas
		rover.ML.Cond.L.Lock()
		for rover.GetActiveMissions() != 0 {
			rover.ML.Cond.Wait() // Espera até todas as missões acabarem
		}
		rover.ML.Cond.L.Unlock()

		// Se não estiver à espera de missões, request de novas missões
		if !rover.ML.Waiting {
			rover.sendRequest()
			print("")
			received := <-rover.ML.MissionReceivedChan
			if received { //Nave-mãe enviou missões
				rover.ML.Waiting = true
			} else {
				// Nave mãe não tem missões para enviar, esperamos 5 segundos para pedir outra vez
				fmt.Println("🚫 Sem missões disponíveis.")
				time.Sleep(5 * time.Second)
			}
		}
	}
}

// Para alterar a flag:
func (rover *Rover) IncrementActiveMission() {
	rover.ML.CondMu.Lock()
	defer rover.ML.CondMu.Unlock()
	rover.ML.ActiveMissions++
}

// Para ler a flag:
func (rover *Rover) GetActiveMissions() uint8 {
	rover.ML.CondMu.Lock()
	defer rover.ML.CondMu.Unlock()
	return rover.ML.ActiveMissions
}

// Para decrementar a flag:
func (rover *Rover) DecrementActiveMission() {
	rover.ML.CondMu.Lock()
	defer rover.ML.CondMu.Unlock()
	if rover.ML.ActiveMissions > 0 {
		rover.ML.ActiveMissions--
		if rover.ML.ActiveMissions == 0 {
			rover.ML.Waiting = false
			rover.ML.Cond.L.Lock()
			rover.ML.Cond.Signal()
			rover.ML.Cond.L.Unlock()
		}
	}
}
