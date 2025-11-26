package ml

import (
	"encoding/binary"
	"fmt"
	"src/utils"
	"math"
)

// MissionData is the data structure for representing mission details.
type MissionData struct {
	MsgID           uint16           // Message unique identifier [0-65535]
	Coordinate      utils.Coordinate //Coordinate where the mission must be performed [Latitude:float64, Longitude:float64]
	TaskType        uint8            // Type of task to be performed [0-15 representing different tasks]
	Duration        uint32           // Duration of the mission in seconds
	UpdateFrequency uint32           // Frequency at which mission updates should be sent [0 seconds - aproximately 136 years]
	Priority        uint8            // Priority level of the mission [0-15]
}

//MissionDataSize is the size in bytes of the MissionData struct when serialized.
const MissionDataSize = 27 // 2 (MsgID) + 8 (Latitude) + 8 (Longitude) + 1 (TaskType + Priority) + 4 (Duration) + 4 (UpdateFrequency)

// Enconde serializes the Data into bytes (BigEndian).
func (d *MissionData) Encode() []byte {
    data := make([]byte, MissionDataSize)
    
    binary.BigEndian.PutUint16(data[0:], d.MsgID)
    binary.BigEndian.PutUint64(data[2:], math.Float64bits(d.Coordinate.Latitude))
    binary.BigEndian.PutUint64(data[10:], math.Float64bits(d.Coordinate.Longitude))
	// taskTypeAndPriority combines TaskType (4 higher bits) and Priority
    data[18] = (d.TaskType << 4) | (d.Priority & 0x0F)
    binary.BigEndian.PutUint32(data[19:], d.Duration)
    binary.BigEndian.PutUint32(data[23:], d.UpdateFrequency)
    
    return data
}

// Decode desserialize the bytes into Data (BigEndian). This function expects the same order used in Encode.
func (d *MissionData) Decode(data []byte) MissionData {
    return MissionData{
        MsgID: binary.BigEndian.Uint16(data[0:]),
        Coordinate: utils.Coordinate{
            Latitude:  math.Float64frombits(binary.BigEndian.Uint64(data[2:])),
            Longitude: math.Float64frombits(binary.BigEndian.Uint64(data[10:])),
        },
		// taskTypeAndPriority combines TaskType (4 higher bits) and Priority (4 lower bits) into 1 byte
        TaskType:        (data[18] >> 4) & 0x0F,
        Priority:        data[18] & 0x0F,
        Duration:        binary.BigEndian.Uint32(data[19:]),
        UpdateFrequency: binary.BigEndian.Uint32(data[23:]),
    }
}

// taskTypeName converts a task type id into a human-readable string.
func taskTypeName(t uint8) string {
	switch t {
	case TASK_IMAGE_CAPTURE:
		return "Image_Capture"
	case TASK_SAMPLE_COLLECTION:
		return "Object_Sample"
	case TASK_ENV_ANALYSIS:
		return "Environmental_Analysis"
	case TASK_REPAIR_RESCUE:
		return "Rescue"
	case TASK_TOPO_MAPPING:
		return "Mapping"
	case TASK_INSTALLATION:
		return "Object_Installation"
	default:
		return "Unknown"
	}
}

// String returns a human-readable representation of the MissionData.
func (d MissionData) String() string {
	return fmt.Sprintf("Data{MsgID:%d, Coordinate:%s, TaskType:%s(%d), Duration:%d, UpdateFreq:%d, Priority:%d}",
		d.MsgID, d.Coordinate.String(), taskTypeName(d.TaskType), d.TaskType, d.Duration, d.UpdateFrequency, d.Priority)
}