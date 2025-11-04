package main

import (
	"context"
	"fmt"
	"net"
	"src/config"
	"src/internal/ml"
	"time"
)

func main() {
	// Inicializa configuração (isRover = true)
	config.InitConfig(true)
	config.PrintConfig()

	runMissionUDP(context.Background())

	fmt.Printf("🤖 Rover conectado à Mothership em %s\n", config.GetMotherIP())
}

func runMissionUDP(ctx context.Context) {

	mothershipAddr := config.GetMotherIP()
	
	addr, err := net.ResolveUDPAddr("udp", mothershipAddr+":9999")
	if err != nil {
		fmt.Println("❌ Erro ao resolver endereço:", err)
		return
	}

	// Conecta à mothership
	conn, err := net.DialUDP("udp", nil, addr)
	if err != nil {
		fmt.Println("❌ Erro ao conectar:", err)
		return
	}
	defer conn.Close()

	// 1) Pedir missão
	req := ml.Packet{MsgType: ml.MSG_REQUEST, SeqNum: 1, AckNum: 0, Payload: []byte{}}
	req.Checksum = ml.Checksum(req.Payload)
	conn.Write(req.ToBytes())

	// 2) Esperar resposta de missão

	buf := make([]byte, 2048)
	_ = conn.SetReadDeadline(time.Now().Add(5 * time.Second))
	n, _, err := conn.ReadFromUDP(buf)
	if err != nil {
		fmt.Println("❌ Timeout missão:", err)
		return
	}

	resp := ml.FromBytes(buf[:n])
	if resp.MsgType != ml.MSG_MISSION {
		fmt.Println("⚠️ Mensagem inesperada:", resp.MsgType)
		return
	}
	mission := ml.DataFromBytes(resp.Payload)
	fmt.Println("📝 Missão recebida:", mission.String())

	// 3) Executar: enviar reports (periódicos ou apenas final)

	deadline := time.NewTimer(time.Duration(mission.Duration) * time.Second)
	defer deadline.Stop()

	if mission.UpdateFrequency > 0 {
		// Modo periódico: enviar reports a cada UpdateFrequency
		ticker := time.NewTicker(time.Duration(mission.UpdateFrequency) * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-deadline.C:
				// Termina quando Duration expirar
				sendFinalReport(conn, mission)
				return
			case <-ticker.C:
				// Enviar report periódico
				sendReport(conn, mission,false)
			}
		}
	} else {
		// Modo sem updates: apenas espera Duration e envia um report final
		select {
		case <-ctx.Done():
			return
		case <-deadline.C:
			// Termina quando Duration expirar
			sendFinalReport(conn, mission)
			return
		}
	}
}

// sendReport serializa e envia um report para a mothership
func sendReport(conn *net.UDPConn, mission ml.MissionData, final bool) {
	payload := buildReportPayload(mission, final)
	if payload == nil {
		return
	}

	pkt := ml.Packet{
		MsgType: ml.MSG_REPORT,
		SeqNum:  uint16(time.Now().Unix() & 0xFFFF),
		AckNum:  0,
		Payload: payload,
	}
	pkt.Checksum = ml.Checksum(pkt.Payload)
	conn.Write(pkt.ToBytes())
	fmt.Printf("📤 Report enviado (Missão %d)\n", mission.MsgID)
}

// sendFinalReport envia o report final antes de terminar a missão
func sendFinalReport(conn *net.UDPConn, mission ml.MissionData) {
	fmt.Println("📤 Enviando report final...")
	sendReport(conn, mission,true)
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
