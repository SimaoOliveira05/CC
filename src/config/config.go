package config

import (
	"flag"
	"fmt"
)

type Config struct {
	MotherShipIP string
	RoverIP  string
}

// Global configuration instance
var GlobalConfig Config

// Initialize flags and populate the global config
func InitConfig(isAgent bool) {
	// Define flags
	flag.StringVar(&GlobalConfig.MotherShipIP, "mothership-ip", "127.0.0.1", "Mothership IP address")
	if isAgent {
		flag.StringVar(&GlobalConfig.RoverIP, "rover-ip", "127.0.0.1", "Rover IP address")
	}

	// Parse flags
	flag.Parse()
}

// Debug function to print the current configuration
func PrintConfig() {
	fmt.Printf("Server IP: %s\n", GlobalConfig.MotherShipIP)
	if GlobalConfig.RoverIP != "" {
		fmt.Printf("Agent IP: %s\n", GlobalConfig.RoverIP)
	}
}

// GetServerIP returns the server IP address
func GetServerIP() string {
	return GlobalConfig.MotherShipIP
}

// GetAgentIP returns the agent IP address
func GetAgentIP() string {
	return GlobalConfig.RoverIP
}