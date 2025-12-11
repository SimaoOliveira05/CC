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

// MotherShip struct embedding core.MotherShip
type MotherShip struct {
	*core.MotherShip
}

func main() {
	// The mothership may not use the IP for anything critical, but we initialize the config for consistency
	config.InitConfig(false, true) // Read flag -ms-ip and print config

	// Initialize metrics if in test mode
	metrics.InitGlobalMetrics(config.IsTestMode())
	if config.IsTestMode() {
		fmt.Println("📊 Test mode enabled - collecting metrics")
	}

	fmt.Println("🛰️ Mother Ship starting on default ports...")

	mothership := MotherShip{
		MotherShip: core.NewMotherShip(),
	}

	// Setup graceful shutdown to print metrics
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigChan
		if config.IsTestMode() && metrics.GlobalMetrics != nil {
			metrics.GlobalMetrics.ExportToJSON("../metrics/mothership_metrics.json")
		}
		os.Exit(0)
	}()

	go mothership.APIServer.Start(config.API_PORT)             // API Ground Control
	go mothership.idAssignmentServer(config.TCP_ID_PORT)       // TCP ID Attribution
	go mothership.receiver(config.UDP_COMM_PORT)               // UDP Communication
	go mothership.telemetryReceiver(config.TCP_TELEMETRY_PORT) // TCP Telemetry

	select {}
}
