package main

import (
	"fmt"
	"net"
	"src/config"
	"src/internal/ml"
	"time"
	"os"
	"sync"
)

type Rover struct {
    id           string
    conn         *net.UDPConn
    addrMother   *net.UDPAddr
    seqNum       uint32
    window       map[uint32]*OutgoingMessage // pacotes enviados mas ainda não ACKed
    sendChan     chan ml.Packet
	activeMission uint8
	mu 			sync.Mutex
	cond 	  *sync.Cond
	waiting 	bool
    //ackChan      chan uint32
    //timeout      time.Duration
}

type OutgoingMessage struct {
    Packet   ml.Packet
    SentAt   time.Time
    Acked    bool
}

func main() {

	// Verifica se o argumento do id foi passado
    if len(os.Args) < 2 {
        fmt.Println("Uso: ./rover1 <id_do_rover>")
        return
    }
    roverID := os.Args[1]

	// Inicializa configuração (isRover = true)
	config.InitConfig(true)
	config.PrintConfig()

	mothershipAddr := config.GetMotherIP()
	udpAddr, err := net.ResolveUDPAddr("udp", mothershipAddr+":9999")

	if err != nil {
		fmt.Println("Erro ao resolver endereço UDP da nave-mãe:", err)
		return
	}

	roverConn, err := net.DialUDP("udp", nil, udpAddr)
	if err != nil {
		fmt.Println("❌ Erro ao conectar:", err)
		return
	}
	defer roverConn.Close()

	rover := Rover{
		id:         roverID,
		conn:       roverConn, // Inicialize com uma conexão UDP real se necessário
		addrMother: udpAddr, // Inicialize com o endereço da mãe resolvido
		seqNum:     0,
		window:     make(map[uint32]*OutgoingMessage),
		sendChan:   make(chan ml.Packet, 100), // buffer de 100, ajuste conforme necessário
		activeMission: 0,
		mu:         sync.Mutex{},
		cond:    sync.NewCond(&sync.Mutex{}),
		waiting:  false,
	}


	go sender(&rover)
	go receiver(&rover)


	for{
		rover.cond.L.Lock()
		for rover.GetActiveMission() != 0 {
			rover.cond.Wait() // Espera até todas as missões acabarem
		}
		rover.cond.L.Unlock()
		
		if(!rover.waiting){
			sendRequest(rover.sendChan)
			rover.waiting = true
		} else {
			time.Sleep(1 * time.Second) // Espera um pouco antes de verificar novamente
		}
	}
}

// Para alterar a flag:
func (r *Rover) IncrementActiveMission() {
    r.mu.Lock()
    defer r.mu.Unlock()
    r.activeMission++
	r.waiting = false
}

// Para ler a flag:
func (r *Rover) GetActiveMission() uint8 {
    r.mu.Lock()
    defer r.mu.Unlock()
    return r.activeMission
}

// Para decrementar a flag:
func (r *Rover) DecrementActiveMission() {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.activeMission > 0 {
		r.activeMission--
		if r.activeMission == 0 {
			r.cond.L.Lock()
			r.cond.Signal()
			r.cond.L.Unlock()
		}
	}
}


func sender(rover *Rover) {
    for pkt := range rover.sendChan {
        // Centraliza o SeqNum
        pkt.SeqNum = uint16(rover.seqNum)
		rover.seqNum++

        // Atualiza checksum após encriptação
        pkt.Checksum = ml.Checksum(pkt.Payload)

        // Envia para a nave-mãe
        _, err := rover.conn.Write(pkt.ToBytes())
        if err != nil {
            fmt.Println("Erro ao enviar pacote:", err)
            continue
        }

        // Regista na window
        rover.window[rover.seqNum] = &OutgoingMessage{
            Packet: pkt,
            SentAt: time.Now(),
            Acked:  false,
        }
        fmt.Printf("Pacote %d enviado e encriptado\n", pkt.SeqNum)
    }
}

func receiver(rover *Rover) {
	buf := make([]byte, 2048)
	for {
		n, _, err := rover.conn.ReadFromUDP(buf)
		if err != nil {
			fmt.Println("Erro ao ler pacote UDP:", err)
			continue
		}
		
		pkt := ml.FromBytes(buf[:n])
		fmt.Printf("📨 Pacote recebido do tipo: %d\n", pkt.MsgType)
	
		switch pkt.MsgType {

			case ml.MSG_MISSION:
				go generate(ml.DataFromBytes(pkt.Payload), rover)
		}
	}
}

func generate(mission ml.MissionData, rover *Rover){

	rover.IncrementActiveMission()
	defer rover.DecrementActiveMission()

	deadline := time.NewTimer(time.Duration(mission.Duration) * time.Second)
	defer deadline.Stop()

	if mission.UpdateFrequency > 0 {
		// Modo periódico: enviar reports a cada UpdateFrequency
		ticker := time.NewTicker(time.Duration(mission.UpdateFrequency) * time.Second)
		defer ticker.Stop()

		for {
			select {

			case <-deadline.C:
				// Termina quando Duration expirar
				sendReport(mission,true, rover.sendChan)
				return
			case <-ticker.C:
				// Enviar report periódico
				sendReport(mission,false, rover.sendChan)
			}
		}
	} else {
		// Modo sem updates: apenas espera Duration e envia um report final
		<-deadline.C
		// Termina quando Duration expirar
		sendReport(mission,true, rover.sendChan)
		return
	}
}


// sendReport serializa e envia um report para a mothership
func sendReport(mission ml.MissionData, final bool, channel chan ml.Packet) {
	payload := buildReportPayload(mission, final)
	if payload == nil {
		return
	}

	pkt := ml.Packet{
		MsgType: ml.MSG_REPORT,
		SeqNum:  0,
		AckNum:  0,
		Checksum: 0,
		Payload: payload,
	}

	channel <- pkt
	fmt.Printf("📤 Report enviado (Missão %d)\n", mission.MsgID)
}

func sendRequest(channel chan ml.Packet){
	req := ml.Packet{MsgType: ml.MSG_REQUEST, SeqNum: 0, AckNum: 0, Checksum: 0, Payload: []byte{}}
	channel <- req
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
