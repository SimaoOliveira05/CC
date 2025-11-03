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
}

//
// ====== CAPTURA DE IMAGEM ======
//

// ImageReport representa um relatório parcial de imagem (chunk)
type ImageReport struct {
	TaskType  uint8
	MissionID uint16
	ChunkID   uint16
	Data      []byte // bytes da imagem (parcial)
}

func (r *ImageReport) ToBytes() []byte {
	buf := new(bytes.Buffer)
	binary.Write(buf, binary.BigEndian, r.TaskType)
	binary.Write(buf, binary.BigEndian, r.MissionID)
	binary.Write(buf, binary.BigEndian, r.ChunkID)
	buf.Write(r.Data)
	return buf.Bytes()
}

func (r *ImageReport) FromBytes(b []byte) error {
	if len(b) < 5 {
		return fmt.Errorf("report demasiado curto")
	}
	r.TaskType = b[0]
	r.MissionID = binary.BigEndian.Uint16(b[1:3])
	r.ChunkID = binary.BigEndian.Uint16(b[3:5])
	r.Data = b[5:]
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
	TaskType   uint8
	MissionID  uint16
	NumSamples uint8
	Components []Component // lista de (nome, %)
}

func (r *SampleReport) ToBytes() []byte {
	buf := new(bytes.Buffer)
	binary.Write(buf, binary.BigEndian, r.TaskType)
	binary.Write(buf, binary.BigEndian, r.MissionID)
	binary.Write(buf, binary.BigEndian, r.NumSamples)
	for _, c := range r.Components {
		// Serializar nome com tamanho
		nameLen := uint8(len(c.Name))
		binary.Write(buf, binary.BigEndian, nameLen)
		buf.WriteString(c.Name)
		// Serializar percentagem
		binary.Write(buf, binary.BigEndian, c.Percentage)
	}
	return buf.Bytes()
}

func (r *SampleReport) FromBytes(b []byte) error {
	if len(b) < 4 {
		return fmt.Errorf("report demasiado curto")
	}
	r.TaskType = b[0]
	r.MissionID = binary.BigEndian.Uint16(b[1:3])
	r.NumSamples = b[3]
	count := int(r.NumSamples)
	r.Components = make([]Component, count)

	buf := bytes.NewReader(b[4:])
	for i := 0; i < count; i++ {
		// Ler tamanho do nome
		var nameLen uint8
		err := binary.Read(buf, binary.BigEndian, &nameLen)
		if err != nil {
			return fmt.Errorf("erro ao ler tamanho do nome: %w", err)
		}

		// Ler nome
		nameBytes := make([]byte, nameLen)
		_, err = buf.Read(nameBytes)
		if err != nil {
			return fmt.Errorf("erro ao ler nome: %w", err)
		}
		r.Components[i].Name = string(nameBytes)

		// Ler percentagem
		err = binary.Read(buf, binary.BigEndian, &r.Components[i].Percentage)
		if err != nil {
			return fmt.Errorf("erro ao ler percentagem: %w", err)
		}
	}
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
	TaskType  uint8
	MissionID uint16
	Temp      float32
	Oxygen    float32
	Pressure  float32
	Humidity  float32
	WindSpeed float32
	Radiation float32
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
	return buf.Bytes()
}

func (r *EnvReport) FromBytes(b []byte) error {
	if len(b) < 27 {
		return fmt.Errorf("report demasiado curto")
	}
	r.TaskType = b[0]
	r.MissionID = binary.BigEndian.Uint16(b[1:3])
	buf := bytes.NewReader(b[3:])
	return binary.Read(buf, binary.BigEndian, &r.Temp)
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
	TaskType   uint8
	MissionID  uint16
	ProblemID  uint8
	Repairable bool
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
	return buf.Bytes()
}

func (r *RepairReport) FromBytes(b []byte) error {
	if len(b) < 5 {
		return fmt.Errorf("report demasiado curto")
	}
	r.TaskType = b[0]
	r.MissionID = binary.BigEndian.Uint16(b[1:3])
	r.ProblemID = b[3]
	r.Repairable = b[4] == 1
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
	TaskType  uint8
	MissionID uint16
	Latitude  float32
	Longitude float32
	Height    float32
}

func (r *TopoReport) ToBytes() []byte {
	buf := new(bytes.Buffer)
	binary.Write(buf, binary.BigEndian, r.TaskType)
	binary.Write(buf, binary.BigEndian, r.MissionID)
	binary.Write(buf, binary.BigEndian, r.Latitude)
	binary.Write(buf, binary.BigEndian, r.Longitude)
	binary.Write(buf, binary.BigEndian, r.Height)
	return buf.Bytes()
}

func (r *TopoReport) FromBytes(b []byte) error {
	if len(b) < 15 {
		return fmt.Errorf("report demasiado curto")
	}
	r.TaskType = b[0]
	r.MissionID = binary.BigEndian.Uint16(b[1:3])
	buf := bytes.NewReader(b[3:])
	return binary.Read(buf, binary.BigEndian, &r.Latitude)
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
	TaskType  uint8
	MissionID uint16
	Success   bool
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
	return buf.Bytes()
}

func (r *InstallReport) FromBytes(b []byte) error {
	if len(b) < 4 {
		return fmt.Errorf("report demasiado curto")
	}
	r.TaskType = b[0]
	r.MissionID = binary.BigEndian.Uint16(b[1:3])
	r.Success = b[3] == 1
	return nil
}

func (r *InstallReport) GetTaskType() uint8 { return TASK_INSTALLATION }
func (r *InstallReport) String() string {
	if r.Success {
		return fmt.Sprintf("[Instalação] Missão %d - concluída com sucesso", r.MissionID)
	}
	return fmt.Sprintf("[Instalação] Missão %d - falhou", r.MissionID)
}
