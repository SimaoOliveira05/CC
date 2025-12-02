package main

import (
	"fmt"
	"math/rand"
	"src/internal/ml"
	pl "src/utils/packetsLogic"
)

// handlePacket processes each packet on a separate goroutine
func (rover *Rover) handlePacket(pkt ml.Packet) {
	// Closure that captures 'rover'
	processor := func(p ml.Packet) {
		switch p.MsgType {
		case ml.MSG_MISSION:
			pl.HandleAck(p, rover.ML.Window)
			rover.processMission(p)
		case ml.MSG_NO_MISSION:
			pl.HandleAck(p, rover.ML.Window)
			rover.ML.MissionReceivedChan <- false
		case ml.MSG_ACK:
			pl.HandleAck(p, rover.ML.Window) // Uses 'p' (closure parameter)
		default:
			fmt.Printf("⚠️ Unknown packet type: %d\n", p.MsgType)
		}
	}

	pl.HandleOrderedPacket(
		pkt,
		&rover.ML.ExpectedSeq,
		rover.ML.Buffer,
		&rover.ML.CondMu,
		rover.MLConn.Conn,
		rover.MLConn.Addr,
		rover.ML.Window,
		rover.ID,
		processor,
		pkt.MsgType == ml.MSG_ACK, // Skip ordering for ACKs
		true,
		func(level, msg string, meta any) {
        fmt.Printf("[%s] %s %+v\n", level, msg, meta)
    	},
	)
}

// processMission extracts and processes the mission
func (rover *Rover) processMission(pkt ml.Packet) {
	rover.ML.MissionReceivedChan <- true
	var mission ml.MissionData
	mission = mission.Decode(pkt.Payload)
	go rover.ExecuteMission(mission)
}

// receiver continuously reads UDP packets
func (rover *Rover) receiver() {
	buf := make([]byte, 2048)
	// Reception loop
	for {
		n, _, err := rover.MLConn.Conn.ReadFromUDP(buf)
		if err != nil {
			fmt.Println("Error reading UDP packet:", err)
			continue
		}

		// Constructs the packet from received bytes and processes it
		var pkt ml.Packet
		pkt.Decode(buf[:n])
		rover.handlePacket(pkt)
	}
}

// sendReport serializes and sends a report to the mothership
func (rover *Rover) sendReport(mission ml.MissionData, final bool) {
	payload := rover.buildReportPayload(mission, final)
	if payload == nil {
		return
	}

	// Increment sequence number and create packet
	rover.ML.SeqNum++
	pkt := ml.Packet{
		RoverId:  rover.ID,
		MsgType:  ml.MSG_REPORT,
		SeqNum:   uint16(rover.ML.SeqNum),
		AckNum:   0,
		Checksum: 0,
		Payload:  payload,
	}

	// Use PacketManager to send the report packet
	pl.PacketManager(rover.MLConn.Conn, 
					rover.MLConn.Addr, 
					pkt, 
					rover.ML.Window,
					func(level, msg string, meta any) {
						fmt.Printf("[%s] %s %+v\n", level, msg, meta)
					})
}

// sendRequest envia um pedido de missão para a mothership
func (rover *Rover) sendRequest() {

	rover.ML.SeqNum++

	req := ml.Packet{
		RoverId:  rover.ID,
		MsgType:  ml.MSG_REQUEST,
		SeqNum:   uint16(rover.ML.SeqNum),
		AckNum:   0,
		Checksum: 0,
		Payload:  []byte{},
	}

	pl.PacketManager(rover.MLConn.Conn, 
					rover.MLConn.Addr, 
					req, 
					rover.ML.Window,	
					func(level, msg string, meta any) {
						fmt.Printf("[%s] %s %+v\n", level, msg, meta)
					})
}

// buildReportPayload creates a generic report header
func (rover *Rover) buildReportPayload(mission ml.MissionData, final bool) []byte {
	header := ml.ReportHeader{
		TaskType:     mission.TaskType,
		MissionID:    mission.MsgID,
		IsLastReport: final,
	}

	payload := rover.buildPayload(mission)

	report := ml.Report{
		Header:  header,
		Payload: payload,
	}

	return report.Encode()
}


// buildPayload creates the payload for different mission types
func (rover *Rover) buildPayload(mission ml.MissionData) []byte {
	var payload []byte
	switch mission.TaskType {
	case ml.TASK_IMAGE_CAPTURE:
		img := ml.ImageReportData{
			ChunkID: 1,
			Data:    rover.Devices.Camera.ReadImageChunk(),
		}
		payload = img.EncodePayload()
	case ml.TASK_SAMPLE_COLLECTION:
		comps := rover.Devices.ChemicalAnalyzer.Analyze()
		// Converter para ml.Component se necessário
		mlComps := make([]ml.Component, len(comps))
		for i, c := range comps {
			mlComps[i] = ml.Component{Name: c.Name, Percentage: c.Percentage}
		}
		sample := ml.SampleReportData{
			Components: mlComps,
		}
		payload = sample.EncodePayload()
	case ml.TASK_ENV_ANALYSIS:
		temp := rover.Devices.Thermometer.GetTemperature()
		oxygen := rover.Devices.Thermometer.GetOxygen()
		pressure := rover.Devices.Thermometer.GetPressure()
		humidity := rover.Devices.Thermometer.GetHumidity()
		wind := rover.Devices.Thermometer.GetWindSpeed()
		radiation := rover.Devices.Thermometer.GetRadiation()
		env := ml.EnvReportData{
			Temp:      temp,
			Oxygen:    oxygen,
			Pressure:  pressure,
			Humidity:  humidity,
			WindSpeed: wind,
			Radiation: radiation,
		}
		payload = env.EncodePayload()
	case ml.TASK_REPAIR_RESCUE:
		repair := ml.RepairReportData{
			ProblemID:  uint8(rover.ID),
			Repairable: true,
		}
		payload = repair.EncodePayload()
	case ml.TASK_TOPO_MAPPING:
		pos := rover.Devices.GPS.GetPosition()
		topo := ml.TopoReportData{
			Latitude:  pos.Latitude,
			Longitude: pos.Longitude,
			Height:    rover.Devices.GPS.GetAltitude() + rand.Float32()*10.0,
		}
		payload = topo.EncodePayload()
	case ml.TASK_INSTALLATION:
		// Installation can fail depending on battery level and randomness
		battery := rover.Devices.Battery.GetLevel()
		successChance := 0.9
		if battery < 20 {
			successChance = 0.7
		} else if battery < 50 {
			successChance = 0.8
		}
		success := rand.Float32() < float32(successChance)
		inst := ml.InstallReportData{
			Success: success,
		}
		payload = inst.EncodePayload()
	default:
		payload = []byte("generic report")
	}
	return payload
}
