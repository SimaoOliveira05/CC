package main

import (
	"fmt"
	"src/internal/ml"
	"time"
)

// sendPacket envia um pacote e gerencia retransmissões até receber o ACK
func (rv *Rover) sendPacket(pkt ml.Packet) {
    rv.window.mu.Lock()
	ch := make(chan int8, 1)
    rv.window.window[uint32(pkt.SeqNum)] = ch
    rv.window.mu.Unlock()
    go rv.packetManager(pkt, ch)
}

// sendReport serializa e envia um report para a mothership
func (rv *Rover) sendReport(mission ml.MissionData, final bool) {
	payload := buildReportPayload(mission, final)
	if payload == nil {
		return
	}

	rv.conn.seqNum++
	pkt := ml.Packet{
		RoverId: rv.id,
		MsgType: ml.MSG_REPORT,
		SeqNum:  uint16(rv.conn.seqNum),
		AckNum:  0,
		Checksum: 0,
		Payload: payload,
	}

	rv.sendPacket(pkt)
}

// sendRequest envia um pedido de missão para a mothership
func (rv *Rover) sendRequest() {

	rv.conn.seqNum++
	req := ml.Packet{
		RoverId: rv.id,
		MsgType: ml.MSG_REQUEST,
		SeqNum:  uint16(rv.conn.seqNum),
		AckNum:  0,
		Checksum: 0,
		Payload: []byte{},
	}

	rv.sendPacket(req)
}


func (rv *Rover) sendAck(ackNum uint16) {
	ackPacket := ml.Packet{
		RoverId: 0,
		MsgType: ml.MSG_ACK,
		SeqNum:  0,
		AckNum:  ackNum + 1,
		Payload: []byte{},
	}
	ackPacket.Checksum = ml.Checksum(ackPacket.Payload)

	if _, err := rv.conn.conn.Write(ackPacket.ToBytes()); err != nil {
		fmt.Println("❌ Erro ao enviar ACK:", err)
		return
	}
	fmt.Printf("📤 ACK enviado, AckNum: %d\n", ackNum)
}


// packetManager gerencia o envio e retransmissão de um pacote até receber o ACK
func (rv *Rover) packetManager(pkt ml.Packet, ch chan int8) {
    retries := 0 
	maxRetries := 5
	for {
        // Envia o pacote
        _, err := rv.conn.conn.Write(pkt.ToBytes())
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
