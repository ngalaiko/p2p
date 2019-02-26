package server

import (
	"fmt"
	"io"
	"time"

	"github.com/ngalayko/p2p/instance/logger"
	"github.com/ngalayko/p2p/instance/messages"
	"github.com/ngalayko/p2p/instance/messages/proto/chat"
)

// Server used to receive messages from other peers.
type Server struct {
	logger *logger.Logger

	out chan *messages.Message
}

// New is server constructor.
func New(
	log *logger.Logger,
) *Server {
	return &Server{
		logger: log.Prefix("grpc-server"),
		out:    make(chan *messages.Message),
	}
}

// Stream implements chat.MessageServer.
func (s *Server) Stream(srv chat.Chat_StreamServer) error {
	if err := s.listen(srv); err != nil {
		return err
	}
	return nil
}

// Received returns channel with incoming messages.
func (s *Server) Received() <-chan *messages.Message {
	return s.out
}

func (s *Server) listen(srv chat.Chat_StreamServer) error {
	for {
		msg, err := srv.Recv()
		switch err {
		case nil:
		case io.EOF:
			return nil
		default:
			return err
		}

		fmt.Printf("\nnikitag: %+v\n\n", msg)

		s.out <- parseMessage(msg)
	}
	return nil
}

func parseMessage(chatMsg *chat.Message) *messages.Message {
	msg := &messages.Message{
		Timestamp: time.Unix(chatMsg.Timestamp.Seconds, int64(chatMsg.Timestamp.Nanos)),
	}

	switch chatMsg.Payload.(type) {
	case *chat.Message_Text:
		msg.Type = messages.TypeText
		msg.Text = chatMsg.Payload.(*chat.Message_Text).Text
	default:
		msg.Type = messages.TypeInvalid
	}
	return msg
}
