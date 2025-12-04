package main

import (
	"fmt"
	"net"
	"src/internal/ts"
	"time"
)

// telemetryReceiver starts a TCP server to receive telemetry data from rovers
func (ms *MotherShip) telemetryReceiver(port string) {
	// Start TCP server
	listener, err := net.Listen("tcp", "0.0.0.0:"+port)
	if err != nil {
		fmt.Println("❌ Error starting telemetry server:", err)
		return
	}
	defer listener.Close()

	fmt.Println("📡 Telemetry server listening on port", port)

	// Accept incoming connections
	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("❌ Error accepting connection:", err)
			continue
		}
		go ms.handleTelemetryConnection(conn)
	}
}

// handleTelemetryConnection processes telemetry data from a single rover connection
func (ms *MotherShip) handleTelemetryConnection(conn net.Conn) {
	defer conn.Close()

	const defaultUpdateFreq = 2 // seconds
	const maxMissed = 3         // failures before declaring inoperational

	// Buffer for incoming data
	buf := make([]byte, 256)

	// Telemetry handling loop
	var roverID uint8
	updateFreq := defaultUpdateFreq
	missed := 0

	for {
		// Read timeout
		readTimeout := time.Duration(2*updateFreq) * time.Second
		_ = conn.SetReadDeadline(time.Now().Add(readTimeout))

		n, err := conn.Read(buf)
		if err != nil {
			// Failed packet
			missed++ // increment missed counter
			if roverID != 0 {
				ms.handleMissedTelemetry(roverID, missed, maxMissed)
				if missed >= maxMissed {
					return // Rover declared inoperational
				}
			}
			continue
		}

		// Received something → handle telemetry
		var telemetry ts.TelemetryPacket
		telemetry.Decode(buf[:n])

		roverID = telemetry.RoverID
		missed = 0 // reset

		ms.updateRoverTelemetry(&telemetry)

		// Send update to WebSocket
		if ms.APIServer != nil {
			if rover := ms.RoverInfo.GetRover(roverID); rover != nil {
				ms.APIServer.PublishUpdate("rover_update", rover)
			}
		}
	}
}

// handleMissedTelemetry processes missed telemetry packets for a rover
func (ms *MotherShip) handleMissedTelemetry(roverID uint8, missed, maxMissed int) {
	// Get current rover info
	rover := ms.RoverInfo.GetRover(roverID)
	if rover == nil {
		return
	}

	// Check if rover should be declared inoperational
	if missed >= maxMissed {
		ms.RoverInfo.UpdateRover(roverID, "Inoperational", rover.Battery, rover.Speed, rover.Position, missed, rover.QueuedMissions)
		ms.EventLogger.Log("ERROR", "TS", fmt.Sprintf("Rover %d declared inoperational due to lack of telemetry", roverID), nil)
		fmt.Printf("❌ Rover %d marked as inoperational\n", roverID)
	} else {
		// Partial update without changing the rest
		ms.RoverInfo.UpdateRover(roverID, rover.State, rover.Battery, rover.Speed, rover.Position, missed, rover.QueuedMissions)
	}
}

// updateRoverTelemetry updates the rover information based on received telemetry
func (ms *MotherShip) updateRoverTelemetry(t *ts.TelemetryPacket) {
	stateText := "Idle"
	if t.State == ts.STATE_IN_MISSION {
		stateText = "In Mission"
	}

	queueInfo := ts.QueueInfo{
		Priority1Count: t.QueueP1Count,
		Priority2Count: t.QueueP2Count,
		Priority3Count: t.QueueP3Count,
		Priority1IDs:   []uint16{}, // IDs not sent in telemetry for bandwidth
		Priority2IDs:   []uint16{},
		Priority3IDs:   []uint16{},
	}

	ms.RoverInfo.UpdateRover(
		t.RoverID,
		stateText,
		t.Battery,
		t.Speed,
		t.Position,
		0,
		queueInfo,
	)
}
