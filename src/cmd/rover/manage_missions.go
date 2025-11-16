package main

import (
	"fmt"
	"time"
)

func (rover *Rover) manageMissions() {
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
