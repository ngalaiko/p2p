package messages

import (
	"context"
	"encoding/hex"
	"fmt"
	"math/rand"
	"time"

	"github.com/golang/protobuf/ptypes/timestamp"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	"github.com/ngalayko/p2p/instance/messages/proto/chat"
	"github.com/ngalayko/p2p/instance/peers"
)

// Sent returns a channel with sent messages.
func (h *Handler) Sent() <-chan *Message {
	return h.sent
}

// SendText sends a text message.
func (h *Handler) SendText(ctx context.Context, text string, toID string) error {
	msg, err := makeText(h.r, text)
	if err != nil {
		return fmt.Errorf("error making message: %s", err)
	}

	to, err := h.getPeer(toID)
	if err != nil {
		return fmt.Errorf("error getting peer: %s", err)
	}

	if err := h.sendMessage(ctx, to, msg); err != nil {
		return fmt.Errorf("error sending a text: %s", err)
	}

	h.sent <- fromProto(h.self, to, msg)

	return nil
}

func (h *Handler) sendMessage(ctx context.Context, to *peers.Peer, msg *chat.Message) error {
	md := metadata.New(map[string]string{
		chat.HeaderPeerID: h.self.ID,
	})
	ctx = metadata.NewOutgoingContext(ctx, md)

	stream, err := h.getStream(ctx, to)
	if err != nil {
		return err
	}

	sendErr := stream.Send(msg)
	s, _ := status.FromError(sendErr)
	switch s.Code() {
	case codes.OK:
		h.logger.Info("sent a message to %s", to.ID)
		return nil
	case codes.Unavailable, codes.Canceled, codes.DeadlineExceeded:
		h.logger.Error("client (%s) terminated connection", to.ID)
		return sendErr
	default:
		h.logger.Error("failed to send to client (%s): %s", to.ID, err)
		return sendErr
	}
}

func makeText(r *rand.Rand, text string) (*chat.Message, error) {
	idBytes := make([]byte, idLen)
	_, err := r.Read(idBytes)
	if err != nil {
		return nil, fmt.Errorf("error reading random bytes: %s", err)
	}
	return &chat.Message{
		ID: hex.EncodeToString(idBytes),
		Timestamp: &timestamp.Timestamp{
			Seconds: time.Now().Unix(),
		},
		Payload: &chat.Message_Text{
			Text: text,
		},
	}, nil
}
