package main

import (
	"fmt"
	"src/config"
	"src/internal/core"
)

type MotherShip struct {
	*core.MotherShip
}

func main() {
	// A mothership pode não usar o IP para nada crítico, mas inicializamos o config na mesma
	config.InitConfig(false)
	config.PrintConfig()

	fmt.Println("🛰️ Nave-Mãe a iniciar nos portos padrão...")

	mothership := MotherShip{
		MotherShip: core.NewMotherShip(),
	}

	// 🔒 PORTAS FIXAS (Hardcoded)
	go mothership.APIServer.Start("8080")       // API Ground Control
	go mothership.idAssignmentServer("9997")    // TCP ID Attribution
	go mothership.receiver("9999")              // UDP Communication
	go mothership.telemetryReceiver("9998")     // TCP Telemetry

	select {}
}