package core

import (
	"fmt"
	"math"
	"src/config"
	"src/internal/devices"
	"src/utils"
	"time"
)

// CalculateDistance calculates the Euclidean distance between two coordinates
// Coordinates are in [-1,1] representing a normalized Cartesian plane
func CalculateDistance(from, to utils.Coordinate) float64 {
	deltaLat := to.Latitude - from.Latitude
	deltaLon := to.Longitude - from.Longitude

	return math.Sqrt(deltaLat*deltaLat + deltaLon*deltaLon)
}

// MoveTo moves the rover to target coordinates, updating GPS and consuming battery
func MoveTo(
	currentPos *utils.Coordinate,
	target utils.Coordinate,
	gps devices.GPS,
	battery devices.Battery,
) error {
	// Calculate distance to target
	distance := CalculateDistance(*currentPos, target)

	if distance < config.ARRIVAL_THRESHOLD { // Already at the destination
		fmt.Println("✅ Already at the destination")
		return nil
	}

	fmt.Printf("🚀 Moving %.4f units to (%.6f, %.6f)...\n",
		distance, target.Latitude, target.Longitude)

	// Calculate travel time
	travelTime := distance / config.MAX_SPEED
	fmt.Printf("⏳ Estimated travel time: %.2fs\n", travelTime)

	startTime := time.Now()
	stepCount := 0

	for {
		stepCount++
		distanceToTarget := CalculateDistance(*currentPos, target)

		if distanceToTarget < config.ARRIVAL_THRESHOLD { // Arrived
			fmt.Printf("✅ Arrived at destination (remaining distance: %.4f)\n", distanceToTarget)
			break
		}

		// Calculate direction vector (normalized)
		directionLat := (target.Latitude - currentPos.Latitude) / distanceToTarget
		directionLon := (target.Longitude - currentPos.Longitude) / distanceToTarget

		// Move MAX_SPEED units towards target
		newLat := currentPos.Latitude + directionLat*config.MAX_SPEED
		newLon := currentPos.Longitude + directionLon*config.MAX_SPEED
		coords := utils.Coordinate{
			Latitude:  newLat,
			Longitude: newLon,
		}

		// If overshoot, snap to target
		if CalculateDistance(coords, target) > distanceToTarget {
			coords = utils.Coordinate{
				Latitude:  target.Latitude,
				Longitude: target.Longitude,
			}
		}

		*currentPos = coords

		// Update mock GPS
		if mockGPS, ok := gps.(*devices.MockGPS); ok {
			mockGPS.SetPosition(*currentPos)
			mockGPS.SetSpeed(float32(config.MAX_SPEED))
		}

		// Consume battery proportional to distance traveled (MAX_SPEED per step)
		batteryDrain := config.MAX_SPEED * config.MOVEMENT_BATTERY_RATE
		ConsumeBattery(battery, batteryDrain)

		// Log every 10 steps
		if stepCount%10 == 0 {
			fmt.Printf("   Step %d - Remaining distance: %.4f - Position: (%.4f, %.4f)\n",
				stepCount, distanceToTarget, currentPos.Latitude, currentPos.Longitude)
		}

		time.Sleep(1 * time.Second)
	}

	elapsed := time.Since(startTime)
	fmt.Printf("✅ Arrived at destination in %.2fs (estimated %.2fs). Battery: %d%%\n",
		elapsed.Seconds(), travelTime, battery.GetLevel())

	// Stop at the destination
	if mockGPS, ok := gps.(*devices.MockGPS); ok {
		mockGPS.SetSpeed(0)
	}

	return nil
}

// ConsumeBattery reduces the battery level by the specified amount (now float64 for precision)
func ConsumeBattery(battery devices.Battery, amount float64) {
	if mockBattery, ok := battery.(*devices.MockBattery); ok {
		currentLevel := float64(mockBattery.GetLevel())
		newLevel := uint8(math.Max(0, currentLevel-amount))
		mockBattery.SetLevel(newLevel)
	}
}
