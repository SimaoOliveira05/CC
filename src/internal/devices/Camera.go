package devices

import (
	"math/rand"
	"os"
)

type Camera interface {
	ReadImageChunk() []byte
}

type MockCamera struct{}

func NewMockCamera() *MockCamera {
	return &MockCamera{}
}

func (c *MockCamera) ReadImageChunk() []byte {
	path := "image.jpg"
	data, err := os.ReadFile(path)
	if err == nil && len(data) > 0 {
		if len(data) > 256 {
			return data[:256]
		}
		return data
	}
	fake := make([]byte, 256)
	for i := range fake {
		fake[i] = byte(rand.Intn(256))
	}
	return fake
}
