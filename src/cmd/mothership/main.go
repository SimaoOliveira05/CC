package main

import (
	"fmt"
	"src/config"
	"src/internal/core"
)

// MotherShip struct embedding core.MotherShip
type MotherShip struct {
	*core.MotherShip
}

func main() {
	// The mothership may not use the IP for anything critical, but we initialize the config for consistency
	config.InitConfig(false, true) // Read flag -ms-ip and print config

	fmt.Println("🛰️ Mother Ship starting on default ports...")

	mothership := MotherShip{
		MotherShip: core.NewMotherShip(),
	}

	go mothership.APIServer.Start(config.API_PORT)             // API Ground Control
	go mothership.idAssignmentServer(config.TCP_ID_PORT)       // TCP ID Attribution
	go mothership.receiver(config.UDP_COMM_PORT)               // UDP Communication
	go mothership.telemetryReceiver(config.TCP_TELEMETRY_PORT) // TCP Telemetry

	select {}
}
