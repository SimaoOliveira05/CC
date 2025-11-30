package core

import (
	"encoding/base64"
	"encoding/json"
	"net"

	"src/internal/api"
	"src/internal/ml"
	"src/internal/ts"
	pl "src/utils/packetsLogic"
	"sync"
	el "src/internal/eventLogger"

	"fmt"
	"os"
)

type RoverState struct {
	Addr             *net.UDPAddr
	SeqNum           uint16
	ExpectedSeq      uint16
	Buffer           map[uint16]ml.Packet
	WindowLock       sync.Mutex
	Window           *pl.Window // Janela deslizante específica deste rover
	NumberOfMissions uint8
}

type MotherShip struct {
	Conn           *net.UDPConn
	Rovers         map[uint8]*RoverState // key: IP (ou ID do rover)
	MissionManager *ml.MissionManager
	MissionQueue   chan ml.MissionState
	Mu             sync.Mutex
	RoverInfo      *ts.RoverManager
	EventLogger    *el.EventLogger
	APIServer      *api.APIServer // ✅ Campo para o API Server
}

// Construtor
func NewMotherShip() *MotherShip {
	ms := &MotherShip{
		Rovers:         make(map[uint8]*RoverState),
		MissionManager: ml.NewMissionManager(),
		MissionQueue:   make(chan ml.MissionState, 100),
		Mu:             sync.Mutex{},
		RoverInfo:      ts.NewRoverManager(),
		APIServer:      api.NewAPIServer(),
	}

	err := loadMissionsFromJSON("missions.json", ms.MissionQueue)
	if err != nil {
		fmt.Printf("erro ao carregar missões iniciais: %v\n", err)
		return nil
	}

	ms.EventLogger = el.NewEventLogger(1000, ms.APIServer)

	// Configura os endpoints com os dados da mothership
	ms.setupAPIEndpoints()

	return ms
}

// loadMissionsFromJSON lê missões de um ficheiro JSON e coloca-as na missionQueue
func loadMissionsFromJSON(filename string, queue chan ml.MissionState) error {
	file, err := os.Open(filename)
	if err != nil {
		return fmt.Errorf("erro ao abrir ficheiro: %v", err)
	}
	defer file.Close()

	data, err := os.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("erro ao ler ficheiro: %v", err)
	}

	var missions []ml.MissionState
	if err := json.Unmarshal(data, &missions); err != nil {
		return fmt.Errorf("erro ao fazer unmarshal do JSON: %v", err)
	}

	// Assign incremental IDs to missions
	for i := range missions {
		missions[i].ID = uint16(i + 1) // IDs start from 1
		queue <- missions[i]
	}

	fmt.Printf("📋 %d missões enfileiradas\n", len(missions))
	return nil
}

