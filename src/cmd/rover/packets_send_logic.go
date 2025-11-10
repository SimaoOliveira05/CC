package main

import (
	"fmt"
	"src/internal/ml"
	"time"
)

// sendPacket envia um pacote e gerencia retransmissões até receber o ACK
func sendPacket(pkt ml.Packet, window *Window, c *RoverMlConection) {
    window.mu.Lock()
	ch := make(chan int8, 1)
    window.window[uint32(pkt.SeqNum)] = ch
    window.mu.Unlock()
    go packetManager(pkt, ch, c)
}

// sendReport serializa e envia um report para a mothership
func sendReport(mission ml.MissionData, final bool, c *RoverMlConection, window *Window, r *Rover) {
	payload := buildReportPayload(mission, final)
	if payload == nil {
		return
	}

	c.seqNum++
	pkt := ml.Packet{
		RoverId: r.id,
		MsgType: ml.MSG_REPORT,
		SeqNum:  uint16(c.seqNum),
		AckNum:  0,
		Checksum: 0,
		Payload: payload,
	}

	sendPacket(pkt, window, c)
}

// sendRequest envia um pedido de missão para a mothership
func sendRequest(c *RoverMlConection, window *Window, r *Rover) {

	c.seqNum++
	req := ml.Packet{
		RoverId: r.id,
		MsgType: ml.MSG_REQUEST,
		SeqNum:  uint16(c.seqNum),
		AckNum:  0,
		Checksum: 0,
		Payload: []byte{},
	}

	sendPacket(req, window, c)
}

// packetManager gerencia o envio e retransmissão de um pacote até receber o ACK
func packetManager(pkt ml.Packet, ch chan int8, c *RoverMlConection) {
    retries := 0 
	maxRetries := 5
	for {
        // Envia o pacote
        _, err := c.conn.Write(pkt.ToBytes())
        if err != nil {
            fmt.Println("Erro ao enviar pacote:", err)
            return
        }
		if(pkt.MsgType == ml.MSG_REQUEST){
			fmt.Printf("📤 Pedido de missão enviado, SeqNum: %d\n", pkt.SeqNum)
		}
		if(pkt.MsgType == ml.MSG_REPORT){
			fmt.Printf("📤 Report enviado, SeqNum: %d\n", pkt.SeqNum)
		}
		if(pkt.MsgType == ml.MSG_ACK){
			fmt.Printf("📤 ACK enviado, SeqNum: %d\n", pkt.SeqNum)
		}

        select {
        case <-ch:
            fmt.Printf("✅ ACK recebido para SeqNum %d\n", pkt.SeqNum)
            return
		case <-time.After(500 * time.Millisecond):
			retries++
			if retries > maxRetries {
				fmt.Printf("❌ Falha ao receber ACK para SeqNum %d após %d tentativas. Abortando...\n", pkt.SeqNum, maxRetries)
				return
			}
			fmt.Printf("⏱️ Timeout esperando ACK para SeqNum %d. Retransmitindo (tentativa %d)...\n", pkt.SeqNum, retries)
        }
    }
}

// buildReportPayload cria o payload correto conforme o TaskType
func buildReportPayload(mission ml.MissionData, final bool) []byte {
	var payload []byte
	switch mission.TaskType {
	case ml.TASK_IMAGE_CAPTURE:
		r := ml.ImageReport{TaskType: ml.TASK_IMAGE_CAPTURE, MissionID: mission.MsgID, ChunkID: 1, Data: []byte("..."), IsLastReport: final}
		payload = r.ToBytes()
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
        
		payload = r.ToBytes()
	case ml.TASK_ENV_ANALYSIS:
		r := ml.EnvReport{TaskType: ml.TASK_ENV_ANALYSIS, MissionID: mission.MsgID, Temp: 23.5, Oxygen: 20.9, IsLastReport: final}
		payload = r.ToBytes()
	case ml.TASK_REPAIR_RESCUE:
		r := ml.RepairReport{TaskType: ml.TASK_REPAIR_RESCUE, MissionID: mission.MsgID, ProblemID: 1, Repairable: true, IsLastReport: final}
		payload = r.ToBytes()
	case ml.TASK_TOPO_MAPPING:
		r := ml.TopoReport{TaskType: ml.TASK_TOPO_MAPPING, MissionID: mission.MsgID, Latitude: 41.545, Longitude: -8.421, Height: 54.3, IsLastReport: final}
		payload = r.ToBytes()
	case ml.TASK_INSTALLATION:
		r := ml.InstallReport{TaskType: ml.TASK_INSTALLATION, MissionID: mission.MsgID, Success: true, IsLastReport: final}
		payload = r.ToBytes()
	default:
		fmt.Printf("⚠️ TaskType desconhecido: %d\n", mission.TaskType)
		return nil
	}
	return payload
}
