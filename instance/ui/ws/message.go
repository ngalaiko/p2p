package ws

import "github.com/ngalayko/p2p/instance/peers"

// messageType defines a message type.
type messageType string

// known MessageTypes
const (
	messageTypeInvalid    messageType = ""
	messageTypeInit       messageType = "init"
	messageTypePeersAdded messageType = "peer_added"
	messageTypeText       messageType = "text"
)

// message is a structure for client-server communication.
type message struct {
	Peer *peers.Peer `json:"peer"`
	Type messageType `json:"type"`
	Text string      `json:"text,omitempty"`
}

func newInitMessage(p *peers.Peer) *message {
	return &message{
		Type: messageTypeInit,
		Peer: p,
	}
}

func newPeersAddedMessage(p *peers.Peer) *message {
	return &message{
		Type: messageTypePeersAdded,
		Peer: p,
	}
}
