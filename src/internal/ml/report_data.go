package ml

import (
	"encoding/binary"
	"fmt"
	"math"
)

const (
	// Task types .
	TASK_IMAGE_CAPTURE     = 0
	TASK_SAMPLE_COLLECTION = 1
	TASK_ENV_ANALYSIS      = 2
	TASK_REPAIR_RESCUE     = 3
	TASK_TOPO_MAPPING      = 4
	TASK_INSTALLATION      = 5
	// Fixed sizes of binary reports.
	IMAGE_REPORT_HEADER_SIZE = 6
	REPAIR_REPORT_SIZE       = 6
	INSTALL_REPORT_SIZE      = 5
	TOPO_REPORT_SIZE         = 16
	ENV_REPORT_SIZE          = 28
)

// Report is the common interface for all rover report types.
type Report interface {
	Encode() []byte
	Decode([]byte) error
	GetTaskType() uint8
	String() string
	IsLast() bool
	GetMissionID() uint16
}

// ====== IMAGE CAPTURE REPORT ======
//
// ImageReport is a partial image (chunk) report.
type ImageReport struct {
	TaskType     uint8  `json:"taskType"`     // Type of the report (always TASK_IMAGE_CAPTURE)
	MissionID    uint16 `json:"missionId"`    // Unique mission identifier
	ChunkID      uint16 `json:"chunkId"`      // Image chunk identifier
	Data         []byte `json:"data"`         // Image chunk bytes
	IsLastReport bool   `json:"isLastReport"` // True if this is the last chunk for the mission
}

// GetMissionID returns the mission ID for the image report.
func (r *ImageReport) GetMissionID() uint16 { return r.MissionID }

// IsLast returns true if this is the last image report.
func (r *ImageReport) IsLast() bool { return r.IsLastReport }

// Encode serializes the image report to binary format.
func (r *ImageReport) Encode() []byte {
	data := make([]byte, IMAGE_REPORT_HEADER_SIZE+len(r.Data))
	data[0] = r.TaskType
	binary.BigEndian.PutUint16(data[1:3], r.MissionID)
	binary.BigEndian.PutUint16(data[3:5], r.ChunkID)
	data[5] = boolToByte(r.IsLastReport)
	copy(data[IMAGE_REPORT_HEADER_SIZE:], r.Data)
	return data
}

// Decode deserializes binary data into the image report.
func (r *ImageReport) Decode(b []byte) error {
	r.TaskType = b[0]
	r.MissionID = binary.BigEndian.Uint16(b[1:3])
	r.ChunkID = binary.BigEndian.Uint16(b[3:5])
	r.IsLastReport = b[5] == 1
	r.Data = make([]byte, len(b)-IMAGE_REPORT_HEADER_SIZE)
	copy(r.Data, b[IMAGE_REPORT_HEADER_SIZE:])
	return nil
}

// GetTaskType returns the task type for image report.
func (r *ImageReport) GetTaskType() uint8 { return TASK_IMAGE_CAPTURE }

// String returns a human-readable summary of the image report.
func (r *ImageReport) String() string {
	return fmt.Sprintf("[Image] Mission %d - Chunk %d (%d bytes)", r.MissionID, r.ChunkID, len(r.Data))
}

//
// ====== Sample collection ======
//

// Component is a chemical component with name and percentage.
type Component struct {
	Name       string  `json:"name"`       // Chemical name (e.g., "O2", "CO2", "H2O")
	Percentage float32 `json:"percentage"` // Percentage (0.0 to 100.0)
}

// SampleReport is a chemical components report.
type SampleReport struct {
	TaskType     uint8       `json:"taskType"`     // Type of the report (always TASK_SAMPLE_COLLECTION)
	MissionID    uint16      `json:"missionId"`    // Unique mission identifier
	NumSamples   uint8       `json:"numSamples"`   // Number of chemical components
	Components   []Component `json:"components"`   // List of chemical components
	IsLastReport bool        `json:"isLastReport"` // True if this is the last sample report for the mission
}

// GetMissionID returns the mission ID for the sample report.
func (r *SampleReport) GetMissionID() uint16 { return r.MissionID }

// IsLast returns true if this is the last sample report.
func (r *SampleReport) IsLast() bool { return r.IsLastReport }

