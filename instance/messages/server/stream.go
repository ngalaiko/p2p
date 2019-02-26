package server

import (
	"github.com/ngalayko/p2p/instance/messages/proto/chat"
)

// Stream implements messages.Stream
type Stream struct {
	chat.Chat_StreamServer

	PeerID string
}

func newStream(
	srv chat.Chat_StreamServer,
	peerID string,
) *Stream {
	s := &Stream{
		Chat_StreamServer: srv,
		PeerID:            peerID,
	}

	return s
}
