package messages

import "github.com/ngalayko/p2p/app/peers"

// Type is a message type.
type Type string

// Supported types.
const (
	TypeUndefined Type = ""
	TypeText      Type = "text"
)

// Message is a single message.
type Message struct {
	To   []*peers.Peer `json:"to"`
	From *peers.Peer   `json:"from"`
	Type Type          `json:"type"`
	Text string        `json:"text,omitempty"`
}