// Encode serializes the sample report to binary format.
func (r *SampleReport) Encode() []byte {
	// Calculate total length
	totalLen := 5
	for _, c := range r.Components {
		totalLen += 1 + len(c.Name) + 4
	}
	data := make([]byte, totalLen)
	data[0] = r.TaskType
	binary.BigEndian.PutUint16(data[1:3], r.MissionID)
	data[3] = r.NumSamples
	data[4] = boolToByte(r.IsLastReport)
	idx := 5
	for _, c := range r.Components {
		nameLen := uint8(len(c.Name))
		data[idx] = nameLen
		idx++
		copy(data[idx:idx+int(nameLen)], []byte(c.Name))
		idx += int(nameLen)
		binary.BigEndian.PutUint32(data[idx:idx+4], math.Float32bits(c.Percentage))
		idx += 4
	}
	return data
}

// Decode deserializes binary data into the sample report.
func (r *SampleReport) Decode(b []byte) error {
	if len(b) < 5 {
		return fmt.Errorf("report too short")
	}
	r.TaskType = b[0]
	r.MissionID = binary.BigEndian.Uint16(b[1:3])
	r.NumSamples = b[3]
	r.IsLastReport = b[4] == 1
	count := int(r.NumSamples)
	r.Components = make([]Component, count)
	idx := 5
	for i := 0; i < count; i++ {
		if idx >= len(b) {
			return fmt.Errorf("insufficient data for component")
		}
		nameLen := int(b[idx])
		idx++
		if idx+nameLen+4 > len(b) {
			return fmt.Errorf("insufficient data for name or percentage")
		}
		r.Components[i].Name = string(b[idx : idx+nameLen])
		idx += nameLen
		r.Components[i].Percentage = math.Float32frombits(binary.BigEndian.Uint32(b[idx : idx+4]))
		idx += 4
	}
	return nil
}

// GetTaskType returns the task type for sample report.
func (r *SampleReport) GetTaskType() uint8 { return TASK_SAMPLE_COLLECTION }

// String returns a human-readable summary of the sample report.
func (r *SampleReport) String() string {
	compStr := ""
	for i, c := range r.Components {
		if i > 0 {
			compStr += ", "
		}
		compStr += fmt.Sprintf("%s=%.2f%%", c.Name, c.Percentage)
	}
	return fmt.Sprintf("[Sample] Mission %d - %d components [%s]", r.MissionID, r.NumSamples, compStr)
}

//
// ====== ENVIRONMENTAL ANALYSIS REPORT ======
//

// EnvReport contains atmospheric measurements.
type EnvReport struct {
	TaskType     uint8   `json:"taskType"`     // Type of the report (always TASK_ENV_ANALYSIS)
	MissionID    uint16  `json:"missionId"`    // Unique mission identifier
	Temp         float32 `json:"temp"`         // Temperature (Celsius)
	Oxygen       float32 `json:"oxygen"`       // Oxygen percentage
	Pressure     float32 `json:"pressure"`     // Atmospheric pressure
	Humidity     float32 `json:"humidity"`     // Humidity percentage
	WindSpeed    float32 `json:"windSpeed"`    // Wind speed
	Radiation    float32 `json:"radiation"`    // Radiation level
	IsLastReport bool    `json:"isLastReport"` // True if this is the last environment report for the mission
}

// GetMissionID returns the mission ID for the environment report.
func (r *EnvReport) GetMissionID() uint16 { return r.MissionID }

// IsLast returns true if this is the last environment report.
func (r *EnvReport) IsLast() bool { return r.IsLastReport }

// Encode serializes the environment report to binary format.
func (r *EnvReport) Encode() []byte {
	data := make([]byte, ENV_REPORT_SIZE)
	data[0] = r.TaskType
	binary.BigEndian.PutUint16(data[1:3], r.MissionID)
	data[3] = boolToByte(r.IsLastReport)
	binary.BigEndian.PutUint32(data[4:8], math.Float32bits(r.Temp))
	binary.BigEndian.PutUint32(data[8:12], math.Float32bits(r.Oxygen))
	binary.BigEndian.PutUint32(data[12:16], math.Float32bits(r.Pressure))
	binary.BigEndian.PutUint32(data[16:20], math.Float32bits(r.Humidity))
	binary.BigEndian.PutUint32(data[20:24], math.Float32bits(r.WindSpeed))
	binary.BigEndian.PutUint32(data[24:28], math.Float32bits(r.Radiation))
	return data
}

