package messages

import (
	"context"
	"crypto/x509"
	"fmt"
	"math/rand"
	"net"
	"sync"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/metadata"
	grpc_resolver "google.golang.org/grpc/resolver"

	"github.com/ngalayko/p2p/instance/logger"
	"github.com/ngalayko/p2p/instance/messages/client"
	"github.com/ngalayko/p2p/instance/messages/client/resolver"
	"github.com/ngalayko/p2p/instance/messages/proto/chat"
	"github.com/ngalayko/p2p/instance/messages/server"
	"github.com/ngalayko/p2p/instance/peers"
)

const idLen = 32

// Handler runs message server and holds connection to known peers.
type Handler struct {
	logger *logger.Logger
	self   *peers.Peer
	r      *rand.Rand

	port string

	srv            *grpc.Server
	messagesServer *server.Server

	streamsGuard *sync.RWMutex
	streams      map[string]Stream

	clientsResolver *resolver.Builder
}

// NewHandler returns new messages handler.
func NewHandler(
	r *rand.Rand,
	log *logger.Logger,
	self *peers.Peer,
	port string,
) *Handler {
	grpcServer := grpc.NewServer(
		grpc.Creds(credentials.NewServerTLSFromCert(self.Certificate)),
	)

	messagesServer := server.New(log)
	chat.RegisterChatServer(grpcServer, messagesServer)

	customResolver := resolver.New()
	grpc_resolver.Register(customResolver)

	return &Handler{
		r:               r,
		logger:          log.Prefix("messages"),
		self:            self,
		port:            port,
		srv:             grpcServer,
		messagesServer:  messagesServer,
		streamsGuard:    &sync.RWMutex{},
		streams:         map[string]Stream{},
		clientsResolver: customResolver,
	}
}

// Start starts the server.
func (h *Handler) Start(ctx context.Context) error {
	addr := net.JoinHostPort("0.0.0.0", h.port)
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	h.logger.Info("starting message server on %s", addr)

	go h.watchStreamsFromServer()

	return h.srv.Serve(lis)
}

func (h *Handler) watchStreamsFromServer() {
	for stream := range h.messagesServer.Streams() {
		h.streamsGuard.Lock()
		h.streams[stream.PeerID] = stream.Chat_StreamServer
		h.streamsGuard.Unlock()

		go h.listenStream(stream.Chat_StreamServer, stream.PeerID)
	}
}

// returns an existed stream or opens a new one.
func (h *Handler) getStream(ctx context.Context, peer *peers.Peer) (Stream, error) {
	h.streamsGuard.RLock()
	stream, found := h.streams[peer.ID]
	h.streamsGuard.RUnlock()
	if found {
		return stream, nil
	}

	h.clientsResolver.Add(peer)

	cp := x509.NewCertPool()
	if !cp.AppendCertsFromPEM(peer.PublicCrt) {
		return nil, fmt.Errorf("failed to append certificates")
	}
	creds := credentials.NewClientTLSFromCert(cp, "")

	md := metadata.New(map[string]string{chat.HeaderPeerID: h.self.ID})
	ctx = metadata.NewOutgoingContext(ctx, md)

	grpcClient, err := client.Connect(ctx, h.r, h.logger, creds, peer)
	if err != nil {
		return nil, fmt.Errorf("can't create client: %s", err)
	}

	h.streamsGuard.Lock()
	h.streams[peer.ID] = grpcClient
	h.streamsGuard.Unlock()

	go h.listenStream(grpcClient, peer.ID)

	h.logger.Info("connected to %s", peer.ID)

	return grpcClient, nil
}

func (h *Handler) getPeer(peerID string) (*peers.Peer, error) {
	switch peerID {
	case h.self.ID:
		return h.self, nil
	default:
		if peer := h.self.KnownPeers.Get(peerID); peer != nil {
			return peer, nil
		}
		return nil, fmt.Errorf("unknown peer: %s", peerID)
	}
}
