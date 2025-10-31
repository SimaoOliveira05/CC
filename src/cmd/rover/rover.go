package main    

import (
    "fmt"
    "src/internal/ml"
    "net"
    "time"
)

func main() {
    // Endereço da Nave-Mãe
    serverAddr, err := net.ResolveUDPAddr("udp", "127.0.0.1:9999")
    if err != nil {
        panic(err)
    }

    // Socket UDP
    conn, err := net.DialUDP("udp", nil, serverAddr)
    if err != nil {
        panic(err)
    }
    defer conn.Close()

    fmt.Println("🚀 Rover iniciado — a pedir missão à Nave-Mãe...")

    seq := uint16(0)
    req := ml.Packet{
        MsgType:  ml.MSG_REQUEST,
        SeqNum:   seq,
        AckNum:   0,
        Checksum: 0,
        Payload:  []byte("REQUEST MISSION"),
    }

    
    req.Checksum = ml.Checksum(req.Payload)

    conn.Write(req.ToBytes())
    fmt.Println("📡 REQUEST enviado.")

    buf := make([]byte, 1024)
    conn.SetReadDeadline(time.Now().Add(5 * time.Second))

    n, _, err := conn.ReadFromUDP(buf)
    if err != nil {
        fmt.Println("❌ Timeout ou erro a receber resposta:", err)
        return
    }

    resp := ml.FromBytes(buf[:n])
    if resp.MsgType == ml.MSG_MISSION {
        fmt.Println("✅ Missão recebida:", ml.DataFromBytes(resp.Payload).String())
    } else {
        fmt.Println("❌ Mensagem inesperada:", resp.MsgType)
    }
}
