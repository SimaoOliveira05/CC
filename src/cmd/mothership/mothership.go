package main

import (
    "fmt"
    "net"
    "src/config"
    "src/internal/ml"
    "time"
    "src/utils"
)

func main() {

    config.InitConfig(false)
    config.PrintConfig()

    addr, _ := net.ResolveUDPAddr("udp", fmt.Sprintf(":%d", config.GetMothershipPort()))
    conn, _ := net.ListenUDP("udp", addr)
    defer conn.Close()

    fmt.Println("🛰️ Nave-Mãe à escuta...")

    // Creates the Mission Manager
    missionManager := ml.NewMissionManager()  // ← MUDA de mission para ml

    // Goroutine (thread) para ler pacotes UDP
    go udpListener(conn, missionManager)

    // infinite loop :)
    select {}
}

// udpListener - thread que lê continuamente da porta UDP
func udpListener(conn *net.UDPConn, mm *ml.MissionManager) {  // ← MUDA de mission para ml
    buf := make([]byte, 1024)

    for {
        n, clientAddr, err := conn.ReadFromUDP(buf)
        if err != nil {
            fmt.Println("❌ Erro ao ler UDP:", err)
            continue
        }

        p := ml.FromBytes(buf[:n])
        fmt.Println("📨 Recebido pacote do tipo:", p.MsgType, "de", clientAddr)
    }
}

// handlePacket - processa cada pacote recebido numa thread separada
func handlePacket(p ml.Packet, clientAddr *net.UDPAddr, conn *net.UDPConn, mm *ml.MissionManager) {  // ← MUDA
    switch p.MsgType {
    case ml.MSG_REQUEST:
        // Gera um ID único para a missão
        missionID := uint32(time.Now().Unix())
        
        // Cria dados da missão
        payload := ml.Data{
            MsgID:           uint16(missionID),
            Coordinate:      utils.Coordinate{Latitude: 32, Longitude: 25},
            TaskType:        ml.Rescue,
            Duration:        300,
            UpdateFrequency: 20,
            Priority:        0,
        }

        // Cria o estado da missão e adiciona ao gestor
        missionState := &ml.MissionState{  // ← MUDA de mission para ml
            ID:              missionID,
            IDRover:         0,
            TaskType:        payload.TaskType,
            Duration:        time.Duration(payload.Duration) * time.Second,
            UpdateFrequency: time.Duration(payload.UpdateFrequency) * time.Second,
            LastUpdate:      time.Now(),
            CreatedAt:       time.Now(),
            Priority:        payload.Priority,
            State:           "Pending",
        }
        
        // Adiciona missão ao gestor
        mm.AddMission(missionState)
        fmt.Printf("📝 Missão %d registada no gestor\n", missionID)

        // Envia a missão ao cliente
        missionPacket := ml.Packet{
            MsgType:  ml.MSG_MISSION,
            SeqNum:   0,
            AckNum:   p.SeqNum + 1,
            Payload:  payload.ToBytes(),
        }
        
        missionPacket.Checksum = ml.Checksum(missionPacket.Payload)
        
        _, err := conn.WriteToUDP(missionPacket.ToBytes(), clientAddr)
        if err != nil {
            fmt.Println("❌ Erro ao enviar missão:", err)
            return
        }
        
        fmt.Printf("✅ Missão %d enviada para %s\n", missionID, clientAddr)

        // Inicia tracking da missão
        go trackMissionProgress(mm, missionID)

    case ml.MSG_ACK:
        fmt.Printf("✅ ACK recebido de %s (SeqNum: %d)\n", clientAddr, p.SeqNum)

    case ml.MSG_REPORT:
        fmt.Printf("📊 Relatório recebido de %s\n", clientAddr)
        // TODO: processar dados do relatório

    case ml.MSG_MISSION_END:
        fmt.Printf("🏁 Fim de missão recebido de %s\n", clientAddr)
        // TODO: atualizar estado da missão no gestor

    default:
        fmt.Printf("⚠️ Tipo de pacote desconhecido: %d\n", p.MsgType)
    }
}
