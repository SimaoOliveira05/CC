package main

import (
    "fmt"
    "net"
    "time"
    "src/internal/ts"
)

func (rv *Rover) telemetrySender(mothershipAddr string) {
    conn, err := net.Dial("tcp", mothershipAddr+":9998")
    if err != nil {
        fmt.Println("❌ Erro ao conectar ao servidor de telemetria:", err)
        return
    }
    defer conn.Close()

    ticker := time.NewTicker(5 * time.Second) // Envia a cada 5 segundos
    defer ticker.Stop()

    for range ticker.C {
        state := ts.STATE_IDLE
        if rv.activeMissions > 0 {
            state = ts.STATE_IN_MISSION
        }

        telemetry := ts.GenerateTelemetry(rv.id, uint8(state))
        data := telemetry.ToBytes()

        _, err := conn.Write(data)
        if err != nil {
            fmt.Println("❌ Erro ao enviar telemetria:", err)
            return
        }
        fmt.Printf("📡 Telemetria enviada: Estado=%d, Bateria=%d%%\n", telemetry.State, telemetry.Battery)
    }
}