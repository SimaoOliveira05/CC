package main

import (
	"fmt"
	"src/internal/ml"
	"src/utils"
	packetslogic "src/utils/packetsLogic"
	"time"
)

// handlePacket processa cada pacote numa goroutine separada
func (ms *MotherShip) handlePacket(state *RoverState, pkt ml.Packet) {
	// Closure que captura o 'state'
	processor := func(p ml.Packet) {
		ms.dispatchPacket(p, state)
	}

	packetslogic.HandleOrderedPacket(
		pkt,
		&state.ExpectedSeq,
		state.Buffer,
		&state.WindowLock,
		ms.conn,
		state.Addr,
		state.Window,
		0,
		processor,
		pkt.MsgType == ml.MSG_ACK,
	)
}

// dispatchPacket encaminha o pacote para o handler correto conforme o tipo
func (ms *MotherShip) dispatchPacket(pkt ml.Packet, state *RoverState) {
	switch pkt.MsgType {

	case ml.MSG_REQUEST:
		ms.handleMissionRequest(state)
	case ml.MSG_ACK:
		packetslogic.HandleAck(pkt, state.Window)
	case ml.MSG_REPORT:
		ms.handleReport(pkt, state)
	default:
		fmt.Printf("⚠️ Tipo de pacote desconhecido: %d\n", pkt.MsgType)
	}
}

// handleMissionRequest processa pedidos de missão do rover
func (ms *MotherShip) handleMissionRequest(state *RoverState) {
	// Gera um ID único para a missão
	missionID := uint16(time.Now().Unix())

	var missionState ml.MissionState
	select {
	case missionState = <-ms.missionQueue:
		// Missão obtida
		missionState.ID = missionID
		missionState.CreatedAt = time.Now()
		ms.missionManager.AddMission(&missionState)
		// Enviar missão para o rover
		missionData := ml.MissionData{
			MsgID:           missionState.ID,
			Coordinate:      utils.Coordinate{Latitude: 0, Longitude: 0},
			TaskType:        missionState.TaskType,
			Duration:        uint32(missionState.Duration),
			UpdateFrequency: uint32(missionState.UpdateFrequency),
			Priority:        missionState.Priority,
		}

		payload := missionData.ToBytes()

		state.WindowLock.Lock()

		pkt := ml.Packet{
			RoverId: 0,
			MsgType: ml.MSG_MISSION,
			SeqNum:  state.SeqNum,
			AckNum:  0,
			Payload: payload,
		}

		state.SeqNum++
		state.WindowLock.Unlock()

		packetslogic.PacketManager(ms.conn, state.Addr, pkt, state.Window)
		fmt.Printf("✅ Missão %d enviada para %s\n", missionID, state.Addr)
		return
	default:
		// Fila vazia
		fmt.Printf("⚠️ Fila de missões vazia. Enviando NO_MISSION para %s\n", state.Addr)

		state.WindowLock.Lock()

		noMissionPkt := ml.Packet{
			RoverId: 0,
			MsgType: ml.MSG_NO_MISSION,
			SeqNum:  state.SeqNum,
			AckNum:  0,
			Payload: []byte{},
		}

		state.SeqNum++
		state.WindowLock.Unlock()

		packetslogic.PacketManager(ms.conn, state.Addr, noMissionPkt, state.Window)
		return
	}
}

// handleReport processa relatórios dos rovers
func (ms *MotherShip) handleReport(p ml.Packet, state *RoverState) {
	fmt.Printf("📊 Relatório recebido de %s\n", state.Addr)

	if len(p.Payload) < 1 {
		fmt.Println("❌ Payload vazio")
		return
	}

	taskType := p.Payload[0]
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
		ms.missionManager.PrintMissions()
	}

	fmt.Printf("✅ %s %s\n", reportInfo.name, reportInfo.report.String())

	// Atualiza o estado da missão no Mission Manager
	ml.UpdateMission(ms.missionManager, reportInfo.report)
}
