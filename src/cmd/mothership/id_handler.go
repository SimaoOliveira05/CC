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
		fmt.Println("❌ Erro ao iniciar servidor de IDs:", err)
		return
	}
	defer listener.Close()

	fmt.Println("🆔 Servidor de atribuição de IDs à escuta na porta", port)

	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("❌ Erro ao aceitar conexão:", err)
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
        fmt.Println("❌ Erro ao enviar ID/updateFrequency:", err)
        return
    }

    fmt.Printf("✅ ID %d atribuído a novo rover (updateFrequency=%d)\n", id, updateFrequency)
    ms.RoverInfo.AddRover(&ts.RoverTSState{
        ID:       id,
        State:    "Desconhecido",
        Battery:  100,
        Speed:    0.0,
        Position: utils.Coordinate{Latitude: 0, Longitude: 0},
        UpdateFrequency: updateFrequency,
    })
}
