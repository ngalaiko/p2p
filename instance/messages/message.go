package messages

import (
	"encoding/hex"
	"fmt"
	"math/rand"
	"time"
)

const idLen = 32

// Type is a message type.
type Type string

// Supported types.
const (
	TypeInvalid Type = ""
	TypeText    Type = "text"
)

// Message is a single message.
type Message struct {
	ID        string    `json:"id"`
	Timestamp time.Time `json:"timestamp"`
	Type      Type      `json:"type"`
	Text      string    `json:"text,omitempty"`
}

// NewText returns new text message.
func NewText(text string) (*Message, error) {
	now := time.Now()

	r := rand.New(rand.NewSource(now.Unix()))
	idBytes := make([]byte, idLen)
	_, err := r.Read(idBytes)
	if err != nil {
		return nil, fmt.Errorf("error reading random bytes: %s", err)
	}

	return &Message{
		ID:        hex.EncodeToString(idBytes),
		Timestamp: now,
		Type:      TypeText,
		Text:      text,
	}, nil
}
