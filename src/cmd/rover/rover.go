package main

import (
    "fmt"
    "net"
    "src/config"
    "src/internal/ml"
)

func main() {
    // Inicializa configuração (isRover = true)
    config.InitConfig(true)
    config.PrintConfig()

    // Obtém o endereço da mothership
    mothershipAddr := config.GetMotherIP()

    // Resolve endereço UDP
    addr, err := net.ResolveUDPAddr("udp", mothershipAddr+":9999")
    if err != nil {
        fmt.Println("❌ Erro ao resolver endereço:", err)
        return
    }

    // Conecta à mothership
    conn, err := net.DialUDP("udp", nil, addr)
    if err != nil {
        fmt.Println("❌ Erro ao conectar:", err)
        return
    }
    defer conn.Close()

    fmt.Printf("🤖 Rover conectado à Mothership em %s\n", mothershipAddr)

    // Cria e envia pedido de missão
    requestPacket := ml.Packet{
        MsgType: ml.MSG_REQUEST,
        SeqNum:  1,
        AckNum:  0,
        Payload: []byte{},
    }
    requestPacket.Checksum = ml.Checksum(requestPacket.Payload)

    _, err = conn.Write(requestPacket.ToBytes())
    if err != nil {
        fmt.Println("❌ Erro ao enviar pedido:", err)
        return
    }

    fmt.Println("📤 Pedido de missão enviado!")

    // Aguarda resposta
    buf := make([]byte, 1024)
    n, err := conn.Read(buf)
    if err != nil {
        fmt.Println("❌ Erro ao receber resposta:", err)
        return
    }

    response := ml.FromBytes(buf[:n])
    fmt.Printf("📥 Resposta recebida: MsgType=%d\n", response.MsgType)

    if response.MsgType == ml.MSG_MISSION {
        missionData := ml.DataFromBytes(response.Payload)
        fmt.Printf("📍 Missão recebida:\n%s\n", missionData.String())
    }
}