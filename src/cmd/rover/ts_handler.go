package main

import (
	"net"
	"src/internal/ts"
	"time"
)

// telemetrySender periodically sends telemetry data to the mothership
func (rover *Rover) telemetrySender(telemetryAddr string) {
	var conn net.Conn
	var err error

	// Function to establish/re-establish connection
	connect := func() bool {
		if conn != nil {
			conn.Close()
		}
		conn, err = net.Dial("tcp", telemetryAddr)
		if err != nil {
			rover.Logger.Errorf("Telemetry", "Error connecting to telemetry server: %v", err)
			return false
		}
		rover.Logger.Info("Telemetry", "Connected to telemetry server", nil)
		return true
	}

	// Initial connection
	if !connect() {
		// Wait and retry initial connection
		time.Sleep(5 * time.Second)
		if !connect() {
			rover.Logger.Error("Telemetry", "Failed to establish initial connection, will retry in loop", nil)
		}
	}
	defer func() {
		if conn != nil {
			conn.Close()
		}
	}()

	// Set up ticker for periodic sending
	ticker := time.NewTicker(time.Duration(rover.TS.UpdateFrequency) * time.Second)
	defer ticker.Stop()

	// Telemetry sending loop
	for range ticker.C {
		state := ts.STATE_IDLE
		// Determine rover state
		if rover.ML.ActiveMissions > 0 {
			state = ts.STATE_IN_MISSION
		}

		// Get queue counts
		rover.ML.MissionQueue.Mu.Lock()
		queueP1 := uint8(len(rover.ML.MissionQueue.Priority1))
		queueP2 := uint8(len(rover.ML.MissionQueue.Priority2))
		queueP3 := uint8(len(rover.ML.MissionQueue.Priority3))
		rover.ML.MissionQueue.Mu.Unlock()

		// Generate and send telemetry data
		telemetry := ts.GenerateTelemetry(rover.ID,
			uint8(state),
			rover.CurrentPos,
			rover.Devices.Battery.GetLevel(),
			rover.Devices.GPS.GetSpeed(),
			queueP1,
			queueP2,
			queueP3)

		// Encode telemetry data
		data := telemetry.Encode()

		// Try to send telemetry data
		if conn == nil {
			// No connection, try to reconnect
			if !connect() {
				continue // Skip this tick, try again next time
			}
		}

		_, err := conn.Write(data)
		if err != nil {
			rover.Logger.Errorf("Telemetry", "Error sending telemetry: %v", err)
			// Try to reconnect
			connect()
			continue // Skip this tick, try again next time
		}

		rover.Logger.Infof("Telemetry", "Telemetry sent: Position=(%.6f, %.6f), Speed=%.2f, State=%d, Battery=%d%%",
			telemetry.Position.Latitude,
			telemetry.Position.Longitude,
			telemetry.Speed,
			telemetry.State,
			telemetry.Battery,
		)
	}
}
