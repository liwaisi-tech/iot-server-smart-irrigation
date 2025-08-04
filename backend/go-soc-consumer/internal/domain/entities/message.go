package entities

import (
	"fmt"
	"time"
)

type Message struct {
	ID        string
	Content   string
	CreatedAt time.Time
}

func NewMessage(content string) (*Message, error) {
	if content == "" {
		return nil, fmt.Errorf("message content cannot be empty")
	}
	
	return &Message{
		ID:        fmt.Sprintf("msg_%d", time.Now().UnixNano()),
		Content:   content,
		CreatedAt: time.Now(),
	}, nil
}

func (m *Message) IsValid() error {
	if m.ID == "" {
		return fmt.Errorf("message ID is required")
	}
	if m.Content == "" {
		return fmt.Errorf("message content is required")
	}
	return nil
}