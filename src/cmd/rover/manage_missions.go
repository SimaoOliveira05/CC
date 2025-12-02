package main

import (
	"fmt"
	"time"
	"src/internal/ml"
	"src/internal/core"
)

// ExecuteMission processes a single mission: moves to location, performs task, and sends reports
func (rover *Rover) ExecuteMission(mission ml.MissionData) {
	// Increment active missions counter
	rover.IncrementActiveMission()
	defer rover.DecrementActiveMission()

	fmt.Printf("🎯 Mission %d received: TaskType=%d\n", mission.MsgID, mission.TaskType)

	// Move to mission location
	fmt.Printf("🚀 Moving to coordinates (%.4f, %.4f)\n", mission.Coordinate.Latitude, mission.Coordinate.Longitude)
	if err := core.MoveTo(
		&rover.RoverBase.CurrentPos,
		mission.Coordinate,
		rover.Devices.GPS,
		rover.Devices.Battery,
	); err != nil {
		fmt.Printf("❌ Error moving: %v\n", err)
		return
	}
	fmt.Printf("✅ Arrived at destination. Starting task...\n")

	// Execute the task with timer
	deadline := time.NewTimer(time.Duration(mission.Duration) * time.Second)
	defer deadline.Stop()

	if mission.UpdateFrequency > 0 {
		// Periodic mode: send reports every UpdateFrequency
		ticker := time.NewTicker(time.Duration(mission.UpdateFrequency) * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-deadline.C:
				// Ends when Duration expires
				rover.sendReport(mission, true)
				return
			case <-ticker.C:
				// Send periodic report
				rover.sendReport(mission, false)
			}
		}
	} else {
		// No updates mode: just wait for Duration and send a final report
		<-deadline.C
		// Ends when Duration expires
		rover.sendReport(mission, true)
	}

	// Consume battery for task execution
	core.ConsumeBattery(rover.Devices.Battery, core.TaskBatteryRate)
}

// manageMissions handles mission requests and execution flow
func (rover *Rover) manageMissions() {
	for {
		// Wait until there are no active missions
		rover.ML.Cond.L.Lock()
		for rover.GetActiveMissions() != 0 {
			rover.ML.Cond.Wait() // Wait until all missions are finished
		}
		rover.ML.Cond.L.Unlock()

		// If not waiting for missions, request new missions
		if !rover.ML.Waiting {
			rover.sendRequest()
			print("")
			received := <-rover.ML.MissionReceivedChan
			if received { // Mothership sent missions
				rover.ML.Waiting = true
			} else { // Mothership has no missions to send, wait 5 seconds before requesting again
				fmt.Println("🚫 No missions available.")
				time.Sleep(5 * time.Second)
			}
		}
	}
}

// IncrementActiveMission increments the active missions counter
func (rover *Rover) IncrementActiveMission() {
	rover.ML.CondMu.Lock()
	defer rover.ML.CondMu.Unlock()
	rover.ML.ActiveMissions++
}

// GetActiveMissions returns the number of active missions
func (rover *Rover) GetActiveMissions() uint8 {
	rover.ML.CondMu.Lock()
	defer rover.ML.CondMu.Unlock()
	return rover.ML.ActiveMissions
}

// DecrementActiveMission decrements the active missions counter
func (rover *Rover) DecrementActiveMission() {
	rover.ML.CondMu.Lock()
	defer rover.ML.CondMu.Unlock()
	if rover.ML.ActiveMissions > 0 {
		rover.ML.ActiveMissions--
		if rover.ML.ActiveMissions == 0 {
			rover.ML.Waiting = false
			rover.ML.Cond.L.Lock()
			rover.ML.Cond.Signal()
			rover.ML.Cond.L.Unlock()
		}
	}
}
