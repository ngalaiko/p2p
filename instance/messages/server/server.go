package server

import (
	"fmt"

	"github.com/ngalayko/p2p/instance/logger"
	"github.com/ngalayko/p2p/instance/messages/proto/chat"
	"google.golang.org/grpc/metadata"
)

// Server used to receive a connection from another peer.
type Server struct {
	logger *logger.Logger

	newStreams chan *Stream
}

// New is server constructor.
func New(
	log *logger.Logger,
) *Server {
	return &Server{
		logger:     log.Prefix("grpc-server"),
		newStreams: make(chan *Stream),
	}
}

// Stream implements chat.MessageServer.
func (s *Server) Stream(srv chat.Chat_StreamServer) error {
	md, ok := metadata.FromIncomingContext(srv.Context())
	if !ok {
		return fmt.Errorf("metadata missing")
	}
	if len(md[chat.HeaderPeerID]) == 0 {
		return fmt.Errorf("peer id header missing")
	}

	peerID := md[chat.HeaderPeerID][0]

	s.logger.Info("%s connected", peerID)
	defer s.logger.Info("%s disconnected", peerID)

	s.newStreams <- newStream(srv, peerID)

	<-srv.Context().Done()

	return nil
}

// Streams returns channel with new streams.
func (s *Server) Streams() <-chan *Stream {
	return s.newStreams
}
