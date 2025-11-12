package ml

import (
	"src/utils"
	"bytes"
	"encoding/binary"
	"fmt"
)

type MissionData struct {
	MsgID           uint16
	Coordinate      utils.Coordinate
	TaskType        uint8
	Duration        uint32
	UpdateFrequency uint32
	Priority        uint8
}


// ToBytes serializa a estrutura Data em bytes (BigEndian).
func (d *MissionData) ToBytes() []byte {
	buf := new(bytes.Buffer)
	// MsgID
	_ = binary.Write(buf, binary.BigEndian, d.MsgID)
	// Coordinate (Latitude, Longitude)
	_ = binary.Write(buf, binary.BigEndian, d.Coordinate.Latitude)
	_ = binary.Write(buf, binary.BigEndian, d.Coordinate.Longitude)
	// TaskType
	_ = binary.Write(buf, binary.BigEndian, d.TaskType)
	// Duration
	_ = binary.Write(buf, binary.BigEndian, d.Duration)
	// UpdateFrequency
	_ = binary.Write(buf, binary.BigEndian, d.UpdateFrequency)
	// Priority
	_ = binary.Write(buf, binary.BigEndian, d.Priority)
	return buf.Bytes()
}

// FromBytes desserializa bytes em Data (espera BigEndian e a mesma ordem usada em ToBytes).
func DataFromBytes(data []byte) MissionData {
	var d MissionData
	buf := bytes.NewReader(data)
	_ = binary.Read(buf, binary.BigEndian, &d.MsgID)
	_ = binary.Read(buf, binary.BigEndian, &d.Coordinate.Latitude)
	_ = binary.Read(buf, binary.BigEndian, &d.Coordinate.Longitude)
	_ = binary.Read(buf, binary.BigEndian, &d.TaskType)
	_ = binary.Read(buf, binary.BigEndian, &d.Duration)
	_ = binary.Read(buf, binary.BigEndian, &d.UpdateFrequency)
	_ = binary.Read(buf, binary.BigEndian, &d.Priority)
	return d
}

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

// String devolve uma representação legível do Data.
func (d MissionData) String() string {
	return fmt.Sprintf("Data{MsgID:%d, Coordinate:%s, TaskType:%s(%d), Duration:%d, UpdateFreq:%d, Priority:%d}",
		d.MsgID, d.Coordinate.String(), taskTypeName(d.TaskType), d.TaskType, d.Duration, d.UpdateFrequency, d.Priority)
}

type ReportData struct {
	EndMission bool
	ReportInfo Report
}

