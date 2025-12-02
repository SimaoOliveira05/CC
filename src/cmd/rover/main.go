package main

import (
	"src/config"
	"src/internal/core"
)

type Rover struct {
	*core.RoverSystem
}



func main() {
	config.InitConfig(true, true) // Read flag -ms-ip and print config

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
	go rover.manageMissions()
	
	select {}
}


