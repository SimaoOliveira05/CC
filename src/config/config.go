package config

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"time"
)

// ==================== NETWORK PORTS ====================
var (
	API_PORT           string
	TCP_ID_PORT        string
	UDP_COMM_PORT      string
	TCP_TELEMETRY_PORT string
)

// ==================== RETRANSMISSION & RTO ====================
var (
	INITIAL_RTO time.Duration
	MIN_RTO     time.Duration
	MAX_RTO     time.Duration
	MAX_RETRIES int
)

// ==================== TELEMETRY ====================
var (
	DEFAULT_TELEMETRY_FREQ time.Duration
	MAX_MISSED_TELEMETRY   int
)

// ==================== ROVER SETTINGS ====================
var (
	MISSION_BATCH_SIZE       uint8
	TCP_TIMEOUT              time.Duration
	INITIAL_BATTERY          uint8
	NO_MISSION_WAIT          time.Duration
	BATTERY_CHECK_INTERVAL   time.Duration
	BATTERY_MONITOR_INTERVAL time.Duration
)

// ==================== MOTHERSHIP SETTINGS ====================
var (
	MISSION_QUEUE_SIZE     int
	EVENT_LOGGER_SIZE      int
	MAX_MISSIONS_PER_ROVER uint8
)

// ==================== MOVEMENT & PHYSICS ====================
var (
	MAX_SPEED             float64
	MOVEMENT_BATTERY_RATE float64
	TASK_BATTERY_RATE     float64
	ARRIVAL_THRESHOLD     float64
)

// ==================== BATTERY MANAGEMENT ====================
var (
	BATTERY_DRAIN_RATE     float64
	BATTERY_CHARGE_RATE    float64
	CRITICAL_BATTERY_LEVEL uint8
	LOW_BATTERY_LEVEL      uint8
	TARGET_RECHARGE_LEVEL  uint8
)

// ==================== DEVICE SETTINGS ====================
var (
	CAMERA_CHUNK_SIZE      int
	CAMERA_FAIL_CHANCE     float32
	INSTALL_SUCCESS_CHANCE float64
)

// Config holds the global configuration settings
type Config struct {
	MotherIP string
	TestMode bool // Enable metrics collection for testing
}

var GlobalConfig Config

// jsonConfig matches the structure of config.json
type jsonConfig struct {
	// Network
	API_PORT           int `json:"API_PORT"`
	TCP_ID_PORT        int `json:"TCP_ID_PORT"`
	UDP_COMM_PORT      int `json:"UDP_COMM_PORT"`
	TCP_TELEMETRY_PORT int `json:"TCP_TELEMETRY_PORT"`

	// Retransmission
	INITIAL_RTO_MS int `json:"INITIAL_RTO_MS"`
	MIN_RTO_MS     int `json:"MIN_RTO_MS"`
	MAX_RTO_MS     int `json:"MAX_RTO_MS"`
	MAX_RETRIES    int `json:"MAX_RETRIES"`

	// Telemetry
	DEFAULT_TELEMETRY_FREQ_SEC int `json:"DEFAULT_TELEMETRY_FREQ_SEC"`
	MAX_MISSED_TELEMETRY       int `json:"MAX_MISSED_TELEMETRY"`

	// Rover
	MISSION_BATCH_SIZE           int `json:"MISSION_BATCH_SIZE"`
	TCP_TIMEOUT_SEC              int `json:"TCP_TIMEOUT_SEC"`
	INITIAL_BATTERY              int `json:"INITIAL_BATTERY"`
	NO_MISSION_WAIT_SEC          int `json:"NO_MISSION_WAIT_SEC"`
	BATTERY_CHECK_INTERVAL_SEC   int `json:"BATTERY_CHECK_INTERVAL_SEC"`
	BATTERY_MONITOR_INTERVAL_SEC int `json:"BATTERY_MONITOR_INTERVAL_SEC"`

	// Mothership
	MISSION_QUEUE_SIZE     int `json:"MISSION_QUEUE_SIZE"`
	EVENT_LOGGER_SIZE      int `json:"EVENT_LOGGER_SIZE"`
	MAX_MISSIONS_PER_ROVER int `json:"MAX_MISSIONS_PER_ROVER"`

	// Movement
	MAX_SPEED             float64 `json:"MAX_SPEED"`
	MOVEMENT_BATTERY_RATE float64 `json:"MOVEMENT_BATTERY_RATE"`
	TASK_BATTERY_RATE     float64 `json:"TASK_BATTERY_RATE"`
	ARRIVAL_THRESHOLD     float64 `json:"ARRIVAL_THRESHOLD"`

	// Battery
	BATTERY_DRAIN_RATE     float64 `json:"BATTERY_DRAIN_RATE"`
	BATTERY_CHARGE_RATE    float64 `json:"BATTERY_CHARGE_RATE"`
	CRITICAL_BATTERY_LEVEL int     `json:"CRITICAL_BATTERY_LEVEL"`
	LOW_BATTERY_LEVEL      int     `json:"LOW_BATTERY_LEVEL"`
	TARGET_RECHARGE_LEVEL  int     `json:"TARGET_RECHARGE_LEVEL"`

	// Devices
	CAMERA_CHUNK_SIZE      int     `json:"CAMERA_CHUNK_SIZE"`
	CAMERA_FAIL_CHANCE     float32 `json:"CAMERA_FAIL_CHANCE"`
	INSTALL_SUCCESS_CHANCE float64 `json:"INSTALL_SUCCESS_CHANCE"`
}

