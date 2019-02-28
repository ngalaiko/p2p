package messages

import "github.com/ngalayko/p2p/instance/messages/proto/chat"

// stream used to communicate with another peer.
type stream interface {
	Send(*chat.Message) error
	Recv() (*chat.Message, error)
}