// Decode deserializes binary data into the environment report.
func (r *EnvReport) Decode(b []byte) error {
	if len(b) < ENV_REPORT_SIZE {
		return fmt.Errorf("report too short")
	}
	r.TaskType = b[0]
	r.MissionID = binary.BigEndian.Uint16(b[1:3])
	r.IsLastReport = b[3] == 1
	r.Temp = math.Float32frombits(binary.BigEndian.Uint32(b[4:8]))
	r.Oxygen = math.Float32frombits(binary.BigEndian.Uint32(b[8:12]))
	r.Pressure = math.Float32frombits(binary.BigEndian.Uint32(b[12:16]))
	r.Humidity = math.Float32frombits(binary.BigEndian.Uint32(b[16:20]))
	r.WindSpeed = math.Float32frombits(binary.BigEndian.Uint32(b[20:24]))
	r.Radiation = math.Float32frombits(binary.BigEndian.Uint32(b[24:28]))
	return nil
}

// GetTaskType returns the task type for environment report.
func (r *EnvReport) GetTaskType() uint8 { return TASK_ENV_ANALYSIS }

// String returns a human-readable summary of the environment report.
func (r *EnvReport) String() string {
	return fmt.Sprintf("[Environment] Mission %d - T=%.2f°C, O2=%.2f%%, P=%.2fhPa, H=%.2f%%, V=%.2fm/s, R=%.2fµSv",
		r.MissionID, r.Temp, r.Oxygen, r.Pressure, r.Humidity, r.WindSpeed, r.Radiation)
}

//
// ====== REPAIR/RESCUE REPORT ======
//

// RepairReport is the result of a repair attempt.
type RepairReport struct {
	TaskType     uint8  `json:"taskType"`     // Type of the report (always TASK_REPAIR_RESCUE)
	MissionID    uint16 `json:"missionId"`    // Unique mission identifier
	ProblemID    uint8  `json:"problemId"`    // Problem identifier
	Repairable   bool   `json:"repairable"`   // True if the problem was repaired
	IsLastReport bool   `json:"isLastReport"` // True if this is the last repair report for the mission
}

// GetMissionID returns the mission ID for the repair report.
func (r *RepairReport) GetMissionID() uint16 {
	return r.MissionID
}

// IsLast returns true if this is the last repair report.
func (r *RepairReport) IsLast() bool {
	return r.IsLastReport
}

// Encode serializes the repair report to binary format.
func (r *RepairReport) Encode() []byte {
	data := make([]byte, REPAIR_REPORT_SIZE)
	data[0] = r.TaskType
	binary.BigEndian.PutUint16(data[1:3], r.MissionID)
	data[3] = r.ProblemID
	data[4] = boolToByte(r.IsLastReport)
	data[5] = boolToByte(r.Repairable)
	return data
}

// Decode deserializes binary data into the repair report.
func (r *RepairReport) Decode(b []byte) error {
	if len(b) < REPAIR_REPORT_SIZE {
		return fmt.Errorf("report too short")
	}
	r.TaskType = b[0]
	r.MissionID = binary.BigEndian.Uint16(b[1:3])
	r.ProblemID = b[3]
	r.IsLastReport = b[4] == 1
	r.Repairable = b[5] == 1
	return nil
}

// GetTaskType returns the task type for repair report.
func (r *RepairReport) GetTaskType() uint8 { return TASK_REPAIR_RESCUE }

// String returns a human-readable summary of the repair report.
func (r *RepairReport) String() string {
	status := "not repairable"
	if r.Repairable {
		status = "repaired"
	}
	return fmt.Sprintf("[Repair] Mission %d - Problem %d (%s)", r.MissionID, r.ProblemID, status)
}

//
// ====== TOPOGRAPHIC MAPPING REPORT ======
//

// TopoReport is a topographic point (coordinate and height).
type TopoReport struct {
	TaskType     uint8   `json:"taskType"`     // Type of the report (always TASK_TOPO_MAPPING)
	MissionID    uint16  `json:"missionId"`    // Unique mission identifier
	Latitude     float32 `json:"latitude"`     // Latitude coordinate
	Longitude    float32 `json:"longitude"`    // Longitude coordinate
	Height       float32 `json:"height"`       // Height value
	IsLastReport bool    `json:"isLastReport"` // True if this is the last topographic report for the mission
}

