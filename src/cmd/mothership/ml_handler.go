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

// handleTelemetryConnection processes each packet on a separate goroutine
func (ms *MotherShip) handlePacket(state *core.RoverState, pkt ml.Packet) {

	// Closure that captures 'ms' and 'state'
	processor := func(p ml.Packet) {
		ms.dispatchPacket(p, state)
	}

	// Determine if we should auto-acknowledge
	// This occurs for all packets except REQUEST
	shouldAutoAck := pkt.MsgType != ml.MSG_REQUEST

	// Use the generic ordered packet handler
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
		shouldAutoAck,
		func(level, msg string, meta any) {
			ms.EventLogger.Log(level, "ML", msg, meta)
    })
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
			ms.NewRoverState(roverID, addr, &packet, &state);
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
		Addr:        addr,
		SeqNum:      0,
		ExpectedSeq: packet.SeqNum,
		Buffer:      make(map[uint16]ml.Packet),
		Window:      pl.NewWindow(),
		NumberOfMissions: 0,
	}

	// Register the new rover state
	ms.Rovers[roverID] = *state
	fmt.Printf("🆕 New rover registered: %d\n", roverID)

	// Register rover in RoverInfo manager
	ms.RoverInfo.AddRover(&ts.RoverTSState{
		ID:       roverID,
		State:    "Connected",
		Battery:  100,
		Speed:    0,
		Position: utils.Coordinate{Latitude: 0, Longitude: 0},
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
func (ms *MotherShip) dispatchPacket(pkt ml.Packet, state *core.RoverState) {
	switch pkt.MsgType {
	case ml.MSG_REQUEST:
		ms.handleMissionRequest(pkt.SeqNum, state)
	case ml.MSG_ACK:
		pl.HandleAck(pkt, state.Window)
	case ml.MSG_REPORT:
		ms.handleReport(pkt, state)
	default:
		fmt.Printf("⚠️ Unknown packet type: %d\n", pkt.MsgType)
	}
}

// handleMissionRequest processes mission requests from the rover
func (ms *MotherShip) handleMissionRequest(pktSeqNum uint16, state *core.RoverState) {
	select {
		case missionState := <-ms.MissionQueue:
			// Find the least loaded rover
			ms.Mu.Lock()
			targetRoverID, targetState := ms.findLeastLoadedRover()
			ms.Mu.Unlock()

			if targetState == nil {
				// All rovers have 3+ missions, put back in queue
				fmt.Printf("⚠️ All rovers are overloaded. Mission %d returned to queue.\n", missionState.ID)
				ms.MissionQueue <- missionState

				// Send NO_MISSION to the requesting rover
				ms.sendNoMission(state)
				return
			}

			ms.assignMissionToRover(missionState, targetRoverID, targetState, pktSeqNum)
		
		default:
			// Empty queue
			fmt.Printf("⚠️ Mission queue empty. Sending NO_MISSION to %s\n", state.Addr)
			ms.sendNoMission(state)
			return
	}
}

// assignMissionToRover assigns a mission to the selected rover and sends it
func (ms *MotherShip) assignMissionToRover(missionState ml.MissionState, roverID uint8, targetState *core.RoverState, pktSeqNum uint16) {
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

	targetState.WindowLock.Lock()
	pkt := ml.Packet{
		RoverId: 0,
		MsgType: ml.MSG_MISSION,
		SeqNum:  targetState.SeqNum,
		AckNum:  pktSeqNum + 1, // MISSION PACKETS ACTS AS A SYN-ACK
		Payload: payload,
	}

	targetState.SeqNum++
	targetState.WindowLock.Unlock()

	pl.PacketManager(ms.Conn, 
					targetState.Addr, 
					pkt, 
					targetState.Window, 
					func(level, msg string, meta any) {
						ms.EventLogger.Log(level, "ML", msg, meta)
					})

	fmt.Printf("✅ Mission %d sent to %s\n", missionState.ID, targetState.Addr)

	// Change state to "Moving to" after sending the mission
	ms.MissionManager.UpdateMissionState(missionState.ID, "Moving to")
	ms.publishMissionEvents(&missionState, "mission_update")
}

// publishMissionEvents publishes mission events to the API server
func (ms *MotherShip) publishMissionEvents(mission *ml.MissionState, eventType string) {
    if ms.APIServer != nil {
        ms.APIServer.PublishUpdate(eventType, mission)
    }
}

// findLeastLoadedRover finds the rover with the fewest active missions (max 3)
func (ms *MotherShip) findLeastLoadedRover() (uint8, *core.RoverState) {
	var bestRoverID uint8
	var bestState *core.RoverState
	minMissions := uint8(255) // Initial highest value

	// Iterate through rovers to find the least loaded one
	for id, state := range ms.Rovers {
		if state.NumberOfMissions < 3 && state.NumberOfMissions < minMissions {
			minMissions = state.NumberOfMissions
			bestRoverID = id
			bestState = state
		}
	}

	if bestState == nil {
		return 0, nil // No available rover found
	}

	return bestRoverID, bestState
}

// sendNoMission sends a NO_MISSION packet to a rover
func (ms *MotherShip) sendNoMission(state *core.RoverState) {
	fmt.Printf("⚠️ Mission queue empty or rovers overloaded. Sending NO_MISSION to %s\n", state.Addr)

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

	pl.PacketManager(ms.Conn, 
					state.Addr, 
					noMissionPkt, 
					state.Window,
					func(level, msg string, meta any) {
						ms.EventLogger.Log(level, "ML", msg, meta)
					})
}

// handleReport processes reports from rovers
func (ms *MotherShip) handleReport(p ml.Packet, state *core.RoverState) {
	fmt.Printf("📊 Report received from %s\n", state.Addr)
	if len(p.Payload) < ml.REPORT_HEADER_SIZE {
		fmt.Println("❌ Empty or incomplete payload")
		return
	}

	var report ml.Report
	if err := report.Decode(p.Payload); err != nil {
		fmt.Printf("❌ Error deserializing report: %v\n", err)
		return
	}

	fmt.Printf("✅ Report received: TaskType=%d, MissionID=%d, IsLast=%v, PayloadLen=%d\n",
		report.Header.TaskType, report.Header.MissionID, report.Header.IsLastReport, len(report.Payload))

	if report.Header.IsLastReport {
		fmt.Printf("🏁 Last report received.\n")
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
