package structs

import (
	"encoding/json"
	"time"

	"hkchat/tables"

	"github.com/google/uuid"
)

type ChatSignal int

const (
	Join ChatSignal = iota
	Leave
	SendMessage
	CreateChat
	DeleteChat
)

type ControlMessage struct {
	Signal        ChatSignal
	UserID        int
	Username      string
	Message       string
	OutputChannel chan Message
}

type RabbitEnvelope struct {
	EventID   string         `json:"event_id"`
	ChatID    uint           `json:"chat_id"`
	Message   tables.Message `json:"message"`
	CreatedAt time.Time      `json:"created_at"`
}

func NewEnvelope(chatID uint, msg tables.Message) *RabbitEnvelope {
	return &RabbitEnvelope{
		EventID:   uuid.New().String(),
		ChatID:    chatID,
		Message:   msg,
		CreatedAt: time.Now(),
	}
}

// ToJSON сериализует в JSON
func (e *RabbitEnvelope) ToJSON() ([]byte, error) {
	return json.Marshal(e)
}

// FromJSON десериализует
func (e *RabbitEnvelope) FromJSON(data []byte) error {
	return json.Unmarshal(data, e)
}
