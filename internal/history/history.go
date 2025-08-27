package history

import (
	"fmt"
	"sync"

	"chat/internal/database"
)

// History manages message history
type History struct {
	messages []string
	capacity int
	mu       sync.Mutex
	db       *database.DB
}

// New creates a new history instance with database integration
func New(capacity int, db *database.DB) *History {
	h := &History{
		messages: make([]string, 0, capacity),
		capacity: capacity,
		db:       db,
	}
	h.loadFromDB()
	return h
}

// Add adds a message to history and database
func (h *History) Add(msg, from, to, timestamp string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if len(h.messages) >= h.capacity {
		h.messages = h.messages[1:]
	}
	h.messages = append(h.messages, msg)
	if err := h.db.SaveMessage(from, to, msg, timestamp); err != nil {
		// Log error if needed
		fmt.Printf("Failed to save message to DB: %v\n", err)
	}
}

// GetAll returns all history messages
func (h *History) GetAll() []string {
	h.mu.Lock()
	defer h.mu.Unlock()
	result := make([]string, len(h.messages))
	copy(result, h.messages)
	return result
}

// loadFromDB loads recent messages from database
func (h *History) loadFromDB() {
	messages, err := h.db.LoadRecentMessages(h.capacity)
	if err != nil {
		// Log error if needed
		fmt.Printf("Failed to load history from DB: %v\n", err)
		return
	}
	h.mu.Lock()
	h.messages = messages
	h.mu.Unlock()
}