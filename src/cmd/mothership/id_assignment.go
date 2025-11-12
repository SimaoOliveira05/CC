package main

import (
    "fmt"
    "net"
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

func (ms *MotherShip) idAssignmentServer(port string, idManager *IDManager) {
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

    // Atribui novo ID
    id := idManager.GetNextID()

    // Envia ID para o rover
    buf := make([]byte, 1)
    buf[0] = id
    _, err := conn.Write(buf)
    if err != nil {
        fmt.Println("❌ Erro ao enviar ID:", err)
        return
    }

    fmt.Printf("✅ ID %d atribuído a novo rover\n", id)
}