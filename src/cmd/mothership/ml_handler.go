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
	"time"
)

// handlePacket processes each packet on a separate goroutine
func (ms *MotherShip) handlePacket(state *core.RoverState, pkt ml.Packet) {

	// Processor handles business logic ONLY
	// ACK processing (implicit and explicit) is handled automatically by HandleOrderedPacket
	processor := func(p ml.Packet) {
		ms.dispatchPacket(p, state)
	}

	// Determine packet handling options
	isPureAck := pkt.MsgType == ml.MSG_ACK
	// Don't auto-ACK for REQUEST (response is MSG_MISSION which acts as implicit ACK)
	// Don't auto-ACK for pure ACK packets
	shouldAutoAck := pkt.MsgType != ml.MSG_REQUEST && !isPureAck

	// Use the generic ordered packet handler
	go pl.HandleOrderedPacket(
		pkt,
		&state.ExpectedSeq,
		state.Buffer,
		&state.WindowLock,
		ms.Conn,
		state.Addr,
		state.Window,
		0,
		processor,
		isPureAck,     // skipOrdering: only for pure ACKs
		shouldAutoAck, // autoAck: send ACK for REPORT, not for REQUEST or ACK
		ms.Logger.CreateLogCallback("ML"),
	)
}

// receiver continuously reads UDP packets
func (ms *MotherShip) receiver(port string) {
	// Convert string to int
	portNum, err := strconv.Atoi(port)

	if err != nil {
		fmt.Println("❌ Error converting port:", err)
		return
	}

	// Create UDP address
	mothershipConn, err := net.ListenUDP("udp", &net.UDPAddr{
		IP:   nil, // Listen on all IPV4 or IPV6 interfaces
		Port: portNum,
	})
	if err != nil {
		fmt.Println("❌ Error starting UDP receiver:", err)
		return
	}
	defer mothershipConn.Close()

	// Assign connection to MotherShip struct
	ms.Conn = mothershipConn

	// Buffer for incoming packets
	buf := make([]byte, 65535)

	// Main loop to read packets
	for {
		n, addr, err := ms.Conn.ReadFromUDP(buf)
		if err != nil {
			fmt.Println("Error reading packet:", err)
			continue
		}

		var packet ml.Packet
		packet.Decode(buf[:n])
		roverID := packet.RoverId

		ms.Mu.Lock()
		state, exists := ms.Rovers[roverID]

		// If rover state does not exist, create it
		if !exists {
			ms.NewRoverState(roverID, addr, &packet, &state)
		}
		ms.Mu.Unlock()

		// Create goroutine to process the packet
		go ms.handlePacket(state, packet)
	}
}

// NewRoverState sets up a new RoverState for a newly connected rover
func (ms *MotherShip) NewRoverState(roverID uint8, addr *net.UDPAddr, packet *ml.Packet, state **core.RoverState) {
	// Create and initialize RoverState
	*state = &core.RoverState{
		Addr:             addr,
		SeqNum:           0,
		ExpectedSeq:      packet.SeqNum,
		Buffer:           make(map[uint16]ml.Packet),
		Window:           pl.NewWindow(),
		NumberOfMissions: 0,
	}

	// Register the new rover state
	ms.Rovers[roverID] = *state
	ms.Logger.Infof("ML", "🆕 New rover registered: %d", roverID)

	// Register rover in RoverInfo manager
	ms.RoverInfo.AddRover(&ts.RoverTSState{
		ID:       roverID,
		State:    "Connected",
		Battery:  100,
		Speed:    0,
		Position: utils.Coordinate{Latitude: 0, Longitude: 0},
		QueuedMissions: ts.QueueInfo{
			Priority1Count: 0,
			Priority2Count: 0,
			Priority3Count: 0,
			Priority1IDs:   []uint16{},
			Priority2IDs:   []uint16{},
			Priority3IDs:   []uint16{},
		},
	})

	// Publish new rover event
	if ms.APIServer != nil {
		rover := ms.RoverInfo.GetRover(roverID)
		if rover != nil {
			ms.APIServer.PublishUpdate("rover_connected", rover)
		}
	}
}

// dispatchPacket forwards the packet to the correct handler based on its type
// Note: ACK processing (implicit and explicit) is handled automatically by HandleOrderedPacket
func (ms *MotherShip) dispatchPacket(pkt ml.Packet, state *core.RoverState) {
	switch pkt.MsgType {
	case ml.MSG_REQUEST:
		ms.handleMissionRequest(pkt, state, pkt.RoverId)
	case ml.MSG_ACK:
		// Pure ACK - already processed by HandleOrderedPacket, nothing else to do
	case ml.MSG_REPORT:
		ms.handleReport(pkt, state)
	default:
		ms.Logger.Warnf("ML", "⚠️ Unknown packet type: %d", pkt.MsgType)
	}
}