// InitConfig initializes the global configuration from command-line flags and config.json
func InitConfig(isRover bool, print bool) {
	// Default IP is localhost
	flag.StringVar(&GlobalConfig.MotherIP, "ms-ip", "127.0.0.1", "Mother Ship IP Address")
	flag.BoolVar(&GlobalConfig.TestMode, "test-mode", false, "Enable metrics collection for testing")
	flag.Parse()

	// Read config from config.json
	file, err := os.Open("config.json")
	if err != nil {
		panic(fmt.Sprintf("Failed to open config.json: %v", err))
	}
	defer file.Close()

	var conf jsonConfig
	if err := json.NewDecoder(file).Decode(&conf); err != nil {
		panic(fmt.Sprintf("Failed to decode config.json: %v", err))
	}

	// Assign Network Ports
	API_PORT = fmt.Sprintf("%d", conf.API_PORT)
	TCP_ID_PORT = fmt.Sprintf("%d", conf.TCP_ID_PORT)
	UDP_COMM_PORT = fmt.Sprintf("%d", conf.UDP_COMM_PORT)
	TCP_TELEMETRY_PORT = fmt.Sprintf("%d", conf.TCP_TELEMETRY_PORT)

	// Assign Retransmission Settings
	INITIAL_RTO = time.Duration(conf.INITIAL_RTO_MS) * time.Millisecond
	MIN_RTO = time.Duration(conf.MIN_RTO_MS) * time.Millisecond
	MAX_RTO = time.Duration(conf.MAX_RTO_MS) * time.Millisecond
	MAX_RETRIES = conf.MAX_RETRIES

	// Assign Telemetry Settings
	DEFAULT_TELEMETRY_FREQ = time.Duration(conf.DEFAULT_TELEMETRY_FREQ_SEC) * time.Second
	MAX_MISSED_TELEMETRY = conf.MAX_MISSED_TELEMETRY

	// Assign Rover Settings
	MISSION_BATCH_SIZE = uint8(conf.MISSION_BATCH_SIZE)
	TCP_TIMEOUT = time.Duration(conf.TCP_TIMEOUT_SEC) * time.Second
	INITIAL_BATTERY = uint8(conf.INITIAL_BATTERY)
	NO_MISSION_WAIT = time.Duration(conf.NO_MISSION_WAIT_SEC) * time.Second
	BATTERY_CHECK_INTERVAL = time.Duration(conf.BATTERY_CHECK_INTERVAL_SEC) * time.Second
	BATTERY_MONITOR_INTERVAL = time.Duration(conf.BATTERY_MONITOR_INTERVAL_SEC) * time.Second

	// Assign Mothership Settings
	MISSION_QUEUE_SIZE = conf.MISSION_QUEUE_SIZE
	EVENT_LOGGER_SIZE = conf.EVENT_LOGGER_SIZE
	MAX_MISSIONS_PER_ROVER = uint8(conf.MAX_MISSIONS_PER_ROVER)

	// Assign Movement Settings
	MAX_SPEED = conf.MAX_SPEED
	MOVEMENT_BATTERY_RATE = conf.MOVEMENT_BATTERY_RATE
	TASK_BATTERY_RATE = conf.TASK_BATTERY_RATE
	ARRIVAL_THRESHOLD = conf.ARRIVAL_THRESHOLD

	// Assign Battery Settings
	BATTERY_DRAIN_RATE = conf.BATTERY_DRAIN_RATE
	BATTERY_CHARGE_RATE = conf.BATTERY_CHARGE_RATE
	CRITICAL_BATTERY_LEVEL = uint8(conf.CRITICAL_BATTERY_LEVEL)
	LOW_BATTERY_LEVEL = uint8(conf.LOW_BATTERY_LEVEL)
	TARGET_RECHARGE_LEVEL = uint8(conf.TARGET_RECHARGE_LEVEL)

	// Assign Device Settings
	CAMERA_CHUNK_SIZE = conf.CAMERA_CHUNK_SIZE
	CAMERA_FAIL_CHANCE = conf.CAMERA_FAIL_CHANCE
	INSTALL_SUCCESS_CHANCE = conf.INSTALL_SUCCESS_CHANCE

	if print {
		PrintConfig()
	}
}

// PrintConfig prints the current configuration settings
func PrintConfig() {
	fmt.Println("=== Network Ports ===")
	fmt.Println("API_PORT:", API_PORT)
	fmt.Println("TCP_ID_PORT:", TCP_ID_PORT)
	fmt.Println("UDP_COMM_PORT:", UDP_COMM_PORT)
	fmt.Println("TCP_TELEMETRY_PORT:", TCP_TELEMETRY_PORT)
}

// GetMotherUDPAddr returns the full UDP address for communication
func GetMotherUDPAddr() string {
	return GlobalConfig.MotherIP + ":" + UDP_COMM_PORT
}

// GetMotherTCPIDAddr returns the full TCP address for ID assignment
func GetMotherTCPIDAddr() string {
	return GlobalConfig.MotherIP + ":" + TCP_ID_PORT
}

// GetMotherTelemetryAddr returns the full TCP address for telemetry
func GetMotherTelemetryAddr() string {
	return GlobalConfig.MotherIP + ":" + TCP_TELEMETRY_PORT
}

// IsTestMode returns true if test mode is enabled
func IsTestMode() bool {
	return GlobalConfig.TestMode
}
