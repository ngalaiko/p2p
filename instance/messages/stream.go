package messages

import "github.com/ngalayko/p2p/instance/messages/proto/chat"

// Stream used to communicate with another peer.
type Stream interface {
	Send(*chat.Message) error
	Recv() (*chat.Message, error)
}
