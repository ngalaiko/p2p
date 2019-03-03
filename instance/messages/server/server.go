package server

import (
	"context"
	"fmt"

	"google.golang.org/grpc/metadata"

	"github.com/ngalayko/p2p/instance/messages/proto/chat"
	"github.com/ngalayko/p2p/instance/messages/proto/greeter"
	"github.com/ngalayko/p2p/instance/peers"
	"github.com/ngalayko/p2p/logger"
)

// Server used to receive a connection from another peer.
type Server struct {
	logger *logger.Logger

	self      *peers.Peer
	selfProto *greeter.Peer

	newStreams chan *Stream
}

// New is server constructor.
func New(
	log *logger.Logger,
	self *peers.Peer,
) *Server {
	selfProto := &greeter.Peer{}
	_ = selfProto.UnmarshalPeer(self)
	return &Server{
		logger:     log.Prefix("grpc-server"),
		newStreams: make(chan *Stream),
		selfProto:  selfProto,
		self:       self,
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

// Greet implements greeter.Greeter.
func (s *Server) Greet(ctx context.Context, peer *greeter.Peer) (*greeter.Peer, error) {
	p, err := peer.MarshalPeer()
	if err != nil {
		return nil, fmt.Errorf("invalid peer: %s", err)
	}
	s.self.KnownPeers.Add(p)

	s.logger.Info("greeted %s", peer.ID)

	return s.selfProto, nil
}

// Streams returns channel with new streams.
func (s *Server) Streams() <-chan *Stream {
	return s.newStreams
}
