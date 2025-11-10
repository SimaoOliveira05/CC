package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net"
	"os"
	"src/config"
	"src/internal/ml"
	"src/utils"
	"sync"
	"time"
)

type RoverState struct {
	Addr        *net.UDPAddr
	SeqNum      uint16
	ExpectedSeq uint16
	Buffer      map[uint16]ml.Packet
	WindowLock  sync.Mutex
}

type MotherShip struct {
	conn           *net.UDPConn
	rovers         map[string]*RoverState // key: IP (ou ID do rover)
	missionManager *ml.MissionManager
	missionQueue   chan ml.MissionState
	mu             sync.Mutex
}

func main() {
	config.InitConfig(false)
	config.PrintConfig()

	addr, _ := net.ResolveUDPAddr("udp", config.GetMotherIP()+":9999")
	conn, _ := net.ListenUDP("udp", addr)
	defer conn.Close()

	fmt.Println("🛰️ Nave-Mãe à escuta...")

	// Cria o estado da Nave-Mãe
	mothership := MotherShip{
		conn:           conn,
		rovers:         make(map[string]*RoverState),
		missionManager: ml.NewMissionManager(),
		missionQueue:   make(chan ml.MissionState, 100),
		mu:             sync.Mutex{},
	}

	// Carrega missões do JSON para a missionQueue
	if err := loadMissionsFromJSON("missions.json", mothership.missionQueue); err != nil {
		fmt.Println("❌ Erro ao carregar missões do JSON:", err)
	} else {
		fmt.Println("✅ Missões carregadas com sucesso na missionQueue")
	}

	// Goroutine para ler pacotes UDP
	go mothership.receiver()

	// Loop infinito
	select {}
}

// loadMissionsFromJSON lê missões de um ficheiro JSON e coloca-as na missionQueue
func loadMissionsFromJSON(filename string, queue chan ml.MissionState) error {
	file, err := os.Open(filename)
	if err != nil {
		return fmt.Errorf("erro ao abrir ficheiro: %v", err)
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		return fmt.Errorf("erro ao ler ficheiro: %v", err)
	}

	var missions []ml.MissionState
	if err := json.Unmarshal(data, &missions); err != nil {
		return fmt.Errorf("erro ao fazer unmarshal do JSON: %v", err)
	}

	for _, mission := range missions {
		queue <- mission
	}

	fmt.Printf("📋 %d missões enfileiradas\n", len(missions))
	return nil
}

// receiver lê continuamente pacotes UDP
func (ms *MotherShip) receiver() {
	buf := make([]byte, 1024)

	for {
		n, addr, err := ms.conn.ReadFromUDP(buf)
		if err != nil {
			fmt.Println("Erro a ler pacote:", err)
			continue
		}

		packet := ml.FromBytes(buf[:n])
		roverID := addr.String()

		ms.mu.Lock()
		state, exists := ms.rovers[roverID]
		if !exists {
			state = &RoverState{
				Addr:        addr,
				SeqNum:      0,
				ExpectedSeq: packet.SeqNum,
				Buffer:      make(map[uint16]ml.Packet),
			}
			ms.rovers[roverID] = state
		}
		ms.mu.Unlock()

		// Criar goroutine para processar o pacote
		go ms.handlePacket(state, packet)
	}
}

// handlePacket processa cada pacote numa goroutine separada
func (ms *MotherShip) handlePacket(state *RoverState, pkt ml.Packet) {
	state.WindowLock.Lock()
	defer state.WindowLock.Unlock()

	seq := pkt.SeqNum
	expected := state.ExpectedSeq

	switch {
	case seq == expected:
		// Pacote esperado
		go ms.processPacket(pkt, state)
		state.ExpectedSeq++
		ms.sendAck(state, seq)

		// 👇 Após processar o esperado, vê se há mais pacotes no buffer
		for {
			if packet, ok := state.Buffer[state.ExpectedSeq]; ok {
				delete(state.Buffer, state.ExpectedSeq)
				bufferedPkt := packet
				go ms.processPacket(bufferedPkt, state)
				ms.sendAck(state, state.ExpectedSeq)
				state.ExpectedSeq++
			} else {
				break // Não há mais pacotes consecutivos
			}
		}

	//case seq < expected:
	// Pacote duplicado — ACK para tranquilizar o rover
	//ms.sendAck(state, seq)

	case seq > expected:
		// Fora de ordem — guarda e ACK cumulativo
		state.Buffer[seq] = pkt
		ms.sendAck(state, expected-1)
	}
}

func (ms *MotherShip) processPacket(pkt ml.Packet, state *RoverState) {
	switch pkt.MsgType {

	case ml.MSG_REQUEST:
		ms.handleMissionRequest(state)
	case ml.MSG_ACK:
		handleACK(pkt, state.Addr)
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

	// Tenta obter 3 missões da fila
	var missionState ml.MissionState
	select {
	case missionState = <-ms.missionQueue:
		// Missão obtida
		missionState.ID = missionID
		missionState.CreatedAt = time.Now()
		ms.missionManager.AddMission(&missionState)
		// Enviar missão para o rover
		missionData := ml.MissionData{
			MsgID:    missionState.ID,
			Coordinate: utils.Coordinate{Latitude: 0, Longitude: 0},
			TaskType: missionState.TaskType,
			Duration: uint32(missionState.Duration),
			UpdateFrequency: uint32(missionState.UpdateFrequency),
			Priority: missionState.Priority,
		}

		payload := missionData.ToBytes()

		state.WindowLock.Lock()

		pkt := ml.Packet{
			RoverId:  0,
			MsgType:  ml.MSG_MISSION,
			SeqNum:   state.SeqNum,
			AckNum:   0,
			Payload:  payload,
		}

		state.SeqNum++
		state.WindowLock.Unlock()

		pkt.Checksum = ml.Checksum(pkt.Payload)
		ms.conn.WriteToUDP(pkt.ToBytes(), state.Addr)
		fmt.Printf("✅ Missão %d enviada para %s\n", missionID, state.Addr)
		return
	default:
		// Fila vazia
		fmt.Printf("⚠️ Fila de missões vazia. Enviando NO_MISSION para %s\n", state.Addr)

		state.WindowLock.Lock()

		noMissionPkt := ml.Packet{
			RoverId: 0,
			MsgType: ml.MSG_NO_MISSION,
			SeqNum:   state.SeqNum,
			AckNum:  0,
			Payload: []byte{},
		}

		state.SeqNum++
		state.WindowLock.Unlock()

		noMissionPkt.Checksum = ml.Checksum(noMissionPkt.Payload)
		ms.conn.WriteToUDP(noMissionPkt.ToBytes(), state.Addr)
		return
	}
}

// handleACK processa confirmações de entrega
func handleACK(p ml.Packet, clientAddr *net.UDPAddr) {
	fmt.Printf("✅ ACK recebido de %s (SeqNum: %d)\n", clientAddr, p.SeqNum)
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

func (ms *MotherShip) sendAck(state *RoverState, ackNum uint16) {
	ackPacket := ml.Packet{
		RoverId: 0,
		MsgType: ml.MSG_ACK,
		SeqNum:  0,
		AckNum:  ackNum,
		Payload: []byte{},
	}
	ackPacket.Checksum = ml.Checksum(ackPacket.Payload)

	if _, err := ms.conn.WriteToUDP(ackPacket.ToBytes(), state.Addr); err != nil {
		fmt.Println("❌ Erro ao enviar ACK:", err)
		return
	}
	fmt.Printf("📤 ACK enviado para %s, AckNum: %d\n", state.Addr, ackNum)
}
