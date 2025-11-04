package ml

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

// Tipos de tarefas (3 bits no protocolo)
const (
	TASK_IMAGE_CAPTURE     = 0
	TASK_SAMPLE_COLLECTION = 1
	TASK_ENV_ANALYSIS      = 2
	TASK_REPAIR_RESCUE     = 3
	TASK_TOPO_MAPPING      = 4
	TASK_INSTALLATION      = 5
)

// Report define a interface comum a todos os tipos de relatório enviados pelos rovers
type Report interface {
	ToBytes() []byte
	FromBytes([]byte) error
	GetTaskType() uint8
	String() string
	IsLast() bool
	GetMissionID() uint16
	Clone() Report
}

//
// ====== CAPTURA DE IMAGEM ======
//

// ImageReport representa um relatório parcial de imagem (chunk)
type ImageReport struct {
	TaskType     uint8
	MissionID    uint16
	ChunkID      uint16
	Data         []byte // bytes da imagem (parcial)
	IsLastReport bool   // indica se é o último report
}

func (r *ImageReport) Clone() Report {
	dataCopy := make([]byte, len(r.Data))
	copy(dataCopy, r.Data)
	return &ImageReport{
		TaskType:     r.TaskType,
		MissionID:    r.MissionID,
		ChunkID:      r.ChunkID,
		Data:         dataCopy,
		IsLastReport: r.IsLastReport,
	}
}

func (r *ImageReport) GetMissionID() uint16 {
	return r.MissionID
}

func (r *ImageReport) IsLast() bool {
	return r.IsLastReport
}
func (r *ImageReport) ToBytes() []byte {
	buf := new(bytes.Buffer)
	binary.Write(buf, binary.BigEndian, r.TaskType)
	binary.Write(buf, binary.BigEndian, r.MissionID)
	binary.Write(buf, binary.BigEndian, r.ChunkID)
	buf.Write(r.Data)
	var last uint8
	if r.IsLastReport {
		last = 1
	}
	binary.Write(buf, binary.BigEndian, last)
	return buf.Bytes()
}

func (r *ImageReport) FromBytes(b []byte) error {
	if len(b) < 6 {
		return fmt.Errorf("report demasiado curto")
	}
	r.TaskType = b[0]
	r.MissionID = binary.BigEndian.Uint16(b[1:3])
	r.ChunkID = binary.BigEndian.Uint16(b[3:5])
	r.Data = b[5 : len(b)-1]
	r.IsLastReport = b[len(b)-1] == 1
	return nil
}

func (r *ImageReport) GetTaskType() uint8 { return TASK_IMAGE_CAPTURE }
func (r *ImageReport) String() string {
	return fmt.Sprintf("[Imagem] Missão %d - Chunk %d (%d bytes)", r.MissionID, r.ChunkID, len(r.Data))
}

//
// ====== COLETA DE AMOSTRAS ======
//

// Component representa um componente químico com nome e percentagem
type Component struct {
	Name       string  // nome do componente (ex: "O2", "CO2", "H2O")
	Percentage float32 // percentagem (0.0 a 100.0)
}

// SampleReport representa um relatório de componentes químicos (t=0)
type SampleReport struct {
	TaskType     uint8
	MissionID    uint16
	NumSamples   uint8
	Components   []Component // lista de (nome, %)
	IsLastReport bool        // indica se é o último report
}

func (r *SampleReport) Clone() Report {
	compsCopy := make([]Component, len(r.Components))
	copy(compsCopy, r.Components)
	return &SampleReport{
		TaskType:     r.TaskType,
		MissionID:    r.MissionID,
		NumSamples:   r.NumSamples,
		Components:   compsCopy,
		IsLastReport: r.IsLastReport,
	}
}

func (r *SampleReport) GetMissionID() uint16 {
	return r.MissionID
}

