package core

import (
	"fmt"
	"math"
	"src/internal/devices"
	"src/utils"
	"time"
)

// Configuração de movimento
const (
	MaxSpeed            = 0.1 // unidades/s no espaço [-1,1]
	MovementBatteryRate = 5.0 // % por unidade de distância
	TaskBatteryRate     = 2.0 // % por tarefa
)

// CalculateDistance calcula distância euclidiana entre duas coordenadas
// Coordenadas estão em [-1,1] representando um plano cartesiano normalizado
func CalculateDistance(from, to utils.Coordinate) float64 {
	deltaLat := to.Latitude - from.Latitude
	deltaLon := to.Longitude - from.Longitude

	// Distância euclidiana simples
	return math.Sqrt(deltaLat*deltaLat + deltaLon*deltaLon)
}

// MoveTo move o rover para coordenadas destino, atualizando GPS e consumindo bateria
func MoveTo(
	currentPos *utils.Coordinate,
	target utils.Coordinate,
	gps devices.GPS,
	battery devices.Battery,
) error {
	distance := CalculateDistance(*currentPos, target)

	if distance < 0.01 { // Já está no destino (menos de 1% do mapa)
		fmt.Println("✅ Já está no destino")
		return nil
	}

	fmt.Printf("🚀 Deslocando %.4f unidades para (%.6f, %.6f)...\n",
		distance, target.Latitude, target.Longitude)

	// Calcula tempo de viagem
	travelTime := distance / MaxSpeed
	steps := int(travelTime) + 1

	// Move em pequenos passos
	fmt.Printf("⏳ Tempo estimado de viagem: %.2fs em %d passos\n", travelTime, steps)
	for i := 0; i < steps; i++ {
		// Interpola posição
		progress := float64(i+1) / float64(steps)
		newPos := utils.Coordinate{
			Latitude: currentPos.Latitude +
				(target.Latitude-currentPos.Latitude)*progress,
			Longitude: currentPos.Longitude +
				(target.Longitude-currentPos.Longitude)*progress,
		}

		*currentPos = newPos

		// Verifica se chegou ao destino (distância < 0.01 unidades)
		remainingDistance := CalculateDistance(*currentPos, target)
		if remainingDistance < 0.01 {
			fmt.Printf("✅ Chegou ao destino antecipadamente (distância restante: %.4f)\n", remainingDistance)
			break
		}

		// Atualiza GPS mock
		if mockGPS, ok := gps.(*devices.MockGPS); ok {
			mockGPS.SetPosition(newPos)
			mockGPS.SetSpeed(MaxSpeed)
		}

		// Consome bateria proporcional à distância percorrida neste step
		stepDistance := distance / float64(steps)
		batteryDrain := uint8(stepDistance * MovementBatteryRate)
		ConsumeBattery(battery, batteryDrain)

		// Log a cada 10 passos
		if i%10 == 0 {
			fmt.Printf("   Passo %d/%d - Distância restante: %.4f - Posição: (%.4f, %.4f)\n",
				i+1, steps, remainingDistance, currentPos.Latitude, currentPos.Longitude)
		}

		time.Sleep(1 * time.Second)
	}

	// Para no destino
	if mockGPS, ok := gps.(*devices.MockGPS); ok {
		mockGPS.SetSpeed(0)
	}

	fmt.Printf("✅ Chegou ao destino. Bateria: %d%%\n", battery.GetLevel())
	return nil
}

// ConsumeBattery reduz o nível de bateria
func ConsumeBattery(battery devices.Battery, amount uint8) {
	if mockBattery, ok := battery.(*devices.MockBattery); ok {
		currentLevel := mockBattery.GetLevel()
		if currentLevel > amount {
			mockBattery.SetLevel(currentLevel - amount)
		} else {
			mockBattery.SetLevel(0)
		}
	}
}
