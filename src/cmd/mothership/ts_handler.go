package main

import (
	"fmt"
	"net"
	"src/internal/ts"
	"time"
)

func (ms *MotherShip) telemetryReceiver(port string) {
	listener, err := net.Listen("tcp", "0.0.0.0:"+port)
	if err != nil {
		fmt.Println("❌ Erro ao iniciar servidor de telemetria:", err)
		return
	}
	defer listener.Close()

	fmt.Println("📡 Servidor de telemetria à escuta na porta", port)

	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("❌ Erro ao aceitar conexão:", err)
			continue
		}
		go ms.handleTelemetryConnection(conn)
	}
}

func (ms *MotherShip) handleTelemetryConnection(conn net.Conn) {
	defer conn.Close()

	const defaultUpdateFreq = 2                              // segundos
	const maxMissed = 3                                      // falhas até declarar inoperacional
	buf := make([]byte, 256)

	var roverID uint8
	updateFreq := defaultUpdateFreq
	missed := 0

	for {
		// Timeout para leitura
		readTimeout := time.Duration(2*updateFreq) * time.Second
		_ = conn.SetReadDeadline(time.Now().Add(readTimeout))

		n, err := conn.Read(buf)
		if err != nil {
			// Pacote falhado
			missed++
			if roverID != 0 {
				ms.handleMissedTelemetry(roverID, missed, maxMissed)
				if missed >= maxMissed {
					return // Rover declarado inoperacional
				}
			}
			continue
		}

		// Recebemos algo → tratar telemetria
		var telemetry ts.TelemetryPacket
		telemetry.Decode(buf[:n])

		roverID = telemetry.RoverID
		missed = 0 // reset

		ms.updateRoverTelemetry(&telemetry)

		// Enviar atualização para WebSocket
		if ms.APIServer != nil {
			if rover := ms.RoverInfo.GetRover(roverID); rover != nil {
				ms.APIServer.PublishUpdate("rover_update", rover)
			}
		}
	}
}

func (ms *MotherShip) handleMissedTelemetry(roverID uint8, missed, maxMissed int) {
	rover := ms.RoverInfo.GetRover(roverID)
	if rover == nil {
		return
	}

	if missed >= maxMissed {
		ms.RoverInfo.UpdateRover(roverID, "Inoperacional", rover.Battery, rover.Speed, rover.Position, missed)
		ms.EventLogger.Log("ERROR", "TS", fmt.Sprintf("Rover %d declarado inoperacional por falta de telemetria", roverID), nil)
		fmt.Printf("❌ Rover %d marcado como inoperacional\n", roverID)
	} else {
		// Atualização parcial sem mexer no resto
		ms.RoverInfo.UpdateRover(roverID, rover.State, rover.Battery, rover.Speed, rover.Position, missed)
	}
}

func (ms *MotherShip) updateRoverTelemetry(t *ts.TelemetryPacket) {
	stateText := "Ocioso"
	if t.State == ts.STATE_IN_MISSION {
		stateText = "Em Missão"
	}

	ms.RoverInfo.UpdateRover(
		t.RoverID,
		stateText,
		t.Battery,
		t.Speed,
		t.Position,
		0,
	)

	//fmt.Println(ms.RoverInfo.String())
}

