package config

import (
	"flag"
	"fmt"
)

type Config struct {
	MothershipIP   string
	MothershipPort int
	RoverIP        string
}

// Global configuration instance
var GlobalConfig Config

// Initialize flags and populate the global config
func InitConfig(isAgent bool) {
	// Define flags
	flag.StringVar(&GlobalConfig.MothershipIP, "mothership-ip", "10.0.0.10", "Mothership IP address")
	flag.IntVar(&GlobalConfig.MothershipPort, "mothership-port", 9999, "Mothership port")
	if isAgent {
		flag.StringVar(&GlobalConfig.RoverIP, "rover-ip", "10.0.1.20", "Rover IP address")
	}

	// Parse flags
	flag.Parse()
}

// Debug function to print the current configuration
func PrintConfig() {
	fmt.Printf("Mothership: %s:%d\n", GlobalConfig.MothershipIP, GlobalConfig.MothershipPort)
	if GlobalConfig.RoverIP != "" {
		fmt.Printf("Rover IP: %s\n", GlobalConfig.RoverIP)
	}
}

// GetMothershipAddr returns the full mothership address (IP:Port)
func GetMothershipAddr() string {
	return fmt.Sprintf("%s:%d", GlobalConfig.MothershipIP, GlobalConfig.MothershipPort)
}

// GetMothershipIP returns the mothership IP address
func GetMothershipIP() string {
	return GlobalConfig.MothershipIP
}

// GetMothershipPort returns the mothership port
func GetMothershipPort() int {
	return GlobalConfig.MothershipPort
}

// GetRoverIP returns the rover IP address
func GetRoverIP() string {
	return GlobalConfig.RoverIP
}