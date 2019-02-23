package ws

import "github.com/ngalayko/p2p/app/peers"

// messageType defines a message type.
type messageType string

// known MessageTypes
const (
	messageTypeInvalid      messageType = ""
	messageTypeInit         messageType = "init"
	messageTypePeersUpdated messageType = "peers_update"
)

// message is a structure for client-server communication.
type message struct {
	Type messageType `json:"type"`
	Self *peers.Peer `json:"self"`
}

func newInitMessage(p *peers.Peer) *message {
	return &message{
		Type: messageTypeInit,
		Self: p,
	}
}

func newPeersUpdateMessage(p *peers.Peer) *message {
	return &message{
		Type: messageTypePeersUpdated,
		Self: p,
	}
}
