package config

import (
	"flag"
	"fmt"
	"os"
	"encoding/json"
)

// Port configuration variables
var (
    API_PORT           string
    TCP_ID_PORT        string
    UDP_COMM_PORT      string
    TCP_TELEMETRY_PORT string
)

// Config holds the global configuration settings
type Config struct {
	MotherIP string
}

var GlobalConfig Config

// InitConfig initializes the global configuration from command-line flags
func InitConfig(isRover bool, print bool) {
	// Default IP is localhost
	flag.StringVar(&GlobalConfig.MotherIP, "ms-ip", "127.0.0.1", "Mother Ship IP Address")
	flag.Parse()

	// Read ports from config.json
	file, err := os.Open("config.json")
	if err != nil {
		panic(err)
	}
	defer file.Close()

	// Configure decoder
	conf := struct {
		API_PORT           int `json:"API_PORT"`
		TCP_ID_PORT        int `json:"TCP_ID_PORT"`
		UDP_COMM_PORT      int `json:"UDP_COMM_PORT"`
		TCP_TELEMETRY_PORT int `json:"TCP_TELEMETRY_PORT"`
	}{}
	
	// Decode JSON configuration
	if err:= json.NewDecoder(file).Decode(&conf); err != nil {
		panic(err)
	}

	// Assign ports to global variables
    API_PORT = fmt.Sprintf("%d", conf.API_PORT)
    TCP_ID_PORT = fmt.Sprintf("%d", conf.TCP_ID_PORT)
    UDP_COMM_PORT = fmt.Sprintf("%d", conf.UDP_COMM_PORT)
    TCP_TELEMETRY_PORT = fmt.Sprintf("%d", conf.TCP_TELEMETRY_PORT)

	if print {
        PrintConfig()
    }
}

// PrintConfig prints the current configuration settings
func PrintConfig() {
    println("API_PORT:", API_PORT)
    println("TCP_ID_PORT:", TCP_ID_PORT)
    println("UDP_COMM_PORT:", UDP_COMM_PORT)
    println("TCP_TELEMETRY_PORT:", TCP_TELEMETRY_PORT)
}

// Helper to get the full UDP address of the Mother Ship (fixed port 9999)
func GetMotherUDPAddr() string {
	return fmt.Sprintf("%s:9999", GlobalConfig.MotherIP)
}