// GetMissionID returns the mission ID for the topographic report.
func (r *TopoReport) GetMissionID() uint16 {
	return r.MissionID
}

// IsLast returns true if this is the last topographic report.
func (r *TopoReport) IsLast() bool {
	return r.IsLastReport
}

// Encode serializes the topographic report to binary format.
func (r *TopoReport) Encode() []byte {
	data := make([]byte, TOPO_REPORT_SIZE)
	data[0] = r.TaskType
	binary.BigEndian.PutUint16(data[1:3], r.MissionID)
	data[3] = boolToByte(r.IsLastReport)
	binary.BigEndian.PutUint32(data[4:8], math.Float32bits(r.Latitude))
	binary.BigEndian.PutUint32(data[8:12], math.Float32bits(r.Longitude))
	binary.BigEndian.PutUint32(data[12:16], math.Float32bits(r.Height))
	return data
}

// Decode deserializes binary data into the topographic report.
func (r *TopoReport) Decode(b []byte) error {
	if len(b) < TOPO_REPORT_SIZE {
		return fmt.Errorf("report too short")
	}
	r.TaskType = b[0]
	r.MissionID = binary.BigEndian.Uint16(b[1:3])
	r.IsLastReport = b[3] == 1
	r.Latitude = math.Float32frombits(binary.BigEndian.Uint32(b[4:8]))
	r.Longitude = math.Float32frombits(binary.BigEndian.Uint32(b[8:12]))
	r.Height = math.Float32frombits(binary.BigEndian.Uint32(b[12:16]))
	return nil
}

// GetTaskType returns the task type for topographic report.
func (r *TopoReport) GetTaskType() uint8 { return TASK_TOPO_MAPPING }

// String returns a human-readable summary of the topographic report.
func (r *TopoReport) String() string {
	return fmt.Sprintf("[Topography] Mission %d - (%.4f, %.4f) h=%.2fm", r.MissionID, r.Latitude, r.Longitude, r.Height)
}

//
// ====== INSTRUMENT INSTALLATION REPORT ======
//

// InstallReport indicates success/failure of installation.
type InstallReport struct {
	TaskType     uint8  `json:"taskType"`     // Type of the report (always TASK_INSTALLATION)
	MissionID    uint16 `json:"missionId"`    // Unique mission identifier
	Success      bool   `json:"success"`      // True if installation succeeded
	IsLastReport bool   `json:"isLastReport"` // True if this is the last installation report for the mission
}

// GetMissionID returns the mission ID for the installation report.
func (r *InstallReport) GetMissionID() uint16 { return r.MissionID }

// IsLast returns true if this is the last installation report.
func (r *InstallReport) IsLast() bool { return r.IsLastReport }

// Encode serializes the installation report to binary format.
func (r *InstallReport) Encode() []byte {
	data := make([]byte, INSTALL_REPORT_SIZE)
	data[0] = r.TaskType
	binary.BigEndian.PutUint16(data[1:3], r.MissionID)
	data[3] = boolToByte(r.IsLastReport)
	data[4] = boolToByte(r.Success)
	return data
}

// Decode deserializes binary data into the installation report.
func (r *InstallReport) Decode(b []byte) error {
	if len(b) < INSTALL_REPORT_SIZE {
		return fmt.Errorf("report too short")
	}
	r.TaskType = b[0]
	r.MissionID = binary.BigEndian.Uint16(b[1:3])
	r.IsLastReport = b[3] == 1
	r.Success = b[4] == 1
	return nil
}

// GetTaskType returns the task type for installation report.
func (r *InstallReport) GetTaskType() uint8 { return TASK_INSTALLATION }

// String returns a human-readable summary of the installation report.
func (r *InstallReport) String() string {
	if r.Success {
		return fmt.Sprintf("[Installation] Mission %d - completed successfully", r.MissionID)
	}
	return fmt.Sprintf("[Installation] Mission %d - failed", r.MissionID)
}

func boolToByte(b bool) byte {
	if b {
		return 1
	}
	return 0
}
