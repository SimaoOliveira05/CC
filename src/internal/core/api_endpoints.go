package core

import (
    "encoding/base64"
    "src/internal/ml"
)

// Register all API endpoints REST for the MotherShip instance.
func (ms *MotherShip) setupAPIEndpoints() {
    // Endpoint: Lists all connected rovers
    ms.APIServer.RegisterEndpoint("/api/rovers", "GET", ms.handleListRovers)

    // Endpoint: Event/log history
    ms.APIServer.RegisterEndpoint("/logs", "GET", ms.handleGetLogs)

    // Endpoint: Lists all missions (with detailed parsing of reports)
    ms.APIServer.RegisterEndpoint("/api/missions", "GET", ms.handleListMissions)
}

// Handler to list all connected rovers.
// Returns an array of RoverTSState structs.
func (ms *MotherShip) handleListRovers() interface{} {
    return ms.RoverInfo.ListRovers()
}

// Handler to get the event/log history.
// Returns an array of events.
func (ms *MotherShip) handleGetLogs() interface{} {
    return ms.EventLogger.GetHistory()
}

// Handler to list all missions, including parsing of reports.
// Returns an array of missions, each with parsed reports and reconstructed image if applicable.
func (ms *MotherShip) handleListMissions() interface{} {
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
        // Adds reconstructed image if applicable
        assembledImage := m.AssembleImage()
        var assembledImageBase64 string
        if len(assembledImage) > 0 {
            assembledImageBase64 = base64.StdEncoding.EncodeToString(assembledImage)
        }
        result = append(result, map[string]interface{}{
            "id":             m.ID,
            "idRover":        m.IDRover,
            "taskType":       m.TaskType,
            "duration":       m.Duration,
            "updateFrequency": m.UpdateFrequency,
            "lastUpdate":     m.LastUpdate,
            "createdAt":      m.CreatedAt,
            "priority":       m.Priority,
            "state":          m.State,
            "coordinate":     m.Coordinate,
            "reports":        parsedReports,
            "assembledImage": assembledImageBase64,
        })
    }
    return result
}