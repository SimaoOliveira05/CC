package main

import (
	"fmt"
	"os"
	"os/signal"
	"src/config"
	"src/internal/core"
	"src/utils/metrics"
	"syscall"
)

// Rover struct embedding core.RoverSystem
type Rover struct {
	*core.RoverSystem
}

func main() {
	config.InitConfig(true, true) // Read flag -ms-ip and print config

	// Initialize metrics if in test mode
	metrics.InitGlobalMetrics(config.IsTestMode())
	if config.IsTestMode() {
		fmt.Println("📊 Test mode enabled - collecting metrics")
	}

	// Obtain mothership addresses from config
	mothershipUDPAddr := config.GetMotherUDPAddr()
	mothershipTCPID := config.GetMotherTCPIDAddr()
	mothershipTelemetry := config.GetMotherTelemetryAddr()

	// Initialize Rover system
	roverSys := core.NewRoverSystem(mothershipUDPAddr, mothershipTCPID)
	if roverSys == nil {
		panic("Failed to initialize Rover System")
	}

	// Setup graceful shutdown to print metrics
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigChan
		if config.IsTestMode() && metrics.GlobalMetrics != nil {
			// Include rover ID in filename
			filename := fmt.Sprintf("../metrics/rover_%d_metrics.json", roverSys.ID)
			metrics.GlobalMetrics.ExportToJSON(filename)
		}
		os.Exit(0)
	}()

	// Create Rover instance
	rover := Rover{RoverSystem: roverSys}

	// Start Rover services
	go rover.receiver()
	go rover.telemetrySender(mothershipTelemetry)
	go rover.manageMissions()
	go rover.batteryMonitor() // Monitor battery level continuously

	select {}
}
