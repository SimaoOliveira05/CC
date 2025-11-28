package devices

import (
	"math/rand"
)

type Camera interface {
	ReadImageChunk() []byte
}

type MockCamera struct{}

func NewMockCamera() *MockCamera {
	return &MockCamera{}
}

func (c *MockCamera) ReadImageChunk() []byte {
	// Simula a leitura de um chunk de imagem retornando bytes aleatórios
	size := 1024
	chunk := make([]byte, size)
	_, err := rand.Read(chunk)
	if err != nil {
		// Em caso de erro, retorna um slice vazio
		return []byte{}
	}
	// Simula uma chance de falha na leitura
	if rand.Float32() < 0.1 {
		return []byte{}
	}
	return chunk
}
