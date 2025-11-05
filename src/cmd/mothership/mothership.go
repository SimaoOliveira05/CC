package main

import (
	"fmt"
	"net"
	"src/config"
	"src/internal/ml"
	"src/utils"
	"time"
)

var sent bool = true

func main() {
	config.InitConfig(false)
	config.PrintConfig()

	addr, _ := net.ResolveUDPAddr("udp", config.GetMotherIP()+":9999")
	conn, _ := net.ListenUDP("udp", addr)
	defer conn.Close()

	fmt.Println("🛰️ Nave-Mãe à escuta...")

	// Cria o Mission Manager
	missionManager := ml.NewMissionManager()

	// Goroutine para ler pacotes UDP
	go mlListener(conn, missionManager)

	// Loop infinito
	select {}
}

// mlListener lê continuamente pacotes UDP
func mlListener(conn *net.UDPConn, mm *ml.MissionManager) {
	buf := make([]byte, 1024)

	for {
		// n é o número de bytes lidos
		n, clientAddr, err := conn.ReadFromUDP(buf)
		if err != nil {
			fmt.Println("❌ Erro ao ler UDP:", err)
			continue
		}

		// buf[:n] contém os bytes lidos :n descarta o resto do buffer
		p := ml.FromBytes(buf[:n])
		go handlePacket(p, clientAddr, conn, mm)

		fmt.Println("📨 Recebido pacote do tipo:", p.MsgType, "de", clientAddr)
	}
}

// handlePacket processa cada pacote numa goroutine separada
func handlePacket(p ml.Packet, clientAddr *net.UDPAddr, conn *net.UDPConn, mm *ml.MissionManager) {
	switch p.MsgType {
	case ml.MSG_REQUEST:
		handleMissionRequest(p, clientAddr, conn, mm)
	case ml.MSG_ACK:
		handleACK(p, clientAddr)
	case ml.MSG_REPORT:

		// Envia ACK de volta ao rover
		ackPacket := ml.Packet{
			MsgType: ml.MSG_ACK,
			SeqNum:  0,
			AckNum:  p.SeqNum + 1,
			Payload: []byte{},
		}
		ackPacket.Checksum = ml.Checksum(ackPacket.Payload)

		if _, err := conn.WriteToUDP(ackPacket.ToBytes(), clientAddr); err != nil {
			fmt.Println("❌ Erro ao enviar ACK:", err)
			return
		}
		handleReport(p, clientAddr,mm)
	default:
		fmt.Printf("⚠️ Tipo de pacote desconhecido: %d\n", p.MsgType)
	}
}

// handleMissionRequest processa pedidos de missão do rover
func handleMissionRequest(p ml.Packet, clientAddr *net.UDPAddr, conn *net.UDPConn, mm *ml.MissionManager) {
	// Gera um ID único para a missão
	missionID := uint16(time.Now().Unix())

	// Cria payload da missão
	payload := ml.MissionData{
		MsgID:           uint16(missionID),
		Coordinate:      utils.Coordinate{Latitude: 32, Longitude: 25},
		TaskType:        ml.TASK_ENV_ANALYSIS,
		Duration:        10,
		UpdateFrequency: 2,
		Priority:        0,
	}

	// Cria estado da missão
	missionState := &ml.MissionState{
		ID:              missionID,
		IDRover:         0,
		TaskType:        payload.TaskType,
		Duration:        time.Duration(payload.Duration) * time.Second,
		UpdateFrequency: time.Duration(payload.UpdateFrequency) * time.Second,
		LastUpdate:      time.Now(),
		CreatedAt:       time.Now(),
		Priority:        payload.Priority,
        Report:          nil,
		State:           "Pending",
	}

	// Adiciona missão ao gestor
	mm.AddMission(missionState)
	fmt.Printf("📝 Missão %d registada no gestor\n", missionID)

	// Envia a missão ao cliente
	missionPacket := ml.Packet{
		MsgType: ml.MSG_MISSION,
		SeqNum:  0,
		AckNum:  p.SeqNum + 1,
		Payload: payload.ToBytes(),
	}


	// Envia a missão ao cliente
	noMissionPacket := ml.Packet{
		MsgType: ml.MSG_NO_MISSION,
		SeqNum:  0,
		AckNum:  p.SeqNum + 1,
		Payload: []byte{},
	}

	toSend := missionPacket

	if sent {
		toSend = missionPacket
	} else {
		toSend = noMissionPacket
	}


	toSend.Checksum = ml.Checksum(toSend.Payload)

	if _, err := conn.WriteToUDP(toSend.ToBytes(), clientAddr); err != nil {
		fmt.Println("❌ Erro ao enviar missão:", err)
		return
	}
	sent = !sent


	fmt.Printf("✅ Missão %d enviada para %s\n", missionID, clientAddr)
}

// handleACK processa confirmações de entrega
func handleACK(p ml.Packet, clientAddr *net.UDPAddr) {
	fmt.Printf("✅ ACK recebido de %s (SeqNum: %d)\n", clientAddr, p.SeqNum)
}

// handleReport processa relatórios dos rovers
func handleReport(p ml.Packet, clientAddr *net.UDPAddr, mm *ml.MissionManager) {
	fmt.Printf("📊 Relatório recebido de %s\n", clientAddr)

	if len(p.Payload) < 1 {
		fmt.Println("❌ Payload vazio")
		return
	}

	taskType := p.Payload[0]
	fmt.Printf("🔍 TaskType detectado: %d\n", taskType)

	reportTypes := map[uint8]struct {
		name   string
		report ml.Report
	}{
		ml.TASK_SAMPLE_COLLECTION: {"[Amostra]", &ml.SampleReport{}},
		ml.TASK_IMAGE_CAPTURE:     {"[Imagem]", &ml.ImageReport{}},
		ml.TASK_ENV_ANALYSIS:      {"[Ambiente]", &ml.EnvReport{}},
		ml.TASK_REPAIR_RESCUE:     {"[Reparação]", &ml.RepairReport{}},
		ml.TASK_TOPO_MAPPING:      {"[Topografia]", &ml.TopoReport{}},
		ml.TASK_INSTALLATION:      {"[Instalação]", &ml.InstallReport{}},
	}

	reportInfo, exists := reportTypes[taskType]
	if !exists {
		fmt.Printf("⚠️ TaskType desconhecido: %d\n", taskType)
		return
	}

	if err := reportInfo.report.FromBytes(p.Payload); err != nil {
		fmt.Printf("❌ Erro ao desserializar %s: %v\n", reportInfo.name, err)
		return
	}

	if reportInfo.report.IsLast() {
		fmt.Printf("🏁 Último relatório recebido.\n")
	}
    
	fmt.Printf("✅ %s %s\n", reportInfo.name, reportInfo.report.String())



	// Atualiza o estado da missão no Mission Manager
	ml.UpdateMission(mm, reportInfo.report)

    mm.PrintMissions()
}

