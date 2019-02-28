package ws

import (
	"github.com/ngalayko/p2p/instance/messages"
	"github.com/ngalayko/p2p/instance/peers"
)

// messageType defines a message type.
type messageType string

// known MessageTypes
const (
	messageTypeInvalid      messageType = ""
	messageTypeInit         messageType = "init"
	messageTypePeersAdded   messageType = "peer_added"
	messageTypeTextSent     messageType = "text_sent"
	messageTypeTextReceived messageType = "text_received"
)

// message is a structure for client-server communication.
type message struct {
	Type    messageType       `json:"type"`
	Peer    *peers.Peer       `json:"peer,omitempty"`
	Message *messages.Message `json:"message,omitempty"`
}

func newInitMessage(p *peers.Peer) *message {
	return &message{
		Type: messageTypeInit,
		Peer: p,
	}
}

func newPeerAddedMessage(p *peers.Peer) *message {
	return &message{
		Type: messageTypePeersAdded,
		Peer: p,
	}
}

func newTextMessageSent(msg *messages.Message) *message {
	return &message{
		Type:    messageTypeTextSent,
		Message: msg,
	}
}

func newTextMessageReceived(msg *messages.Message) *message {
	return &message{
		Type:    messageTypeTextReceived,
		Message: msg,
	}
}
