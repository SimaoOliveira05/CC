package main

import (
	"src/config"
	"src/internal/core"
)

// Rover struct embedding core.RoverSystem
type Rover struct {
	*core.RoverSystem
}

func main() {
	config.InitConfig(true, true) // Read flag -ms-ip and print config

	// Obtain mothership addresses from config
	mothershipUDPAddr := config.GetMotherUDPAddr()
	mothershipTCPID := config.GetMotherTCPIDAddr()
	mothershipTelemetry := config.GetMotherTelemetryAddr()

	// Initialize Rover system
	roverSys := core.NewRoverSystem(mothershipUDPAddr, mothershipTCPID)
	if roverSys == nil {
		panic("❌ Failed to initialize Rover System")
	}

	// Create Rover instance
	rover := Rover{RoverSystem: roverSys}

	// Start Rover services
	go rover.receiver()
	go rover.telemetrySender(mothershipTelemetry)
	go rover.manageMissions()
	go rover.batteryMonitor() // Monitor battery level continuously

	select {}
}