// handleMissionRequest processes mission requests from the rover
func (ms *MotherShip) handleMissionRequest(pkt ml.Packet, state *core.RoverState, roverID uint8) {
	// Extract number of missions requested from payload (default to 1 if empty)
	numMissionsRequested := uint8(1)
	if len(pkt.Payload) > 0 {
		numMissionsRequested = pkt.Payload[0]
	}

	ms.Logger.Infof("ML", "📬 Rover %d requesting %d missions", roverID, numMissionsRequested)

	missionsSent := uint8(0)

	// Calculate AckNum for the REQUEST packet using protocol helper
	ackNumForRequest := pl.CalculateAckNum(pkt)

	for i := uint8(0); i < numMissionsRequested; i++ {
		// Use ackNum only for the first mission/response as implicit ACK for the REQUEST
		ackNum := uint16(0)
		if i == 0 {
			ackNum = ackNumForRequest
		}
		select {
		case missionState := <-ms.MissionQueue:

			ms.assignMissionToRover(missionState, roverID, state, ackNum)
			missionsSent++

		default:
			// Empty queue - no more missions available
			ms.Logger.Warnf("ML", "⚠️ Mission queue empty after sending %d/%d missions", missionsSent, numMissionsRequested)
			if missionsSent == 0 {
				ms.sendNoMission(state, ackNum)
			}
			return
		}
	}

	ms.Logger.Infof("ML", "✅ Sent %d missions to rover %d", missionsSent, roverID)
}

// assignMissionToRover assigns a mission to the selected rover and sends it
func (ms *MotherShip) assignMissionToRover(missionState ml.MissionState, roverID uint8, targetState *core.RoverState, ackNum uint16) {
	// Mission obtained
	missionState.IDRover = roverID // Assign the rover to the mission
	missionState.CreatedAt = time.Now()
	missionState.LastUpdate = time.Now()
	missionState.State = "Pending"
	ms.MissionManager.AddMission(&missionState)

	// Increment rover's mission count
	targetState.NumberOfMissions++

	// Publish mission created event
	ms.publishMissionEvents(&missionState, "mission_created")

	// Send mission to the selected rover
	missionData := ml.MissionData{
		MsgID:           missionState.ID,
		Coordinate:      missionState.Coordinate,
		TaskType:        missionState.TaskType,
		Duration:        uint32(missionState.Duration),
		UpdateFrequency: uint32(missionState.UpdateFrequency),
		Priority:        missionState.Priority,
	}

	payload := missionData.Encode()

	pl.CreateAndSendPacket(
		ms.Conn,
		targetState.Addr,
		0,
		ml.MSG_MISSION,
		&targetState.SeqNum,
		ackNum, // Implicit ACK for the REQUEST
		payload,
		targetState.Window,
		&targetState.WindowLock,
		ms.Logger.CreateLogCallback("ML"),
	)

	ms.Logger.Infof("ML", "✅ Mission %d sent to %s", missionState.ID, targetState.Addr)

	// Change state to "Moving to" after sending the mission
	ms.MissionManager.UpdateMissionState(missionState.ID, "Pending")
	ms.publishMissionEvents(&missionState, "mission_update")
}

// publishMissionEvents publishes mission events to the API server
func (ms *MotherShip) publishMissionEvents(mission *ml.MissionState, eventType string) {
	if ms.APIServer != nil {
		ms.APIServer.PublishUpdate(eventType, mission)
	}
}

// sendNoMission sends a NO_MISSION packet to a rover
func (ms *MotherShip) sendNoMission(state *core.RoverState, ackNum uint16) {
	ms.Logger.Warnf("ML", "⚠️ Mission queue empty or rovers overloaded. Sending NO_MISSION to %s", state.Addr)

	pl.CreateAndSendPacket(
		ms.Conn,
		state.Addr,
		0,
		ml.MSG_NO_MISSION,
		&state.SeqNum,
		ackNum, // Implicit ACK for the REQUEST
		[]byte{},
		state.Window,
		&state.WindowLock,
		ms.Logger.CreateLogCallback("ML"),
	)

}

// handleReport processes reports from rovers
func (ms *MotherShip) handleReport(p ml.Packet, state *core.RoverState) {
	ms.Logger.Debugf("ML", "📊 Report received from %s", state.Addr)
	if len(p.Payload) < ml.REPORT_HEADER_SIZE {
		ms.Logger.Errorf("ML", "❌ Empty or incomplete payload")
		return
	}

	var report ml.Report
	if err := report.Decode(p.Payload); err != nil {
		ms.Logger.Errorf("ML", "❌ Error deserializing report: %v", err)
		return
	}

	ms.Logger.Infof("ML", "✅ Report received: TaskType=%d, MissionID=%d, IsLast=%v, PayloadLen=%d",
		report.Header.TaskType, report.Header.MissionID, report.Header.IsLastReport, len(report.Payload))

	if report.Header.IsLastReport {
		ms.Logger.Infof("ML", "🏁 Last report received for mission %d", report.Header.MissionID)
		ms.Mu.Lock()
		if state.NumberOfMissions > 0 {
			state.NumberOfMissions--
		}
		ms.Mu.Unlock()
		ms.MissionManager.PrintMissions()
	}

	// Update mission state in Mission Manager
	ml.UpdateMission(ms.MissionManager, report)

	// Publish mission update event
	if ms.APIServer != nil {
		mission := ms.MissionManager.GetMission(report.Header.MissionID)
		if mission != nil {
			ms.APIServer.PublishUpdate("mission_update", mission)
		}
	}

}
