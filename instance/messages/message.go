package messages

import (
	"time"

	"github.com/ngalayko/p2p/instance/messages/proto/chat"
	"github.com/ngalayko/p2p/instance/peers"
)

// Type is a message type.
type Type string

// Known types.
var (
	TypeInvalid Type
	TypeText    Type = "text"
)

// Message is a single message.
type Message struct {
	ID        string
	From      *peers.Peer
	Timestamp time.Time
	Type      Type
	Text      string
}

func fromProto(from *peers.Peer, m *chat.Message) *Message {
	msg := &Message{
		ID:        m.ID,
		Timestamp: time.Unix(m.Timestamp.Seconds, 0),
		From:      from,
	}
	switch m.Payload.(type) {
	case *chat.Message_Text:
		msg.Type = TypeText
		msg.Text = m.Payload.(*chat.Message_Text).Text
	}
	return msg
}
