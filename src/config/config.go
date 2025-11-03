package config

import (
	"flag"
	"fmt"
)

// Config guarda a configuração de IPs usados pelo sistema MissionLink.
type Config struct {
	MotherIP string
	RoverIP  string
}

// GlobalConfig é a instância global acessível em qualquer package.
var GlobalConfig Config

// InitConfig inicializa a configuração com flags.
// Se isRover = true, lê também o IP do rover (para debug).
func InitConfig(isRover bool) {
	flag.StringVar(&GlobalConfig.MotherIP, "mother-ip", "127.0.0.1", "Endereço IP da Nave Mãe")

	if isRover {
		flag.StringVar(&GlobalConfig.RoverIP, "rover-ip", "127.0.0.1", "Endereço IP do Rover (debug)")
	}

	flag.Parse()
}

// PrintConfig mostra os IPs configurados (debug).
func PrintConfig() {
	fmt.Printf("🛰️ Nave Mãe IP: %s\n", GlobalConfig.MotherIP)
	if GlobalConfig.RoverIP != "" {
		fmt.Printf("🤖 Rover IP: %s\n", GlobalConfig.RoverIP)
	}
}

// GetMotherIP devolve o IP da Nave Mãe.
func GetMotherIP() string {
	return GlobalConfig.MotherIP
}

// GetRoverIP devolve o IP do Rover (opcional).
func GetRoverIP() string {
	return GlobalConfig.RoverIP
}
