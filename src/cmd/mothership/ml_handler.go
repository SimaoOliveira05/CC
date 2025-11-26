package main

import (
	"fmt"
	"net"
	"src/internal/core"
	"src/internal/ml"
	"src/internal/ts"
	"src/utils"
	pl "src/utils/packetsLogic"
	"strconv"
	"sync"
	"time"
)

// handlePacket processa cada pacote numa goroutine separada
func (ms *MotherShip) handlePacket(state *core.RoverState, pkt ml.Packet) {

	// Closure que captura o 'state'
	processor := func(p ml.Packet) {
		ms.dispatchPacket(p, state)
	}

	pl.HandleOrderedPacket(
		pkt,
		&state.ExpectedSeq,
		state.Buffer,
		&state.WindowLock,
		ms.Conn,
		state.Addr,
		state.Window,
		0,
		processor,
		pkt.MsgType == ml.MSG_ACK,
	)
}

// receiver lê continuamente pacotes UDP
func (ms *MotherShip) receiver(port string) {
	// Converte string para int
	portNum, err := strconv.Atoi(port)

	if err != nil {
		fmt.Println("❌ Erro ao converter porta:", err)
		return
	}

	// Cria o endereço UDP
	mothershipConn, err := net.ListenUDP("udp", &net.UDPAddr{
		IP:   net.ParseIP("0.0.0.0"),
		Port: portNum,
	})
	if err != nil {
		fmt.Println("❌ Erro ao iniciar receptor UDP:", err)
		return
	}
	defer mothershipConn.Close()

	ms.Conn = mothershipConn
	buf := make([]byte, 1024)

	for {
		n, addr, err := ms.Conn.ReadFromUDP(buf)
		if err != nil {
			fmt.Println("Erro a ler pacote:", err)
			continue
		}

		packet := ml.FromBytes(buf[:n])
		roverID := packet.RoverId

		ms.Mu.Lock()
		state, exists := ms.Rovers[roverID]
		if !exists {
			state = &core.RoverState{
				Addr:        addr,
				SeqNum:      0,
				ExpectedSeq: packet.SeqNum,
				Buffer:      make(map[uint16]ml.Packet),
				Window: &pl.Window{
					LastAckReceived: -1,
					Window:          make(map[uint32](chan int8)),
					Mu:              sync.Mutex{},
				},
				NumberOfMissions: 0,
			}
			ms.Rovers[roverID] = state
			fmt.Printf("🆕 Novo rover registado: %d\n", roverID)

			// 🔥 Register rover in RoverInfo manager
			ms.RoverInfo.AddRover(&ts.RoverTSState{
				ID:       roverID,
				State:    "Conectado",
				Battery:  100,
				Speed:    0,
				Position: utils.Coordinate{Latitude: 0, Longitude: 0},
			})

			// 🔥 Publish new rover event
			if ms.APIServer != nil {
				rover := ms.RoverInfo.GetRover(roverID)
				if rover != nil {
					ms.APIServer.PublishUpdate("rover_connected", rover)
				}
			}
		}
		ms.Mu.Unlock()

		// Criar goroutine para processar o pacote
		go ms.handlePacket(state, packet)
	}
}

// dispatchPacket encaminha o pacote para o handler correto conforme o tipo
func (ms *MotherShip) dispatchPacket(pkt ml.Packet, state *core.RoverState) {
	switch pkt.MsgType {

	case ml.MSG_REQUEST:
		ms.handleMissionRequest(state)
	case ml.MSG_ACK:
		pl.HandleAck(pkt, state.Window)
	case ml.MSG_REPORT:
		ms.handleReport(pkt, state)
	default:
		fmt.Printf("⚠️ Tipo de pacote desconhecido: %d\n", pkt.MsgType)
	}
}

// handleMissionRequest processa pedidos de missão do rover
func (ms *MotherShip) handleMissionRequest(state *core.RoverState) {
	// Gera um ID único para a missão
	select {
	case missionState := <-ms.MissionQueue:

		ms.Mu.Lock()
		targetRoverID, targetState := ms.findLeastLoadedRover()
		ms.Mu.Unlock()

		if targetState == nil {
			// Todos os rovers estão com 3+ missões, recoloca na fila
			fmt.Printf("⚠️ Todos os rovers estão sobrecarregados. Missão %d devolvida à fila.\n", missionState.ID)
			ms.MissionQueue <- missionState

			// Envia NO_MISSION ao rover que pediu
			ms.sendNoMission(state)
			return
		}
		// Missão obtida
		missionState.IDRover = targetRoverID // 🔥 Atribuir o rover à missão
		missionState.CreatedAt = time.Now()
		missionState.LastUpdate = time.Now()
		missionState.State = "Pending"
		ms.MissionManager.AddMission(&missionState)

		// 4. Incrementar contador de missões do rover
		targetState.NumberOfMissions++

		// 🔥 Publish mission created event
		if ms.APIServer != nil {
			ms.APIServer.PublishUpdate("mission_created", &missionState)
		}
		// Enviar missão para o rover
		missionData := ml.MissionData{
			MsgID:           missionState.ID,
			Coordinate:      missionState.Coordinate,
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

		pl.PacketManager(ms.Conn, state.Addr, pkt, state.Window)
		fmt.Printf("✅ Missão %d enviada para %s\n", missionState.ID, state.Addr)

		// Muda estado para "Moving to" após enviar a missão
		ms.MissionManager.UpdateMissionState(missionState.ID, "Moving to")
		if ms.APIServer != nil {
			ms.APIServer.PublishUpdate("mission_update", &missionState)
		}
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

		pl.PacketManager(ms.Conn, state.Addr, noMissionPkt, state.Window)
		return
	}
}

// findLeastLoadedRover encontra o rover com menos missões ativas (máx 3)
func (ms *MotherShip) findLeastLoadedRover() (uint8, *core.RoverState) {
	var bestRoverID uint8
	var bestState *core.RoverState
	minMissions := uint8(255) // Valor alto inicial

	for id, state := range ms.Rovers {
		if state.NumberOfMissions < 3 && state.NumberOfMissions < minMissions {
			minMissions = state.NumberOfMissions
			bestRoverID = id
			bestState = state
		}
	}

	if bestState == nil {
		return 0, nil // Nenhum rover disponível
	}

	return bestRoverID, bestState
}

// sendNoMission envia pacote NO_MISSION para um rover
func (ms *MotherShip) sendNoMission(state *core.RoverState) {
	fmt.Printf("⚠️ Fila de missões vazia ou rovers sobrecarregados. Enviando NO_MISSION para %s\n", state.Addr)

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

	pl.PacketManager(ms.Conn, state.Addr, noMissionPkt, state.Window)
}

// handleReport processa relatórios dos rovers
func (ms *MotherShip) handleReport(p ml.Packet, state *core.RoverState) {
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
		ms.Mu.Lock()
		if state.NumberOfMissions > 0 {
			state.NumberOfMissions--
		}
		ms.Mu.Unlock()

		ms.MissionManager.PrintMissions()
	}

	fmt.Printf("✅ %s %s\n", reportInfo.name, reportInfo.report.String())

	// Atualiza o estado da missão no Mission Manager
	ml.UpdateMission(ms.MissionManager, reportInfo.report)

	// 🔥 Publish mission update event
	if ms.APIServer != nil {
		mission := ms.MissionManager.GetMission(reportInfo.report.GetMissionID())
		if mission != nil {
			ms.APIServer.PublishUpdate("mission_update", mission)
		}
	}

}