// setupAPIEndpoints configura todos os endpoints REST da API
func (ms *MotherShip) setupAPIEndpoints() {
	// Endpoint: Lista todos os rovers
	ms.APIServer.RegisterEndpoint("/api/rovers", "GET", func() interface{} {
		return ms.RoverInfo.ListRovers()
	})

	ms.APIServer.RegisterEndpoint("/logs", "GET", func() interface{} {
		return ms.EventLogger.GetHistory()
	})

	// Endpoint: Lista todas as missões
	ms.APIServer.RegisterEndpoint("/api/missions", "GET", func() interface{} {
		missions := ms.MissionManager.ListMissions()
		var result []map[string]interface{}
		for _, m := range missions {
			var parsedReports []interface{}
			for _, rep := range m.Report {
				switch rep.Header.TaskType {
				case ml.TASK_IMAGE_CAPTURE:
					var img ml.ImageReportData
					img.DecodePayload(rep.Payload)
					parsedReports = append(parsedReports, map[string]interface{}{
						"taskType":     rep.Header.TaskType,
						"missionId":    rep.Header.MissionID,
						"chunkId":      img.ChunkID,
						"data":         img.Data,
						"isLastReport": rep.Header.IsLastReport,
					})
				case ml.TASK_SAMPLE_COLLECTION:
					var sample ml.SampleReportData
					sample.DecodePayload(rep.Payload)
					comps := make([]map[string]interface{}, len(sample.Components))
					for i, c := range sample.Components {
						comps[i] = map[string]interface{}{
							"name":       c.Name,
							"percentage": c.Percentage,
						}
					}
					parsedReports = append(parsedReports, map[string]interface{}{
						"taskType":     rep.Header.TaskType,
						"missionId":    rep.Header.MissionID,
						"numSamples":   len(sample.Components),
						"components":   comps,
						"isLastReport": rep.Header.IsLastReport,
					})
				case ml.TASK_ENV_ANALYSIS:
					var env ml.EnvReportData
					env.DecodePayload(rep.Payload)
					parsedReports = append(parsedReports, map[string]interface{}{
						"taskType":     rep.Header.TaskType,
						"missionId":    rep.Header.MissionID,
						"temp":         env.Temp,
						"oxygen":       env.Oxygen,
						"pressure":     env.Pressure,
						"humidity":     env.Humidity,
						"windSpeed":    env.WindSpeed,
						"radiation":    env.Radiation,
						"isLastReport": rep.Header.IsLastReport,
					})
				case ml.TASK_REPAIR_RESCUE:
					var repair ml.RepairReportData
					repair.DecodePayload(rep.Payload)
					parsedReports = append(parsedReports, map[string]interface{}{
						"taskType":     rep.Header.TaskType,
						"missionId":    rep.Header.MissionID,
						"problemId":    repair.ProblemID,
						"repairable":   repair.Repairable,
						"isLastReport": rep.Header.IsLastReport,
					})
				case ml.TASK_TOPO_MAPPING:
					var topo ml.TopoReportData
					topo.DecodePayload(rep.Payload)
					parsedReports = append(parsedReports, map[string]interface{}{
						"taskType":     rep.Header.TaskType,
						"missionId":    rep.Header.MissionID,
						"latitude":     topo.Latitude,
						"longitude":    topo.Longitude,
						"height":       topo.Height,
						"isLastReport": rep.Header.IsLastReport,
					})
				case ml.TASK_INSTALLATION:
					var inst ml.InstallReportData
					inst.DecodePayload(rep.Payload)
					parsedReports = append(parsedReports, map[string]interface{}{
						"taskType":     rep.Header.TaskType,
						"missionId":    rep.Header.MissionID,
						"success":      inst.Success,
						"isLastReport": rep.Header.IsLastReport,
					})
				}
			}
			// If mission completed and has image chunks, assemble and include base64 image
			assembled := m.AssembleImage()
			entry := map[string]interface{}{
				"id":         m.ID,
				"state":      m.State,
				"idRover":    m.IDRover,
				"reports":    parsedReports,
				"taskType":   m.TaskType,
				"coordinate": m.Coordinate,
			}
			if len(assembled) > 0 {
				b64 := base64.StdEncoding.EncodeToString(assembled)
				// Attach assembled image to the last image report if present
				for i, pr := range parsedReports {
					if prmap, ok := pr.(map[string]interface{}); ok {
						// Extract numeric taskType safely from interface{}
						var tt uint8
						switch v := prmap["taskType"].(type) {
						case uint8:
							tt = v
						case uint16:
							tt = uint8(v)
						case int:
							tt = uint8(v)
						case float64:
							tt = uint8(v)
						default:
							tt = 255
						}
						if tt == ml.TASK_IMAGE_CAPTURE {
							if isLast, ok2 := prmap["isLastReport"].(bool); ok2 && isLast {
								prmap["assembledImage"] = b64
								parsedReports[i] = prmap
								break
							}
						}
					}
				}
				entry["assembledImage"] = b64
			}
			result = append(result, entry)
		}
		return result
	})

}
