package main

import (
	"fmt"
	"net"
	"src/internal/ts"
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
	buf := make([]byte, 256)

	for {
		n, err := conn.Read(buf)
		if err != nil {
			return
		}

		telemetry, err := ts.FromBytes(buf[:n])
		if err != nil {
			fmt.Println("❌ Erro ao deserializar telemetria:", err)
			continue
		}

		var estado string
		if telemetry.State == ts.STATE_IN_MISSION {
			estado = "Em Missão"
		} else {
			estado = "Ocioso"
		}

		ms.RoverInfo.UpdateRover(telemetry.RoverID, estado, telemetry.Battery, telemetry.Speed, telemetry.Position)

		fmt.Println(ms.RoverInfo.String())

		// 🔥 Publish real-time update via WebSocket
		if ms.APIServer != nil {
			rover := ms.RoverInfo.GetRover(telemetry.RoverID)
			if rover != nil {
				ms.APIServer.PublishUpdate("rover_update", rover)
			}
		}
	}
}
