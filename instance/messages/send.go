package messages

import (
	"context"
	"encoding/hex"
	"fmt"
	"math/rand"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	"github.com/golang/protobuf/ptypes/timestamp"
	"github.com/ngalayko/p2p/instance/messages/proto/chat"
)

// SendText sends a text message.
func (h *Handler) SendText(ctx context.Context, text string, toID string) error {
	msg, err := makeText(h.r, text)
	if err != nil {
		return fmt.Errorf("error making message: %s", err)
	}

	return h.sendMessage(ctx, toID, msg)
}

func (h *Handler) sendMessage(ctx context.Context, toID string, msg *chat.Message) error {
	to, err := h.getPeer(toID)
	if err != nil {
		return fmt.Errorf("error getting peer: %s", err)
	}

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
		h.logger.Info("sent a message to %s", toID)
		return nil
	case codes.Unavailable, codes.Canceled, codes.DeadlineExceeded:
		h.logger.Error("client (%s) terminated connection", toID)
		return sendErr
	default:
		h.logger.Error("failed to send to client (%s): %s", toID, err)
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
		Id: hex.EncodeToString(idBytes),
		Timestamp: &timestamp.Timestamp{
			Seconds: time.Now().Unix(),
		},
		Payload: &chat.Message_Text{
			Text: text,
		},
	}, nil
}
