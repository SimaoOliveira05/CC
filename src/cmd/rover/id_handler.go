package main

import (
    "fmt"
    "net"
    "time"
)

func requestID(mothershipAddr string) (uint8, uint, error) {
    conn, err := net.Dial("tcp", mothershipAddr+":9997")
    if err != nil {
        return 0, 0, fmt.Errorf("erro ao conectar ao servidor de IDs: %v", err)
    }
    defer conn.Close()

    buf := make([]byte, 2)
    conn.SetReadDeadline(time.Now().Add(3 * time.Second))
    _, err = conn.Read(buf)
    if err != nil {
        return 0, 0, fmt.Errorf("timeout ou erro ao receber ID: %v", err)
    }

    id := buf[0]
    updateFrequency := uint(buf[1])
    fmt.Printf("✅ ID recebido da nave-mãe: %d (updateFrequency=%d)\n", id, updateFrequency)
    return id, updateFrequency, nil
}