package messages

import (
	"io"
)

func (h *Handler) listenStream(s Stream, peerID string) {
	h.logger.Info("listening for messages from %s", peerID)
	defer h.logger.Info("stop listening for messages from %s", peerID)

	for {
		msg, err := s.Recv()
		switch err {
		case nil, io.EOF:
		default:
			continue
		}

		if msg == nil {
			continue
		}

		h.logger.Info("new message from %s", peerID)

		peer, err := h.getPeer(peerID)
		if err != nil {
			h.logger.Error("message from unknown peer: %s", err)
			continue
		}

		h.received <- fromProto(peer, h.self, msg)
	}
}

// Received returns a channel with new messages.
func (h *Handler) Received() <-chan *Message {
	return h.received
}
