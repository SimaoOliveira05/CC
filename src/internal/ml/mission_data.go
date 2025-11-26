package ml

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"src/utils"
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
// TaskType e Priority são combinados num único byte: [4 bits TaskType | 4 bits Priority]
func (d *MissionData) ToBytes() []byte {
	buf := new(bytes.Buffer)
	// MsgID
	_ = binary.Write(buf, binary.BigEndian, d.MsgID)
	// Coordinate (Latitude, Longitude)
	_ = binary.Write(buf, binary.BigEndian, d.Coordinate.Latitude)
	_ = binary.Write(buf, binary.BigEndian, d.Coordinate.Longitude)
	// TaskType (4 bits superiores) + Priority (4 bits inferiores) em 1 byte
	taskTypeAndPriority := (d.TaskType << 4) | (d.Priority & 0x0F)
	_ = binary.Write(buf, binary.BigEndian, taskTypeAndPriority)
	// Duration
	_ = binary.Write(buf, binary.BigEndian, d.Duration)
	// UpdateFrequency
	_ = binary.Write(buf, binary.BigEndian, d.UpdateFrequency)
	return buf.Bytes()
}

// FromBytes desserializa bytes em Data (espera BigEndian e a mesma ordem usada em ToBytes).
// TaskType e Priority são extraídos de 1 byte combinado: [4 bits TaskType | 4 bits Priority]
func DataFromBytes(data []byte) MissionData {
	var d MissionData
	buf := bytes.NewReader(data)
	_ = binary.Read(buf, binary.BigEndian, &d.MsgID)
	_ = binary.Read(buf, binary.BigEndian, &d.Coordinate.Latitude)
	_ = binary.Read(buf, binary.BigEndian, &d.Coordinate.Longitude)
	// Lê o byte combinado e separa TaskType e Priority
	var taskTypeAndPriority uint8
	_ = binary.Read(buf, binary.BigEndian, &taskTypeAndPriority)
	d.TaskType = (taskTypeAndPriority >> 4) & 0x0F // 4 bits superiores
	d.Priority = taskTypeAndPriority & 0x0F        // 4 bits inferiores
	_ = binary.Read(buf, binary.BigEndian, &d.Duration)
	_ = binary.Read(buf, binary.BigEndian, &d.UpdateFrequency)
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