func (r *SampleReport) IsLast() bool {
	return r.IsLastReport
}
func (r *SampleReport) ToBytes() []byte {
	buf := new(bytes.Buffer)
	binary.Write(buf, binary.BigEndian, r.TaskType)
	binary.Write(buf, binary.BigEndian, r.MissionID)
	binary.Write(buf, binary.BigEndian, r.NumSamples)
	for _, c := range r.Components {
		nameLen := uint8(len(c.Name))
		binary.Write(buf, binary.BigEndian, nameLen)
		buf.WriteString(c.Name)
		binary.Write(buf, binary.BigEndian, c.Percentage)
	}
	var last uint8
	if r.IsLastReport {
		last = 1
	}
	binary.Write(buf, binary.BigEndian, last)
	return buf.Bytes()
}

func (r *SampleReport) FromBytes(b []byte) error {
	if len(b) < 5 {
		return fmt.Errorf("report demasiado curto")
	}
	r.TaskType = b[0]
	r.MissionID = binary.BigEndian.Uint16(b[1:3])
	r.NumSamples = b[3]
	count := int(r.NumSamples)
	r.Components = make([]Component, count)

	buf := bytes.NewReader(b[4 : len(b)-1])
	for i := 0; i < count; i++ {
		var nameLen uint8
		err := binary.Read(buf, binary.BigEndian, &nameLen)
		if err != nil {
			return fmt.Errorf("erro ao ler tamanho do nome: %w", err)
		}
		nameBytes := make([]byte, nameLen)
		_, err = buf.Read(nameBytes)
		if err != nil {
			return fmt.Errorf("erro ao ler nome: %w", err)
		}
		r.Components[i].Name = string(nameBytes)
		err = binary.Read(buf, binary.BigEndian, &r.Components[i].Percentage)
		if err != nil {
			return fmt.Errorf("erro ao ler percentagem: %w", err)
		}
	}
	r.IsLastReport = b[len(b)-1] == 1
	return nil
}

func (r *SampleReport) GetTaskType() uint8 { return TASK_SAMPLE_COLLECTION }
func (r *SampleReport) String() string {
	compStr := ""
	for i, c := range r.Components {
		if i > 0 {
			compStr += ", "
		}
		compStr += fmt.Sprintf("%s=%.2f%%", c.Name, c.Percentage)
	}
	return fmt.Sprintf("[Amostra] Missão %d - %d componentes [%s]", r.MissionID, r.NumSamples, compStr)
}

//
// ====== ANÁLISE AMBIENTAL ======
//

// EnvReport representa medições atmosféricas (T, O2, P, humidade, vento, radiação)
type EnvReport struct {
	TaskType     uint8
	MissionID    uint16
	Temp         float32
	Oxygen       float32
	Pressure     float32
	Humidity     float32
	WindSpeed    float32
	Radiation    float32
	IsLastReport bool // indica se é o último report
}

func (r *EnvReport) Clone() Report {
	return &EnvReport{
		TaskType:     r.TaskType,
		MissionID:    r.MissionID,
		Temp:         r.Temp,
		Oxygen:       r.Oxygen,
		Pressure:     r.Pressure,
		Humidity:     r.Humidity,
		WindSpeed:    r.WindSpeed,
		Radiation:    r.Radiation,
		IsLastReport: r.IsLastReport,
	}
}

func (r *EnvReport) GetMissionID() uint16 {
	return r.MissionID
}

func (r *EnvReport) IsLast() bool {
	return r.IsLastReport
}
func (r *EnvReport) ToBytes() []byte {
	buf := new(bytes.Buffer)
	binary.Write(buf, binary.BigEndian, r.TaskType)
	binary.Write(buf, binary.BigEndian, r.MissionID)
	binary.Write(buf, binary.BigEndian, r.Temp)
	binary.Write(buf, binary.BigEndian, r.Oxygen)
	binary.Write(buf, binary.BigEndian, r.Pressure)
	binary.Write(buf, binary.BigEndian, r.Humidity)
	binary.Write(buf, binary.BigEndian, r.WindSpeed)
	binary.Write(buf, binary.BigEndian, r.Radiation)
	var last uint8
	if r.IsLastReport {
		last = 1
	}
	binary.Write(buf, binary.BigEndian, last)
	return buf.Bytes()
}

