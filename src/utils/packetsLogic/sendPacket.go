package packetslogic
import (
	"fmt"
	"net"
	"src/internal/ml"
	"time"
	"sync"
)

type Window struct {
	LastAckReceived int16
	Window          map[uint32](chan int8) // pacotes enviados mas ainda não ACKed
	Mu              sync.Mutex

	// Campos para cálculo dinâmico do RTO
	SRTT   time.Duration
	RTTVAR time.Duration
	RTO    time.Duration
}

func NewWindow() *Window {
	return &Window{
		LastAckReceived: -1,
		Window:          make(map[uint32](chan int8)),
		Mu:			  sync.Mutex{},
		SRTT:            0,
		RTTVAR:          0,
		RTO:             1200 * time.Millisecond, // fallback inicial
	}
}

func (w *Window) UpdateRTO(sampleRTT time.Duration) {
	if w.SRTT == 0 {
		// Primeiro RTT medido
		w.SRTT = sampleRTT
		w.RTTVAR = sampleRTT / 2
	} else {
		w.RTTVAR = (3*w.RTTVAR + absDuration(w.SRTT-sampleRTT)) / 4
		w.SRTT = (7*w.SRTT + sampleRTT) / 8
	}

	w.RTO = w.SRTT + 4*w.RTTVAR

	// Limites de segurança
	if w.RTO < 200*time.Millisecond {
		w.RTO = 200 * time.Millisecond
	}
	if w.RTO > 5*time.Second {
		w.RTO = 5 * time.Second
	}
}

func absDuration(x time.Duration) time.Duration {
	if x < 0 { return -x }
	return x
}



// SendPacketUDP encodes and sends a packet through UDP connection
func SendPacketUDP(conn *net.UDPConn, addr *net.UDPAddr, packet ml.Packet) error {
	// Encodes the packet
	packet.Checksum = ml.Checksum(packet.Payload)
	encodedPacket := packet.ToBytes()


	// Sends the encoded data to the specified address port
	_, error := conn.WriteToUDP(encodedPacket, addr)
	pktType := "UNKNOWN"
	switch packet.MsgType {
		case ml.MSG_MISSION:
			pktType = "MSG_MISSION"
		case ml.MSG_NO_MISSION:
			pktType = "MSG_NO_MISSION"
		case ml.MSG_ACK:
			pktType = "MSG_ACK"
		case ml.MSG_REPORT:
			pktType = "MSG_REPORT"
		case ml.MSG_REQUEST:
			pktType = "MSG_REQUEST"
	}
	fmt.Printf("[PACKET-LOGIC] Sent packet of type %s with seq: %d to %s\n", pktType, packet.SeqNum, addr.String())
	return error
}

// packetManager gerencia o envio e retransmissão de um pacote até receber o ACK
func PacketManager(conn *net.UDPConn, addr *net.UDPAddr, pkt ml.Packet, window *Window) {

	if(pkt.MsgType == ml.MSG_ACK){
		fmt.Printf("📤 ACK enviado, AckNum: %d\n", pkt.AckNum)
		if err := SendPacketUDP(conn, addr, pkt); err != nil {
			fmt.Println("❌ Erro ao enviar pacote:", err)
			return
		}
		return
	}

    retries := 0 
	maxRetries := 5
	window.Mu.Lock()

	ch := make(chan int8, 1)
    window.Window[uint32(pkt.SeqNum)] = ch
    window.Mu.Unlock()
	for {

		sendTime := time.Now()

        if err := SendPacketUDP(conn, addr, pkt); err != nil {
			fmt.Println("❌ Erro ao enviar pacote:", err)
			return
		}

		// Ler timeout com lock (mínimo possível)
		window.Mu.Lock()
		rto := window.RTO
		window.Mu.Unlock()

        select {
        case <-ch:
			// ACK recebido → medir RTT
			rtt := time.Since(sendTime)

			window.Mu.Lock()
			window.UpdateRTO(rtt)
			window.Mu.Unlock()

            fmt.Printf("✅ ACK para %d | RTT: %v | Novo RTO: %v\n",
				pkt.SeqNum, rtt, window.RTO)

            return
		case <-time.After(rto):
			retries++
			if retries > maxRetries {
				fmt.Printf("❌ Falha ao receber ACK para SeqNum %d após %d tentativas. Abortando...\n", pkt.SeqNum, maxRetries)
				return
			}
			fmt.Printf("⏱️ Timeout esperando ACK para SeqNum %d. Retransmitindo (tentativa %d)...\n", pkt.SeqNum, retries)
        }
    }
}


func SendAck(conn *net.UDPConn, addr *net.UDPAddr, ackNum uint16, window *Window, roverId uint8) {
	ackPacket := ml.Packet{
		RoverId: roverId,
		MsgType: ml.MSG_ACK,
		SeqNum:  0,
		AckNum:  ackNum + 1,
		Payload: []byte{},
	}

	PacketManager(conn, addr, ackPacket, window)
}



