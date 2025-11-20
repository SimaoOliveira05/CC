package main

import (
	"fmt"
	"src/config"
	"src/internal/core"
)

type MotherShip struct {
	*core.MotherShip // Embedding - herda todos os campos
}



func main() {
	config.InitConfig(false)
	config.PrintConfig()

	fmt.Println("🛰️ Nave-Mãe à escuta...")

	// Cria o estado da Nave-Mãe
	mothership := MotherShip{
		MotherShip: core.NewMotherShip(),
	}

	// ✅ Inicia API Server para Ground Control (já foi criado no construtor)
	
	go mothership.APIServer.Start("8080")

	// Servidor de atribuição de IDs (TCP)
	go mothership.idAssignmentServer("9997")

	// Goroutine para ler pacotes UDP
	go mothership.receiver("9999")

	go mothership.telemetryReceiver("9998")

	// Loop infinito
	select {}
}
