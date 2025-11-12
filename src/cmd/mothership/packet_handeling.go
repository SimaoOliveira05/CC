package main

import (
	"fmt"
	"src/internal/ml"
	"src/utils"
	"time"
)

// handlePacket processa cada pacote numa goroutine separada
func (ms *MotherShip) handlePacket(state *RoverState, pkt ml.Packet) {
	state.WindowLock.Lock()
	defer state.WindowLock.Unlock()

	if pkt.MsgType == ml.MSG_ACK {
		// Pacote ACK — processa diretamente
		ms.handleAck(pkt, state)
		return
	}

	seq := pkt.SeqNum
	expected := state.ExpectedSeq

	switch {
	case seq == expected:
		// Pacote esperado
		go ms.dispatchPacket(pkt, state)
		state.ExpectedSeq++
		ms.sendAck(state, seq)

		for {
			if packet, ok := state.Buffer[state.ExpectedSeq]; ok {
				delete(state.Buffer, state.ExpectedSeq)
				bufferedPkt := packet
				go ms.dispatchPacket(bufferedPkt, state)
				ms.sendAck(state, state.ExpectedSeq)
				state.ExpectedSeq++
			} else {
				break // Não há mais pacotes consecutivos
			}
		}

	case seq < expected:
		ms.sendAck(state, seq)

	case seq > expected:
		state.Buffer[seq] = pkt
		ms.sendAck(state, expected)
	}
}

// dispatchPacket encaminha o pacote para o handler correto conforme o tipo
func (ms *MotherShip) dispatchPacket(pkt ml.Packet, state *RoverState) {
	switch pkt.MsgType {

	case ml.MSG_REQUEST:
		ms.handleMissionRequest(state)
	case ml.MSG_ACK:
		ms.handleAck(pkt, state)
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

		pkt.Checksum = ml.Checksum(pkt.Payload)
		ms.sendPacket(pkt, state)
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

		noMissionPkt.Checksum = ml.Checksum(noMissionPkt.Payload)
		ms.sendPacket(noMissionPkt, state)
		return
	}
}

// handleACK processa confirmações de entrega
func (ms *MotherShip) handleAck(p ml.Packet, state *RoverState) {
	state.Window.mu.Lock()
	for i := state.Window.lastAckReceived + 1; i < int16(p.AckNum); i++ {
		if ch, exists := state.Window.window[uint32(i)]; exists {
			ch <- 1 // Sinaliza o ACK recebido
			delete(state.Window.window, uint32(i))
		}
	}
	state.Window.lastAckReceived = int16(p.AckNum - 1)
	state.Window.mu.Unlock()
	fmt.Printf("✅ ACK recebido de %s, AckNum: %d (confirmou até SeqNum %d)\n", state.Addr, p.AckNum, p.AckNum-1)
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
	}

	fmt.Printf("✅ %s %s\n", reportInfo.name, reportInfo.report.String())

	// Atualiza o estado da missão no Mission Manager
	ml.UpdateMission(ms.missionManager, reportInfo.report)

	//ms.missionManager.PrintMissions()
}
