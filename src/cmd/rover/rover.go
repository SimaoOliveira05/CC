package main

import (
	"fmt"
	"net"
	"src/config"
	"src/internal/ml"
    "context"
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
    req := ml.Packet{ MsgType: ml.MSG_REQUEST, SeqNum: 1, AckNum: 0, Payload: []byte{} }
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

    // 3) Executar: enviar reports num ticker até Duration
    start := time.Now()
    ticker := time.NewTicker(time.Duration(mission.UpdateFrequency) * time.Second)
    defer ticker.Stop()

    for {
        select {
        case <-ctx.Done():
            return
        case <-ticker.C:
            // Escolher report por mission.TaskType
            var payload []byte
            switch mission.TaskType {
            case ml.TASK_IMAGE_CAPTURE:
                r := ml.ImageReport{ TaskType: ml.TASK_IMAGE_CAPTURE, MissionID: mission.MsgID, ChunkID: 1, Data: []byte("...") }
                payload = r.ToBytes()
            case ml.TASK_SAMPLE_COLLECTION:
                r := ml.SampleReport{
                    TaskType:   ml.TASK_SAMPLE_COLLECTION,
                    MissionID:  mission.MsgID,
                    NumSamples: 2,
                    Components: []ml.Component{
                        {Name: "H2O", Percentage: 60.0},
                        {Name: "SiO2", Percentage: 40.0},
                    },
                }
                payload = r.ToBytes()
            case ml.TASK_ENV_ANALYSIS:
                r := ml.EnvReport{ TaskType: ml.TASK_ENV_ANALYSIS, MissionID: mission.MsgID, Temp: 23.5, Oxygen: 20.9 }
                payload = r.ToBytes()
            case ml.TASK_REPAIR_RESCUE:
                r := ml.RepairReport{ TaskType: ml.TASK_REPAIR_RESCUE, MissionID: mission.MsgID, ProblemID: 1, Repairable: true }
                payload = r.ToBytes()
            case ml.TASK_TOPO_MAPPING:
                r := ml.TopoReport{ TaskType: ml.TASK_TOPO_MAPPING, MissionID: mission.MsgID, Latitude: 41.545, Longitude: -8.421, Height: 54.3 }
                payload = r.ToBytes()
            case ml.TASK_INSTALLATION:
                r := ml.InstallReport{ TaskType: ml.TASK_INSTALLATION, MissionID: mission.MsgID, Success: true }
                payload = r.ToBytes()
            default:
                fmt.Println("⚠️ TaskType desconhecido:", mission.TaskType)
                continue
            }

            pkt := ml.Packet{
                MsgType: ml.MSG_REPORT,
                SeqNum:  uint16(time.Since(start)/time.Second) + 2,
                AckNum:  0,
                Payload: payload,
            }
            pkt.Checksum = ml.Checksum(pkt.Payload)
            conn.Write(pkt.ToBytes())
        }

        // 4) Termina quando Duration expirar
        if time.Since(start) >= time.Duration(mission.Duration)*time.Second {
            end := ml.Packet{ MsgType: ml.MSG_MISSION_END, SeqNum: 0, AckNum: 0, Payload: []byte{} }
            end.Checksum = ml.Checksum(end.Payload)
            conn.Write(end.ToBytes())
            fmt.Println("🏁 Missão terminada — MSG_MISSION_END enviado")
            return
        }
    }
}