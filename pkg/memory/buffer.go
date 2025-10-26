package memory

import (
	"fmt"
	"sync"

	"github.com/taipm/go-llm-agent/pkg/types"
)

// BufferMemory implements in-memory conversation storage with a fixed buffer size
type BufferMemory struct {
	mu       sync.RWMutex
	messages []types.Message
	maxSize  int
}

// NewBuffer creates a new buffer memory with the specified max size
func NewBuffer(maxSize int) *BufferMemory {
	if maxSize <= 0 {
		maxSize = 100 // Default size
	}
	return &BufferMemory{
		messages: make([]types.Message, 0, maxSize),
		maxSize:  maxSize,
	}
}

// Add adds a message to memory
func (m *BufferMemory) Add(message types.Message) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Add message
	m.messages = append(m.messages, message)

	// Trim if exceeds max size (FIFO - keep most recent)
	if len(m.messages) > m.maxSize {
		// Remove oldest messages
		removeCount := len(m.messages) - m.maxSize
		m.messages = m.messages[removeCount:]
	}

	return nil
}

// GetHistory returns recent messages up to the limit
// If limit <= 0, returns all messages
func (m *BufferMemory) GetHistory(limit int) ([]types.Message, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if limit <= 0 || limit >= len(m.messages) {
		// Return all messages (copy to avoid external modification)
		result := make([]types.Message, len(m.messages))
		copy(result, m.messages)
		return result, nil
	}

	// Return last 'limit' messages
	start := len(m.messages) - limit
	result := make([]types.Message, limit)
	copy(result, m.messages[start:])
	return result, nil
}

// Clear clears all messages from memory
func (m *BufferMemory) Clear() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.messages = make([]types.Message, 0, m.maxSize)
	return nil
}

// Size returns the number of messages in memory
func (m *BufferMemory) Size() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.messages)
}

// GetAll returns all messages (useful for debugging)
func (m *BufferMemory) GetAll() []types.Message {
	messages, _ := m.GetHistory(0)
	return messages
}

// SetMaxSize updates the maximum buffer size
func (m *BufferMemory) SetMaxSize(size int) error {
	if size <= 0 {
		return fmt.Errorf("max size must be positive")
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	m.maxSize = size

	// Trim if current size exceeds new max
	if len(m.messages) > m.maxSize {
		removeCount := len(m.messages) - m.maxSize
		m.messages = m.messages[removeCount:]
	}

	return nil
}