func (r *EnvReport) FromBytes(b []byte) error {
	if len(b) < 28 {
		return fmt.Errorf("report demasiado curto")
	}
	r.TaskType = b[0]
	r.MissionID = binary.BigEndian.Uint16(b[1:3])
	buf := bytes.NewReader(b[3:27])
	binary.Read(buf, binary.BigEndian, &r.Temp)
	binary.Read(buf, binary.BigEndian, &r.Oxygen)
	binary.Read(buf, binary.BigEndian, &r.Pressure)
	binary.Read(buf, binary.BigEndian, &r.Humidity)
	binary.Read(buf, binary.BigEndian, &r.WindSpeed)
	binary.Read(buf, binary.BigEndian, &r.Radiation)
	r.IsLastReport = b[len(b)-1] == 1
	return nil
}

func (r *EnvReport) GetTaskType() uint8 { return TASK_ENV_ANALYSIS }
func (r *EnvReport) String() string {
	return fmt.Sprintf("[Ambiente] Missão %d - T=%.2f°C, O2=%.2f%%, P=%.2fhPa, H=%.2f%%, V=%.2fm/s, R=%.2fµSv",
		r.MissionID, r.Temp, r.Oxygen, r.Pressure, r.Humidity, r.WindSpeed, r.Radiation)
}

//
// ====== REPARAÇÃO / RESGATE ======
//

// RepairReport representa o resultado de uma tentativa de reparação (t=0)
type RepairReport struct {
	TaskType     uint8
	MissionID    uint16
	ProblemID    uint8
	Repairable   bool
	IsLastReport bool // indica se é o último report
}

func (r *RepairReport) Clone() Report {
	return &RepairReport{
		TaskType:     r.TaskType,
		MissionID:    r.MissionID,
		ProblemID:    r.ProblemID,
		Repairable:   r.Repairable,
		IsLastReport: r.IsLastReport,
	}
}

func (r *RepairReport) GetMissionID() uint16 {
	return r.MissionID
}

func (r *RepairReport) IsLast() bool {
	return r.IsLastReport
}
func (r *RepairReport) ToBytes() []byte {
	buf := new(bytes.Buffer)
	binary.Write(buf, binary.BigEndian, r.TaskType)
	binary.Write(buf, binary.BigEndian, r.MissionID)
	binary.Write(buf, binary.BigEndian, r.ProblemID)
	var flag uint8
	if r.Repairable {
		flag = 1
	}
	binary.Write(buf, binary.BigEndian, flag)
	var last uint8
	if r.IsLastReport {
		last = 1
	}
	binary.Write(buf, binary.BigEndian, last)
	return buf.Bytes()
}

func (r *RepairReport) FromBytes(b []byte) error {
	if len(b) < 6 {
		return fmt.Errorf("report demasiado curto")
	}
	r.TaskType = b[0]
	r.MissionID = binary.BigEndian.Uint16(b[1:3])
	r.ProblemID = b[3]
	r.Repairable = b[4] == 1
	r.IsLastReport = b[len(b)-1] == 1
	return nil
}

func (r *RepairReport) GetTaskType() uint8 { return TASK_REPAIR_RESCUE }
func (r *RepairReport) String() string {
	status := "não reparável"
	if r.Repairable {
		status = "reparado"
	}
	return fmt.Sprintf("[Reparação] Missão %d - Problema %d (%s)", r.MissionID, r.ProblemID, status)
}

//
// ====== MAPEAMENTO TOPOGRÁFICO ======
//

// TopoReport representa um ponto topográfico (coordenada e altura)
type TopoReport struct {
	TaskType     uint8
	MissionID    uint16
	Latitude     float32
	Longitude    float32
	Height       float32
	IsLastReport bool // indica se é o último report
}

