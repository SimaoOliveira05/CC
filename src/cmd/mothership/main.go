package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"src/config"
	"src/internal/core"
	"src/internal/ml"
)

type MotherShip struct {
	*core.MotherShip            // Embedding - herda todos os campos
	apiServer        *APIServer // ✅ Campo para o API Server
}

func initConnection(mothershipAddr string) (*MotherShip, error) {

	mothershipConn, err := net.ListenUDP("udp", &net.UDPAddr{
		IP:   net.ParseIP(mothershipAddr),
		Port: 9999,
	})

	if err != nil {
		return nil, fmt.Errorf("erro ao conectar: %v", err)
	}

	// Cria o estado da Nave-Mãe
	mothership := MotherShip{
		MotherShip: core.NewMotherShip(mothershipConn),
	}

	// Carrega missões iniciais de um ficheiro JSON
	err = loadMissionsFromJSON("missions.json", mothership.MissionQueue)
	if err != nil {
		return nil, fmt.Errorf("erro ao carregar missões iniciais: %v", err)
	}

	return &mothership, nil
}

func main() {
	config.InitConfig(false)
	config.PrintConfig()

	fmt.Println("🛰️ Nave-Mãe à escuta...")

	mothership, err := initConnection(config.GetMotherIP())
	if err != nil {
		fmt.Println("Erro ao iniciar conexão:", err)
		return
	}

	idManager := NewIDManager()

	// ✅ Inicia API Server para Ground Control
	mothership.apiServer = NewAPIServer(mothership.MotherShip)
	go mothership.apiServer.Start("8080")

	// Servidor de atribuição de IDs (TCP)
	go mothership.idAssignmentServer("9997", idManager)

	// Goroutine para ler pacotes UDP
	go mothership.receiver()

	go mothership.telemetryReceiver("9998")

	// Loop infinito
	select {}
}

// loadMissionsFromJSON lê missões de um ficheiro JSON e coloca-as na missionQueue
func loadMissionsFromJSON(filename string, queue chan ml.MissionState) error {
	file, err := os.Open(filename)
	if err != nil {
		return fmt.Errorf("erro ao abrir ficheiro: %v", err)
	}
	defer file.Close()

	data, err := ioutil.ReadAll(file)
	if err != nil {
		return fmt.Errorf("erro ao ler ficheiro: %v", err)
	}

	var missions []ml.MissionState
	if err := json.Unmarshal(data, &missions); err != nil {
		return fmt.Errorf("erro ao fazer unmarshal do JSON: %v", err)
	}

	for _, mission := range missions {
		queue <- mission
	}

	fmt.Printf("📋 %d missões enfileiradas\n", len(missions))
	return nil
}
