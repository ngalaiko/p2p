package ws

import "github.com/ngalayko/p2p/app/peers"

// messageType defines a message type.
type messageType string

// known MessageTypes
const (
	messageTypeInvalid    messageType = ""
	messageTypeInit       messageType = "init"
	messageTypePeersAdded messageType = "peer_added"
)

// message is a structure for client-server communication.
type message struct {
	Type messageType `json:"type"`
	Peer *peers.Peer `json:"peer"`
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