func (r *TopoReport) Clone() Report {
	return &TopoReport{
		TaskType:     r.TaskType,
		MissionID:    r.MissionID,
		Latitude:     r.Latitude,
		Longitude:    r.Longitude,
		Height:       r.Height,
		IsLastReport: r.IsLastReport,
	}
}

func (r *TopoReport) GetMissionID() uint16 {
	return r.MissionID
}

func (r *TopoReport) IsLast() bool {
	return r.IsLastReport
}
func (r *TopoReport) ToBytes() []byte {
	buf := new(bytes.Buffer)
	binary.Write(buf, binary.BigEndian, r.TaskType)
	binary.Write(buf, binary.BigEndian, r.MissionID)
	binary.Write(buf, binary.BigEndian, r.Latitude)
	binary.Write(buf, binary.BigEndian, r.Longitude)
	binary.Write(buf, binary.BigEndian, r.Height)
	var last uint8
	if r.IsLastReport {
		last = 1
	}
	binary.Write(buf, binary.BigEndian, last)
	return buf.Bytes()
}

func (r *TopoReport) FromBytes(b []byte) error {
	if len(b) < 16 {
		return fmt.Errorf("report demasiado curto")
	}
	r.TaskType = b[0]
	r.MissionID = binary.BigEndian.Uint16(b[1:3])
	buf := bytes.NewReader(b[3:15])
	binary.Read(buf, binary.BigEndian, &r.Latitude)
	binary.Read(buf, binary.BigEndian, &r.Longitude)
	binary.Read(buf, binary.BigEndian, &r.Height)
	r.IsLastReport = b[len(b)-1] == 1
	return nil
}

func (r *TopoReport) GetTaskType() uint8 { return TASK_TOPO_MAPPING }
func (r *TopoReport) String() string {
	return fmt.Sprintf("[Topografia] Missão %d - (%.4f, %.4f) h=%.2fm", r.MissionID, r.Latitude, r.Longitude, r.Height)
}

//
// ====== INSTALAÇÃO DE INSTRUMENTOS ======
//

// InstallReport representa o sucesso/insucesso da instalação (1 ou 0)
type InstallReport struct {
	TaskType     uint8
	MissionID    uint16
	Success      bool
	IsLastReport bool // indica se é o último report
}

func (r *InstallReport) Clone() Report {
	return &InstallReport{
		TaskType:     r.TaskType,
		MissionID:    r.MissionID,
		Success:      r.Success,
		IsLastReport: r.IsLastReport,
	}
}

func (r *InstallReport) GetMissionID() uint16 {
	return r.MissionID
}

func (r *InstallReport) IsLast() bool {
	return r.IsLastReport
}
func (r *InstallReport) ToBytes() []byte {
	buf := new(bytes.Buffer)
	binary.Write(buf, binary.BigEndian, r.TaskType)
	binary.Write(buf, binary.BigEndian, r.MissionID)
	var flag uint8
	if r.Success {
		flag = 1
	}
	binary.Write(buf, binary.BigEndian, flag)
	var last uint8
	if r.IsLastReport {
		last = 1
	}
	binary.Write(buf, binary.BigEndian, last)
	return buf.Bytes()
}

func (r *InstallReport) FromBytes(b []byte) error {
	if len(b) < 4 {
		return fmt.Errorf("report demasiado curto")
	}
	r.TaskType = b[0]
	r.MissionID = binary.BigEndian.Uint16(b[1:3])
	r.Success = b[3] == 1
	r.IsLastReport = b[len(b)-1] == 1
	return nil
}

func (r *InstallReport) GetTaskType() uint8 { return TASK_INSTALLATION }
func (r *InstallReport) String() string {
	if r.Success {
		return fmt.Sprintf("[Instalação] Missão %d - concluída com sucesso", r.MissionID)
	}
	return fmt.Sprintf("[Instalação] Missão %d - falhou", r.MissionID)
}
