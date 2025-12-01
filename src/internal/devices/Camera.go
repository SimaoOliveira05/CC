package devices

import (
	"math/rand"
)

// Camera interface
type Camera interface {
	ReadImageChunk() []byte
}

// MockCamera simulate a device camera for testing purposes
type MockCamera struct{}

// NewMockCamera creates a new MockCamera
func NewMockCamera() *MockCamera {
	return &MockCamera{}
}

// ReadImageChunk simulates reading a chunk of image data
func (c *MockCamera) ReadImageChunk() []byte {
	// Simulate reading a chunk of image data by returning random bytes
	size := 1024
	chunk := make([]byte, size)
	_, err := rand.Read(chunk)
	if err != nil {
		// In case of error, return an empty slice
		return []byte{}
	}
	// Simulate a chance of read failure
	if rand.Float32() < 0.1 {
		return []byte{}
	}
	return chunk
}
