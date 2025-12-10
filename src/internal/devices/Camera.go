package devices

import (
	"io"
	"math/rand"
	"os"
	"src/config"
)

// Camera interface
type Camera interface {
	ReadImageChunk() []byte
	LoadImage(path string) error
	GetTotalChunks() int
	GetChunk(index int) []byte
}

// MockCamera simulate a device camera for testing purposes
type MockCamera struct {
	imageData   []byte // Full image data
	chunkSize   int    // Size of each chunk
	totalChunks int    // Total number of chunks
}

// NewMockCamera creates a new MockCamera
func NewMockCamera() *MockCamera {
	return &MockCamera{
		chunkSize: config.CAMERA_CHUNK_SIZE,
	}
}

// LoadImage loads an image from disk
func (c *MockCamera) LoadImage(path string) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		return err
	}

	c.imageData = data
	c.totalChunks = (len(data) + c.chunkSize - 1) / c.chunkSize
	return nil
}

// GetTotalChunks returns the total number of chunks
func (c *MockCamera) GetTotalChunks() int {
	return c.totalChunks
}

// GetChunk returns a specific chunk by index (0-based)
func (c *MockCamera) GetChunk(index int) []byte {
	if index >= c.totalChunks || c.imageData == nil {
		return []byte{}
	}

	start := index * c.chunkSize
	end := start + c.chunkSize
	if end > len(c.imageData) {
		end = len(c.imageData)
	}

	chunk := make([]byte, end-start)
	copy(chunk, c.imageData[start:end])
	return chunk
}

// ReadImageChunk simulates reading a chunk of image data (for backward compatibility)
func (c *MockCamera) ReadImageChunk() []byte {
	// If image is loaded, return first chunk
	if len(c.imageData) > 0 {
		return c.GetChunk(0)
	}

	// Simulate reading a chunk of image data by returning random bytes
	chunk := make([]byte, config.CAMERA_CHUNK_SIZE)
	_, err := rand.Read(chunk)
	if err != nil {
		// In case of error, return an empty slice
		return []byte{}
	}
	// Simulate a chance of read failure
	if rand.Float32() < config.CAMERA_FAIL_CHANCE {
		return []byte{}
	}
	return chunk
}
