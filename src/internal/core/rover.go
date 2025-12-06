package core

import (
	"fmt"
	"net"
	"src/internal/devices"
	"src/internal/ml"
	"src/internal/ts"
	"src/utils"
	pl "src/utils/packetsLogic"
	"sync"
	"time"
	"src/utils/logger"
)

// RoverBase is a basic structure used in both ML and TS contexts
type RoverBase struct {
	ID         uint8            // Unique Rover ID
	CurrentPos utils.Coordinate // Current Coordinates of the Rover
}

// MissionQueue manages missions with 3 priority levels
type MissionQueue struct {
	Priority1 []ml.MissionData // High priority missions
	Priority2 []ml.MissionData // Medium priority missions
	Priority3 []ml.MissionData // Low priority missions
	Mu        sync.Mutex       // Mutex for queue operations
	BatchSize uint8            // Number of missions to request at once
}

// RoverMLState holds the state related to MissionLink connection
type RoverMLState struct {
	// Mission management
	ActiveMissions      uint8         // Number of active missions
	Cond                *sync.Cond    // Condition for mission synchronization
	CondMu              sync.Mutex    // Mutex for the condition
	Waiting             bool          // Indicates if the rover is waiting for a mission
	MissionReceivedChan chan bool     // Channel to signal mission reception
	SeqNum              uint16        // Sequence number for sending packets
	Suspended           bool          // Indicates if rover is suspended due to low battery
	SuspendMu           sync.Mutex    // Mutex for suspension state
	MissionQueue        *MissionQueue // Queue for managing missions by priority

	// Packet and sequence number management
	ExpectedSeq uint16
	Buffer      map[uint16]ml.Packet
	BufferMu    sync.Mutex

	// Sliding window for ACK and retransmission control
	Window *pl.Window // Sliding window specific to this rover
}

// RoverMLConnection holds the UDP connection details for MissionLink
type RoverMLConnection struct {
	Conn *net.UDPConn // UDP connection with the mothership
	Addr *net.UDPAddr // Mothership address
}

// RoverTSState holds the state related to TelemetryLink connection
type RoverInfo struct {
	State   string  // e.g., "Idle", "Moving", "Sampling"
	Battery uint8   // Battery level percentage
	Speed   float32 // Speed in m/s
}

// Device interfaces and mock implementations would go here
type Devices struct {
	GPS              devices.GPS
	Thermometer      devices.Thermometer
	Battery          devices.Battery
	Camera           devices.Camera
	ChemicalAnalyzer devices.ChemicalAnalyzer
}

// RoverSystem encapsulates all subsystems of the rover
type RoverSystem struct {
	*RoverBase                    // Basic rover info
	ML         *RoverMLState      // MissionLink state
	TS         *ts.RoverTSState   // TelemetryLink state
	MLConn     *RoverMLConnection // MissionLink connection
	Devices    *Devices           // Attached devices
	Logger	   *logger.Logger    // Logger instance
}

// requestID contacts the mothership to request a unique rover ID and update frequency
func requestID(mothershipAddr string) (uint8, uint, error) {
	// Make TCP connection to mothership
	conn, err := net.Dial("tcp", mothershipAddr)
	if err != nil {
		return 0, 0, fmt.Errorf("error connecting to ID server: %v", err)
	}
	defer conn.Close()

	// Read 2 bytes: 1 for ID and 1 for update frequency
	buf := make([]byte, 2)
	conn.SetReadDeadline(time.Now().Add(3 * time.Second))
	_, err = conn.Read(buf)
	if err != nil {
		return 0, 0, fmt.Errorf("timeout or error receiving ID: %v", err)
	}

	// Parse ID and update frequency
	id := buf[0]
	updateFrequency := uint(buf[1])
	fmt.Printf("✅ ID received from mothership: %d (updateFrequency=%d)\n", id, updateFrequency)

	return id, updateFrequency, nil
}

// initConnection initializes the UDP connection to the mothership for MissionLink
func initConnection(targetAddr string) (*RoverMLConnection, error) {
	// Resolve mothership UDP address
	motherAddr, err := net.ResolveUDPAddr("udp", targetAddr)
	if err != nil {
		return nil, err
	}

	// Create local UDP connection
	roverConn, err := net.ListenUDP("udp", &net.UDPAddr{Port: 0})
	if err != nil {
		return nil, err
	}

	fmt.Printf("✅ Local UDP connection on port %d -> Target %s\n", roverConn.LocalAddr().(*net.UDPAddr).Port, motherAddr)

	return &RoverMLConnection{
		Conn: roverConn,
		Addr: motherAddr,
	}, nil
}

// NewRoverSystem creates and initializes a RoverSystem
func NewRoverSystem(motherUDP string, motherTCPID string) *RoverSystem {
	// Request ID via TCP
	roverID, updateFrequency, err := requestID(motherTCPID)
	if err != nil {
		fmt.Println("❌ Error obtaining ID:", err)
		return nil
	}

	log, err := logger.NewLogger(
		fmt.Sprintf("rover_%d.log", roverID),
		logger.DestConsole|logger.DestFile,
		logger.DEBUG,
		nil,
	)
	if err != nil {
		fmt.Println("❌ Error initializing logger:", err)
		return nil
	}

	// Connect UDP
	roverConn, err := initConnection(motherUDP)
	if err != nil {
		return nil
	}

	log.Infof("Rover", "Rover %d initialized with update frequency %d", roverID, updateFrequency)


	// Return initialized RoverSystem
	return &RoverSystem{
		RoverBase: &RoverBase{
			ID: roverID,
			// Start near the center with a tiny id-based offset
			CurrentPos: utils.Coordinate{
				Latitude:  0.000 + float64(roverID)*0.001,
				Longitude: 0.000 + float64(roverID)*0.001,
			},
		},
		ML: &RoverMLState{
			ActiveMissions:      0,
			Cond:                sync.NewCond(&sync.Mutex{}),
			CondMu:              sync.Mutex{},
			ExpectedSeq:         0,
			Waiting:             false,
			MissionReceivedChan: make(chan bool, 1),
			Buffer:              make(map[uint16]ml.Packet),
			BufferMu:            sync.Mutex{},
			Window:              pl.NewWindow(),
			Suspended:           false,
			SuspendMu:           sync.Mutex{},
			MissionQueue: &MissionQueue{
				Priority1: make([]ml.MissionData, 0),
				Priority2: make([]ml.MissionData, 0),
				Priority3: make([]ml.MissionData, 0),
				BatchSize: 3,
			},
		},
		TS: &ts.RoverTSState{
			State:           "Idle",
			Battery:         100,
			Speed:           0.0,
			UpdateFrequency: updateFrequency,
		},

		MLConn: roverConn,

		Devices: &Devices{
			GPS: devices.NewMockGPS(utils.Coordinate{
				Latitude:  0.000 + float64(roverID)*0.001,
				Longitude: 0.000 + float64(roverID)*0.001,
			}),
			Thermometer:      devices.NewMockThermometer(),
			Battery:          devices.NewMockBattery(100),
			Camera:           devices.NewMockCamera(),
			ChemicalAnalyzer: devices.NewMockChemicalAnalyzer(),
		},
		Logger: log,
	}
}
