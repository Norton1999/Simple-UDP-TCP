package message

import "fmt"

// MessageType defines types of messages
type MessageType int

const (
	TypeSystem MessageType = iota
	TypeUser
	TypePrivate
)

// Message represents a chat message
type Message struct {
	Type   MessageType
	From   string
	Target string // For private messages
	Content string
}

// NewSystemMessage creates a system message
func NewSystemMessage(content string) Message {
	return Message{
		Type:    TypeSystem,
		Content: fmt.Sprintf("[SYSTEM] %s", content),
	}
}

// NewUserMessage creates a user message
func NewUserMessage(from, content string) Message {
	return Message{
		Type:    TypeUser,
		From:    from,
		Content: fmt.Sprintf("[%s] %s", from, content),
	}
}

// NewPrivateMessage creates a private message
func NewPrivateMessage(from, target, content string) Message {
	return Message{
		Type:    TypePrivate,
		From:    from,
		Target:  target,
		Content: fmt.Sprintf("[PRIVATE from %s] %s", from, content),
	}
}

// String returns the string representation of the message
func (m Message) String() string {
	return m.Content
}