package main

import (
	"src/config"
	"src/internal/core"
	"src/internal/devices"
	"src/internal/ml"
	"time"
)

// ExecuteMission processes a single mission: moves to location, performs task, and sends reports
func (rover *Rover) ExecuteMission(mission ml.MissionData) {
	rover.IncrementActiveMission()
	defer rover.DecrementActiveMission()

	rover.Logger.Infof("Mission", "Mission %d received: TaskType=%d", mission.MsgID, mission.TaskType)

	// For image capture tasks, load the image
	if mission.TaskType == ml.TASK_IMAGE_CAPTURE {
		imagePath := "image.jpg" // Default image path
		if err := rover.Devices.Camera.LoadImage(imagePath); err != nil {
			rover.Logger.Errorf("Camera", "Failed to load image %s: %v", imagePath, err)
			return
		}
		rover.Logger.Infof("Camera", "Image loaded successfully: %d chunks", rover.Devices.Camera.GetTotalChunks())
	}

	// Move to mission location
	rover.Logger.Infof("Movement", "Moving to coordinates (%.4f, %.4f)", mission.Coordinate.Latitude, mission.Coordinate.Longitude)
	if err := core.MoveTo(
		&rover.RoverBase.CurrentPos,
		mission.Coordinate,
		rover.Devices.GPS,
		rover.Devices.Battery,
		rover.Logger,
	); err != nil {
		rover.Logger.Errorf("Movement", "Error moving: %v", err)
		return
	}
	rover.Logger.Info("Movement", "Arrived at destination. Starting task", nil)

	deadline := time.NewTimer(time.Duration(mission.Duration) * time.Second)
	defer deadline.Stop()

	batteryCheck := time.NewTicker(config.BATTERY_CHECK_INTERVAL)
	defer batteryCheck.Stop()

	if mission.UpdateFrequency > 0 {
		ticker := time.NewTicker(time.Duration(mission.UpdateFrequency) * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-batteryCheck.C:
				if rover.checkBatteryAndAbort(mission.MsgID) {
					rover.SuspendForLowBattery()
					rover.Logger.Infof("Battery", "Battery recharged. Resuming mission %d", mission.MsgID)
				}
			case <-deadline.C:
				rover.sendReport(mission, true)
				core.ConsumeBattery(rover.Devices.Battery, config.TASK_BATTERY_RATE)
				return
			case <-ticker.C:
				rover.sendReport(mission, false)
			}
		}
	} else {
		for {
			select {
			case <-batteryCheck.C:
				if rover.checkBatteryAndAbort(mission.MsgID) {
					rover.SuspendForLowBattery()
					rover.Logger.Infof("Battery", "Battery recharged. Resuming mission %d", mission.MsgID)
				}
			case <-deadline.C:
				rover.sendReport(mission, true)
				core.ConsumeBattery(rover.Devices.Battery, config.TASK_BATTERY_RATE)
				return
			}
		}
	}
}

