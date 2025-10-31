package ml

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"src/utils"
)

const (
	Image_Capture          = 0
	Object_Sample          = 1
	Environmental_Analysis = 2
	Rescue                 = 3
	Mapping                = 4
	Object_Installation    = 5
)

type Data struct {
	MsgID           uint16
	Coordinate      utils.Coordinate
	TaskType        uint8
	Duration        uint32
	UpdateFrequency uint32
	Priority        uint8
}

// ToBytes serializa a estrutura Data em bytes (BigEndian).
func (d *Data) ToBytes() []byte {
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
func DataFromBytes(data []byte) Data {
	var d Data
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
	case Image_Capture:
		return "Image_Capture"
	case Object_Sample:
		return "Object_Sample"
	case Environmental_Analysis:
		return "Environmental_Analysis"
	case Rescue:
		return "Rescue"
	case Mapping:
		return "Mapping"
	case Object_Installation:
		return "Object_Installation"
	default:
		return "Unknown"
	}
}

// String devolve uma representação legível do Data.
func (d Data) String() string {
	return fmt.Sprintf("Data{MsgID:%d, Coordinate:%s, TaskType:%s(%d), Duration:%d, UpdateFreq:%d, Priority:%d}",
		d.MsgID, d.Coordinate.String(), taskTypeName(d.TaskType), d.TaskType, d.Duration, d.UpdateFrequency, d.Priority)
}
