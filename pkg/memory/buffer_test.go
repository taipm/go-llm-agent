package memory

import (
	"testing"

	"github.com/taipm/go-llm-agent/pkg/types"
)

func TestBufferMemory_Add(t *testing.T) {
	mem := NewBuffer(3)

	msg1 := types.Message{Role: types.RoleUser, Content: "Hello"}
	msg2 := types.Message{Role: types.RoleAssistant, Content: "Hi"}
	msg3 := types.Message{Role: types.RoleUser, Content: "How are you?"}
	msg4 := types.Message{Role: types.RoleAssistant, Content: "I'm good"}

	mem.Add(msg1)
	mem.Add(msg2)
	mem.Add(msg3)

	if mem.Size() != 3 {
		t.Errorf("Expected size 3, got %d", mem.Size())
	}

	// Add 4th message, should remove oldest
	mem.Add(msg4)
	if mem.Size() != 3 {
		t.Errorf("Expected size 3 after overflow, got %d", mem.Size())
	}

	// Check oldest message was removed
	history, _ := mem.GetHistory(0)
	if history[0].Content != "Hi" {
		t.Errorf("Expected oldest message to be removed, got %s", history[0].Content)
	}
}

func TestBufferMemory_GetHistory(t *testing.T) {
	mem := NewBuffer(10)

	for i := 0; i < 5; i++ {
		mem.Add(types.Message{
			Role:    types.RoleUser,
			Content: string(rune('A' + i)),
		})
	}

	// Get last 3 messages
	history, err := mem.GetHistory(3)
	if err != nil {
		t.Fatalf("GetHistory failed: %v", err)
	}

	if len(history) != 3 {
		t.Errorf("Expected 3 messages, got %d", len(history))
	}

	if history[0].Content != "C" {
		t.Errorf("Expected first message to be 'C', got %s", history[0].Content)
	}
}

func TestBufferMemory_Clear(t *testing.T) {
	mem := NewBuffer(10)

	mem.Add(types.Message{Role: types.RoleUser, Content: "Test"})

	if mem.Size() != 1 {
		t.Errorf("Expected size 1, got %d", mem.Size())
	}

	mem.Clear()

	if mem.Size() != 0 {
		t.Errorf("Expected size 0 after clear, got %d", mem.Size())
	}
}

func TestBufferMemory_SetMaxSize(t *testing.T) {
	mem := NewBuffer(10)

	for i := 0; i < 5; i++ {
		mem.Add(types.Message{Role: types.RoleUser, Content: "Test"})
	}

	// Reduce max size
	err := mem.SetMaxSize(3)
	if err != nil {
		t.Fatalf("SetMaxSize failed: %v", err)
	}

	if mem.Size() != 3 {
		t.Errorf("Expected size 3 after reducing max size, got %d", mem.Size())
	}

	// Try invalid size
	err = mem.SetMaxSize(0)
	if err == nil {
		t.Error("Expected error for invalid max size")
	}
}
