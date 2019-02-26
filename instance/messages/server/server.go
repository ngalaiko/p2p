package server

import (
	"io"

	"github.com/ngalayko/p2p/instance/logger"
	"github.com/ngalayko/p2p/instance/messages/proto/chat"
)

// Server used to receive messages from other peers.
type Server struct {
	logger *logger.Logger

	out chan *chat.Message
}

// New is server constructor.
func New(
	log *logger.Logger,
) *Server {
	return &Server{
		logger: log.Prefix("grpc-server"),
		out:    make(chan *chat.Message),
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
func (s *Server) Received() <-chan *chat.Message {
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

		s.logger.Info("received a message")

		s.out <- msg
	}
	return nil
}
