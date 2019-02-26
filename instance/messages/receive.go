package messages

import (
	"fmt"
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

		fmt.Printf("\nnikitag: %+v\n\n", msg)
	}
}
