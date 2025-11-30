package config

import (
	"flag"
	"fmt"
)

type Config struct {
	MotherIP string
}

var GlobalConfig Config

func InitConfig(isRover bool) {
	// IP padrão é localhost
	flag.StringVar(&GlobalConfig.MotherIP, "ms-ip", "127.0.0.1", "Endereço IP da Nave Mãe")
	flag.Parse()
}

func PrintConfig() {
	fmt.Printf("🔧 Configuração: Nave Mãe em %s (Portas Padrão)\n", GlobalConfig.MotherIP)
}

// Helper para obter o endereço UDP completo da Mãe (porta fixa 9999)
func GetMotherUDPAddr() string {
	return fmt.Sprintf("%s:9999", GlobalConfig.MotherIP)
}