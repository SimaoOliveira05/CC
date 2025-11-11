package main

import (
	"fmt"
	"time"
)

func (rv *Rover) manageMissions() {
	// Espera até que não haja missões ativas
		rv.cond.L.Lock()
		for rv.GetActiveMissions() != 0 {
			rv.cond.Wait() // Espera até todas as missões acabarem
		}
		rv.cond.L.Unlock()

		// Se não estiver à espera de missões, request de novas missões
		if !rv.waiting {
			rv.sendRequest()
			print("")
			received := <-rv.missionReceivedChan
			if received { //Nave-mãe enviou missões
				rv.waiting = true
			} else {
				// Nave mãe não tem missões para enviar, esperamos 5 segundos para pedir outra vez
				fmt.Println("🚫 Sem missões disponíveis.")
				time.Sleep(5 * time.Second)
			}
		}
}

// Para alterar a flag:
func (r *Rover) IncrementActiveMission() {
    r.mu.Lock()
    defer r.mu.Unlock()
    r.activeMissions++
}

// Para ler a flag:
func (r *Rover) GetActiveMissions() uint8 {
    r.mu.Lock()
    defer r.mu.Unlock()
    return r.activeMissions
}

// Para decrementar a flag:
func (r *Rover) DecrementActiveMission() {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.activeMissions > 0 {
		r.activeMissions--
		if r.activeMissions == 0 {
			r.waiting = false
			r.cond.L.Lock()
			r.cond.Signal()
			r.cond.L.Unlock()
		}
	}
}
