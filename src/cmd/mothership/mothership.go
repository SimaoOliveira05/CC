package main

import (
    "fmt"
    "src/internal/ml"
    "net"
	"src/utils"
)

func main() {
    addr, _ := net.ResolveUDPAddr("udp", ":9999")
    conn, _ := net.ListenUDP("udp", addr)
    defer conn.Close()

    fmt.Println("🛰️ Nave-Mãe à escuta...")

    buf := make([]byte, 1024)

    for {
        n, clientAddr, _ := conn.ReadFromUDP(buf)
        p := ml.FromBytes(buf[:n])
        fmt.Println("📨 Recebido pacote do tipo:", p.MsgType, "de", clientAddr)

        if p.MsgType == ml.MSG_REQUEST {
			payload := ml.Data{
				MsgID: 33,
				Coordinate: utils.Coordinate{Latitude: 32, Longitude: 25},
				TaskType: ml.Rescue,
				Duration: 300,
				UpdateFrequency: 20,
				Priority: 0,
			}

            mission := ml.Packet{
                MsgType:  ml.MSG_MISSION,
                SeqNum:   0,
                AckNum:   p.SeqNum + 1,
                Payload:  payload.ToBytes(),
            }
            mission.Checksum = ml.Checksum(mission.Payload)
            conn.WriteToUDP(mission.ToBytes(), clientAddr)
            fmt.Println("✅ Missão enviada.")
        }
    }
}
