package messages

// Type is a message type.
type Type uint

// Supported types.
const (
	TypeUndefined = iota
	TypeText
)

// Message is a single message.
type Message struct {
	Type    Type   `json:"type"`
	Payload []byte `json:"payload"`
}

// Text is a text message.
type Text struct {
	Data []byte
}