// manageMissions handles mission requests and execution flow with priority queue
func (rover *Rover) manageMissions() {
	for {
		// Wait until there are no active missions
		rover.ML.Cond.L.Lock()
		for rover.GetActiveMissions() != 0 {
			rover.ML.Cond.Wait() // Wait until all missions are finished
		}
		rover.ML.Cond.L.Unlock()

		// Check if we need to request new missions
		if rover.isQueueEmpty() {
			rover.Logger.Infof("Mission", "Requesting %d missions from mothership", rover.ML.MissionQueue.BatchSize)
			rover.sendRequest()
			print("")

			// Wait for all missions in the batch to arrive
			for i := uint8(0); i < rover.ML.MissionQueue.BatchSize; i++ {
				received := <-rover.ML.MissionReceivedChan
				if !received {
					rover.Logger.Info("Mission", "No more missions available", nil)
					time.Sleep(config.NO_MISSION_WAIT)
					break
				}
			}
		}

		// Process next mission from queue by priority
		mission, found := rover.dequeueNextMission()
		if found {
			// Execute mission synchronously (not as goroutine) to prevent multiple simultaneous missions
			rover.ExecuteMission(mission)
		} else {
			// No missions in queue, wait a bit before checking again
			time.Sleep(1 * time.Second)
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

// IsSuspended checks if the rover is currently suspended
func (rover *Rover) IsSuspended() bool {
	rover.ML.SuspendMu.Lock()
	defer rover.ML.SuspendMu.Unlock()
	return rover.ML.Suspended
}

// checkBatteryAndAbort checks if battery is critical and returns true if mission should abort
func (rover *Rover) checkBatteryAndAbort(missionID uint16) bool {
	if rover.Devices.Battery.GetLevel() < config.CRITICAL_BATTERY_LEVEL {
		rover.Logger.Warnf("Battery", "Battery critical during mission! Aborting mission %d", missionID)
		return true
	}
	return false
}

// SuspendForLowBattery suspends rover operations and recharges battery
func (rover *Rover) SuspendForLowBattery() {
	// Set suspended state
	rover.ML.SuspendMu.Lock()
	rover.ML.Suspended = true
	rover.ML.SuspendMu.Unlock()

	rover.Logger.Warnf("Battery", "Rover suspended - Battery: %d%%", rover.Devices.Battery.GetLevel())

	// Cast to MockBattery to access charging methods
	mockBattery, ok := rover.Devices.Battery.(*devices.MockBattery)
	if !ok {
		rover.Logger.Error("Battery", "Battery type not supported for charging", nil)
		return
	}

	mockBattery.StartCharging()

	// Recharge until battery reaches target level
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		<-ticker.C
		mockBattery.Recharge()
		currentLevel := mockBattery.GetLevel()

		if currentLevel%10 == 0 { // Log every 10%
			rover.Logger.Infof("Battery", "Charging... Battery: %d%%", currentLevel)
		}

		if currentLevel >= config.TARGET_RECHARGE_LEVEL {
			rover.Logger.Infof("Battery", "Battery recharged to %d%%", currentLevel)
			break
		}
	}

	mockBattery.StopCharging()

	// Resume operations
	rover.ML.SuspendMu.Lock()
	rover.ML.Suspended = false
	rover.ML.SuspendMu.Unlock()
}

// batteryMonitor continuously monitors battery level
func (rover *Rover) batteryMonitor() {
	ticker := time.NewTicker(config.BATTERY_MONITOR_INTERVAL)
	defer ticker.Stop()

	for range ticker.C {
		level := rover.Devices.Battery.GetLevel()

		// Warning at 20%
		if level <= 20 && level > 5 {
			rover.Logger.Warnf("Battery", "Low battery level: %d%%", level)
		}

		// Critical - suspend immediately if not already suspended
		if level <= config.CRITICAL_BATTERY_LEVEL && !rover.IsSuspended() {
			rover.Logger.Errorf("Battery", "Critical battery level: %d%%", level)
			rover.Logger.Warn("Battery", "Suspending all operations for recharge", nil)
			go rover.SuspendForLowBattery()
		}
	}
}

// isQueueEmpty checks if all priority queues are empty
func (rover *Rover) isQueueEmpty() bool {
	rover.ML.MissionQueue.Mu.Lock()
	defer rover.ML.MissionQueue.Mu.Unlock()

	return len(rover.ML.MissionQueue.Priority1) == 0 &&
		len(rover.ML.MissionQueue.Priority2) == 0 &&
		len(rover.ML.MissionQueue.Priority3) == 0
}

// dequeueNextMission gets the next mission from highest priority queue
func (rover *Rover) dequeueNextMission() (ml.MissionData, bool) {
	rover.ML.MissionQueue.Mu.Lock()
	defer rover.ML.MissionQueue.Mu.Unlock()

	// Check Priority 1 first
	if len(rover.ML.MissionQueue.Priority1) > 0 {
		mission := rover.ML.MissionQueue.Priority1[0]
		rover.ML.MissionQueue.Priority1 = rover.ML.MissionQueue.Priority1[1:]
		rover.Logger.Infof("Mission", "Dequeued mission %d from Priority 1 queue", mission.MsgID)
		return mission, true
	}

	// Check Priority 2
	if len(rover.ML.MissionQueue.Priority2) > 0 {
		mission := rover.ML.MissionQueue.Priority2[0]
		rover.ML.MissionQueue.Priority2 = rover.ML.MissionQueue.Priority2[1:]
		rover.Logger.Infof("Mission", "Dequeued mission %d from Priority 2 queue", mission.MsgID)
		return mission, true
	}

	// Check Priority 3
	if len(rover.ML.MissionQueue.Priority3) > 0 {
		mission := rover.ML.MissionQueue.Priority3[0]
		rover.ML.MissionQueue.Priority3 = rover.ML.MissionQueue.Priority3[1:]
		rover.Logger.Infof("Mission", "Dequeued mission %d from Priority 3 queue", mission.MsgID)
		return mission, true
	}

	// No missions available
	return ml.MissionData{}, false
}
