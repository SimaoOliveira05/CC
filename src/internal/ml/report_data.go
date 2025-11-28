package ml

import (
    "encoding/binary"
    "fmt"
    "math"
)

const (
    // Task Types for Reports.
    TASK_IMAGE_CAPTURE     = 0
    TASK_SAMPLE_COLLECTION = 1
    TASK_ENV_ANALYSIS      = 2
    TASK_REPAIR_RESCUE     = 3
    TASK_TOPO_MAPPING      = 4
    TASK_INSTALLATION      = 5

    REPORT_HEADER_SIZE = 4 // 1 (TaskType) + 2 (MissionID) + 1 (IsLastReport)
)

// Generic Header for all reports.
type ReportHeader struct {
    TaskType     uint8      // Task type of the report
    MissionID    uint16     // Mission ID associated with the report
    IsLastReport bool       // Indicates if this is the last report in the sequence
}

// EncodeHeader serializes the ReportHeader into bytes.
func (h *ReportHeader) EncodeHeader() []byte {
    data := make([]byte, REPORT_HEADER_SIZE)
    data[0] = h.TaskType
    binary.BigEndian.PutUint16(data[1:3], h.MissionID)
    data[3] = boolToByte(h.IsLastReport)
    return data
}

// DecodeHeader deserializes bytes into ReportHeader.
func (h *ReportHeader) DecodeHeader(b []byte) {
    h.TaskType = b[0]
    h.MissionID = binary.BigEndian.Uint16(b[1:3])
    h.IsLastReport = b[3] == 1
}

// Report with generic payload.
type Report struct {
    Header  ReportHeader    // Report header
    Payload []byte          // Report payload
}

// Encode serializes the Report into bytes.
func (r *Report) Encode() []byte {
    data := make([]byte, REPORT_HEADER_SIZE+len(r.Payload))
    copy(data[0:REPORT_HEADER_SIZE], r.Header.EncodeHeader())
    copy(data[REPORT_HEADER_SIZE:], r.Payload)
    return data
}

// Decode deserializes bytes into Report.
func (r *Report) Decode(b []byte) error {

    r.Header.DecodeHeader(b[:REPORT_HEADER_SIZE])
    r.Payload = make([]byte, len(b)-REPORT_HEADER_SIZE)
    copy(r.Payload, b[REPORT_HEADER_SIZE:])
    return nil
}

// PayloadEncoder is an interface for all report data types.
// Each report data type must implement EncodePayload().
type PayloadEncoder interface {
    EncodePayload() []byte
}

// NewReport creates a generic Report from header and a PayloadEncoder.
func NewReport(taskType uint8, missionID uint16, isLast bool, encoder PayloadEncoder) *Report {
    return &Report{
        Header: ReportHeader{
            TaskType:     taskType,
            MissionID:    missionID,
            IsLastReport: isLast,
        },
        Payload: encoder.EncodePayload(),
    }
}

// DecodeTyped decodes the payload into the correct report data type.
func (r *Report) DecodeTyped() (any, error) {
    switch r.Header.TaskType {
    case TASK_IMAGE_CAPTURE:
        var img ImageReportData
        img.DecodePayload(r.Payload)
        return img, nil
    case TASK_SAMPLE_COLLECTION:
        var sample SampleReportData
        sample.DecodePayload(r.Payload)
        return sample, nil
    case TASK_ENV_ANALYSIS:
        var env EnvReportData
        env.DecodePayload(r.Payload)
        return env, nil
    case TASK_REPAIR_RESCUE:
        var rep RepairReportData
        rep.DecodePayload(r.Payload)
        return rep, nil
    case TASK_TOPO_MAPPING:
        var topo TopoReportData
        topo.DecodePayload(r.Payload)
        return topo, nil
    case TASK_INSTALLATION:
        var inst InstallReportData
        inst.DecodePayload(r.Payload)
        return inst, nil
    default:
        return nil, fmt.Errorf("unknown TaskType: %d", r.Header.TaskType)
    }
}

// String returns a human-readable representation of the Report.
func (r *Report) String() string {
    return fmt.Sprintf(
        "Report: Type=%d MissionID=%d IsLast=%v PayloadSize=%d",
        r.Header.TaskType,
        r.Header.MissionID,
        r.Header.IsLastReport,
        len(r.Payload),
    )
}

// GetMissionID returns the MissionID from the report header.
func (r *Report) GetMissionID() uint16 {
    return r.Header.MissionID
}

// IsLast returns true if the report is marked as the last report.
func (r *Report) IsLast() bool {
    return r.Header.IsLastReport
}

// ====== IMAGE CAPTURE DATA ======
type ImageReportData struct {
    ChunkID uint16      // ID of the image chunk
    Data    []byte      // Image data bytes
}

func (img *ImageReportData) EncodePayload() []byte {
    payload := make([]byte, 2+len(img.Data))
    binary.BigEndian.PutUint16(payload[0:2], img.ChunkID)
    copy(payload[2:], img.Data)
    return payload
}

func (img *ImageReportData) DecodePayload(payload []byte) {
    img.ChunkID = binary.BigEndian.Uint16(payload[0:2])
    img.Data = make([]byte, len(payload)-2)
    copy(img.Data, payload[2:])
}

// ====== SAMPLE COLLECTION DATA ======
// Component represents a single component in the sample.
type Component struct {
    Name       string       // Name of the chemical element
    Percentage float32      // Percentage of the component in the sample
}

// SampleReportData holds data for sample collection reports.
type SampleReportData struct {
    Components []Component
}

