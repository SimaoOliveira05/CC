package main

import (
	"fmt"
	"src/internal/ml"
	pl "src/utils/packetsLogic"
)

// handlePacket processa cada pacote recebido
func (rover *Rover) handlePacket(pkt ml.Packet) {
	// Closure que captura 'rover'
	processor := func(p ml.Packet) {
		switch p.MsgType {
		case ml.MSG_MISSION:
			pl.HandleAck(p, rover.ML.Window)
			rover.processMission(p)
		case ml.MSG_NO_MISSION:
			pl.HandleAck(p, rover.ML.Window)
			rover.ML.MissionReceivedChan <- false
		case ml.MSG_ACK:
			pl.HandleAck(p, rover.ML.Window) // ✅ Usa 'p' (parâmetro da closure)
		default:
			fmt.Printf("⚠️ Tipo de pacote desconhecido: %d\n", p.MsgType)
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
		pkt.MsgType == ml.MSG_ACK, // ✅ Skip ordering para ACKs
		true,
	)
}

// processMission extrai e processa a missão
func (rover *Rover) processMission(pkt ml.Packet) {
	rover.ML.MissionReceivedChan <- true
	var mission ml.MissionData
	mission = mission.Decode(pkt.Payload)
	go rover.generate(mission)
}

func (rover *Rover) receiver() {
	buf := make([]byte, 2048)
	// Loop de recepção
	for {
		n, _, err := rover.MLConn.Conn.ReadFromUDP(buf)
		if err != nil {
			fmt.Println("Erro ao ler pacote UDP:", err)
			continue
		}

		// Constrói o pacote a partir dos bytes recebidos e trata-o
		var pkt ml.Packet
		pkt.Decode(buf[:n])
		rover.handlePacket(pkt)
	}
}

// sendReport serializa e envia um report para a mothership
func (rover *Rover) sendReport(mission ml.MissionData, final bool) {
	payload := buildReportPayload(mission, final)
	if payload == nil {
		return
	}

	rover.ML.SeqNum++
	pkt := ml.Packet{
		RoverId:  rover.ID,
		MsgType:  ml.MSG_REPORT,
		SeqNum:   uint16(rover.ML.SeqNum),
		AckNum:   0,
		Checksum: 0,
		Payload:  payload,
	}

	pl.PacketManager(rover.MLConn.Conn, rover.MLConn.Addr, pkt, rover.ML.Window)
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

	pl.PacketManager(rover.MLConn.Conn, rover.MLConn.Addr, req, rover.ML.Window)
}

// buildReportPayload cria o payload correto conforme o TaskType
func buildReportPayload(mission ml.MissionData, final bool) []byte {
	var payload []byte
	switch mission.TaskType {
	case ml.TASK_IMAGE_CAPTURE:
		r := ml.ImageReport{TaskType: ml.TASK_IMAGE_CAPTURE, MissionID: mission.MsgID, ChunkID: 1, Data: []byte("..."), IsLastReport: final}
		payload = r.Encode()
	case ml.TASK_SAMPLE_COLLECTION:
		r := ml.SampleReport{
			TaskType:   ml.TASK_SAMPLE_COLLECTION,
			MissionID:  mission.MsgID,
			NumSamples: 2,
			Components: []ml.Component{
				{Name: "H2O", Percentage: 60.0},
				{Name: "SiO2", Percentage: 40.0},
			}, IsLastReport: final,
		}

		payload = r.Encode()
	case ml.TASK_ENV_ANALYSIS:
		r := ml.EnvReport{TaskType: ml.TASK_ENV_ANALYSIS, MissionID: mission.MsgID, Temp: 23.5, Oxygen: 20.9, IsLastReport: final}
		payload = r.Encode()
	case ml.TASK_REPAIR_RESCUE:
		r := ml.RepairReport{TaskType: ml.TASK_REPAIR_RESCUE, MissionID: mission.MsgID, ProblemID: 1, Repairable: true, IsLastReport: final}
		payload = r.Encode()
	case ml.TASK_TOPO_MAPPING:
		r := ml.TopoReport{TaskType: ml.TASK_TOPO_MAPPING, MissionID: mission.MsgID, Latitude: 41.545, Longitude: -8.421, Height: 54.3, IsLastReport: final}
		payload = r.Encode()
	case ml.TASK_INSTALLATION:
		r := ml.InstallReport{TaskType: ml.TASK_INSTALLATION, MissionID: mission.MsgID, Success: true, IsLastReport: final}
		payload = r.Encode()
	default:
		fmt.Printf("⚠️ TaskType desconhecido: %d\n", mission.TaskType)
		return nil
	}
	return payload
}
