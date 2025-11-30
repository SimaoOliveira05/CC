package main

import (
	"fmt"
	"net"
	"src/internal/ts"
	"src/utils"
	"sync"
)

type IDManager struct {
	nextID uint8
	mu     sync.Mutex
}

func NewIDManager() *IDManager {
	return &IDManager{nextID: 1}
}

func (idm *IDManager) GetNextID() uint8 {
	idm.mu.Lock()
	defer idm.mu.Unlock()
	id := idm.nextID
	idm.nextID++
	return id
}

func (ms *MotherShip) idAssignmentServer(port string) {

	idManager := NewIDManager()

	listener, err := net.Listen("tcp", "0.0.0.0:"+port)
	if err != nil {
		ms.EventLogger.Log("ERROR", "IDHandler", fmt.Sprintf("Erro ao iniciar servidor de IDs: %v", err), nil)
		return
	}
	defer listener.Close()

	ms.EventLogger.Log("INFO", "IDHandler", fmt.Sprintf("Servidor de atribuição de IDs à escuta na porta %s", port), nil)

	for {
		conn, err := listener.Accept()
		if err != nil {
			ms.EventLogger.Log("ERROR", "IDHandler", fmt.Sprintf("Erro ao aceitar conexão: %v", err), nil)
			continue
		}
		go ms.handleIDRequest(conn, idManager)
	}
}

func (ms *MotherShip) handleIDRequest(conn net.Conn, idManager *IDManager) {
    defer conn.Close()

    id := idManager.GetNextID()
    var updateFrequency uint = 2 // exemplo: 2 segundos

    buf := []byte{id, byte(updateFrequency)}
    _, err := conn.Write(buf)
    if err != nil {
        ms.EventLogger.Log("ERROR", "IDHandler", fmt.Sprintf("Erro ao enviar ID/updateFrequency: %v", err), nil)
        return
    }

    ms.EventLogger.Log("INFO", "IDHandler", fmt.Sprintf("ID %d atribuído a novo rover (updateFrequency=%d)", id, updateFrequency), nil)
    ms.RoverInfo.AddRover(&ts.RoverTSState{
        ID:       id,
        State:    "Desconhecido",
        Battery:  100,
        Speed:    0.0,
        Position: utils.Coordinate{Latitude: 0, Longitude: 0},
        UpdateFrequency: updateFrequency,
    })
}