// EncodePayload serializes the SampleReportData into bytes.
func (s *SampleReportData) EncodePayload() []byte {
    totalLen := 1 // numComponents
    for _, c := range s.Components {
        totalLen += 1 + len(c.Name) + 4
    }
    payload := make([]byte, totalLen)
    payload[0] = uint8(len(s.Components))
    idx := 1
    for _, c := range s.Components {
        nameLen := uint8(len(c.Name))
        payload[idx] = nameLen
        idx++
        copy(payload[idx:idx+int(nameLen)], []byte(c.Name))
        idx += int(nameLen)
        binary.BigEndian.PutUint32(payload[idx:idx+4], math.Float32bits(c.Percentage))
        idx += 4
    }
    return payload
}

// DecodePayload deserializes bytes into SampleReportData.
func (s *SampleReportData) DecodePayload(payload []byte) {
    count := int(payload[0])
    s.Components = make([]Component, count)
    idx := 1
    for i := 0; i < count; i++ {
        nameLen := int(payload[idx])
        idx++
        s.Components[i].Name = string(payload[idx : idx+nameLen])
        idx += nameLen
        s.Components[i].Percentage = math.Float32frombits(binary.BigEndian.Uint32(payload[idx : idx+4]))
        idx += 4
    }
}

// ====== ENVIRONMENTAL ANALYSIS DATA ======
// EnvReportData holds data for environmental analysis reports.
type EnvReportData struct {
    Temp      float32      // Temperature in degrees Celsius
    Oxygen    float32      // Oxygen level percentage
    Pressure  float32      // Atmospheric pressure in hPa
    Humidity  float32      // Humidity percentage
    WindSpeed float32      // Wind speed in m/s
    Radiation float32      // Radiation level in µSv/h
}

// EncodePayload serializes the EnvReportData into bytes.
func (e *EnvReportData) EncodePayload() []byte {
    payload := make([]byte, 4*6)
    binary.BigEndian.PutUint32(payload[0:4], math.Float32bits(e.Temp))
    binary.BigEndian.PutUint32(payload[4:8], math.Float32bits(e.Oxygen))
    binary.BigEndian.PutUint32(payload[8:12], math.Float32bits(e.Pressure))
    binary.BigEndian.PutUint32(payload[12:16], math.Float32bits(e.Humidity))
    binary.BigEndian.PutUint32(payload[16:20], math.Float32bits(e.WindSpeed))
    binary.BigEndian.PutUint32(payload[20:24], math.Float32bits(e.Radiation))
    return payload
}

// DecodePayload deserializes bytes into EnvReportData.
func (e *EnvReportData) DecodePayload(payload []byte) {
    e.Temp = math.Float32frombits(binary.BigEndian.Uint32(payload[0:4]))
    e.Oxygen = math.Float32frombits(binary.BigEndian.Uint32(payload[4:8]))
    e.Pressure = math.Float32frombits(binary.BigEndian.Uint32(payload[8:12]))
    e.Humidity = math.Float32frombits(binary.BigEndian.Uint32(payload[12:16]))
    e.WindSpeed = math.Float32frombits(binary.BigEndian.Uint32(payload[16:20]))
    e.Radiation = math.Float32frombits(binary.BigEndian.Uint32(payload[20:24]))
}

// ====== REPAIR/RESCUE DATA ======
// RepairReportData holds data for repair/rescue reports.
type RepairReportData struct {
    ProblemID  uint8    // ID of the reported problem
    Repairable bool     // Indicates if the problem is repairable
}

// EncodePayload serializes the RepairReportData into bytes.
func (r *RepairReportData) EncodePayload() []byte {
    payload := make([]byte, 2)
    payload[0] = r.ProblemID
    payload[1] = boolToByte(r.Repairable)
    return payload
}

// DecodePayload deserializes bytes into RepairReportData.
func (r *RepairReportData) DecodePayload(payload []byte) {
    r.ProblemID = payload[0]
    r.Repairable = payload[1] == 1
}

// ====== TOPOGRAPHIC MAPPING DATA ======
// TopoReportData holds data for topographic mapping reports.
type TopoReportData struct {
    Latitude  float32   // Latitude of the mapped point
    Longitude float32   // Longitude of the mapped point
    Height    float32   // Height/elevation in meters
}

// EncodePayload serializes the TopoReportData into bytes.
func (t *TopoReportData) EncodePayload() []byte {
    payload := make([]byte, 12)
    binary.BigEndian.PutUint32(payload[0:4], math.Float32bits(t.Latitude))
    binary.BigEndian.PutUint32(payload[4:8], math.Float32bits(t.Longitude))
    binary.BigEndian.PutUint32(payload[8:12], math.Float32bits(t.Height))
    return payload
}

// DecodePayload deserializes bytes into TopoReportData.
func (t *TopoReportData) DecodePayload(payload []byte) {
    t.Latitude = math.Float32frombits(binary.BigEndian.Uint32(payload[0:4]))
    t.Longitude = math.Float32frombits(binary.BigEndian.Uint32(payload[4:8]))
    t.Height = math.Float32frombits(binary.BigEndian.Uint32(payload[8:12]))
}

// ====== INSTRUMENT INSTALLATION DATA ======
// InstallReportData holds data for instrument installation reports.
type InstallReportData struct {
    Success bool
}

// EncodePayload serializes the InstallReportData into bytes.
func (i *InstallReportData) EncodePayload() []byte {
    payload := make([]byte, 1)
    payload[0] = boolToByte(i.Success)
    return payload
}

// DecodePayload deserializes bytes into InstallReportData.
func (i *InstallReportData) DecodePayload(payload []byte) {
    i.Success = payload[0] == 1
}

// Helper to convert bool to byte.
func boolToByte(b bool) byte {
    if b {
        return 1
    }
    return 0
}