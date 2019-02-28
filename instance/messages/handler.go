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
	"github.com/ngalayko/p2p/instance/messages/proto/greeter"
	"github.com/ngalayko/p2p/instance/messages/server"
	"github.com/ngalayko/p2p/instance/peers"
)

const idLen = 32

// Handler runs message server and holds connection to known peers.
type Handler struct {
	logger *logger.Logger
	self   *peers.Peer
	r      *rand.Rand

	secureServer   *grpc.Server
	secureResolver *resolver.Builder

	insecureServer   *grpc.Server
	insecureResolver *resolver.Builder

	messagesServer *server.Server

	streamsGuard *sync.RWMutex
	streams      map[string]Stream

	received chan *Message
}

// NewHandler returns new messages handler.
func NewHandler(
	r *rand.Rand,
	log *logger.Logger,
	self *peers.Peer,
) *Handler {
	messagesServer := server.New(log, self)

	secureGRPCserver := grpc.NewServer(
		grpc.Creds(credentials.NewServerTLSFromCert(self.Certificate)),
	)
	chat.RegisterChatServer(secureGRPCserver, messagesServer)

	insecureGRPCServer := grpc.NewServer()
	greeter.RegisterGreeterServer(insecureGRPCServer, messagesServer)

	secureResolver := resolver.New(true)
	grpc_resolver.Register(secureResolver)

	insecureResolver := resolver.New(false)
	grpc_resolver.Register(insecureResolver)

	return &Handler{
		r:      r,
		logger: log.Prefix("messages"),
		self:   self,

		secureServer:   secureGRPCserver,
		secureResolver: secureResolver,

		insecureServer:   insecureGRPCServer,
		insecureResolver: insecureResolver,

		messagesServer: messagesServer,

		streamsGuard: &sync.RWMutex{},
		streams:      map[string]Stream{},

		received: make(chan *Message),
	}
}

// Start starts the server.
func (h *Handler) Start(ctx context.Context) error {
	lis, err := net.Listen("tcp", net.JoinHostPort("0.0.0.0", h.self.Port))
	if err != nil {
		return err
	}

	insecureL, err := net.Listen("tcp", net.JoinHostPort("0.0.0.0", h.self.InsecurePort))
	if err != nil {
		return err
	}

	secureL := lis
	go func() {
		h.logger.Info("starting message server on %s", secureL.Addr())
		defer h.logger.Info("stopping message server on %s", secureL.Addr())
		if err := h.secureServer.Serve(secureL); err != nil {
			h.logger.Error("failed to start message server: %s", err)
		}
	}()

	go func() {
		h.logger.Info("starting greeting server on %s", insecureL.Addr())
		defer h.logger.Info("stopping greeting server on %s", insecureL.Addr())
		if err := h.insecureServer.Serve(insecureL); err != nil {
			h.logger.Error("failed to start greeting server: %s", err)
		}
	}()

	go h.watchStreamsFromServer()

	<-ctx.Done()

	h.secureServer.GracefulStop()
	h.insecureServer.GracefulStop()

	return nil
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

	h.insecureResolver.Add(peer)
	h.secureResolver.Add(peer)

	knownPeer, err := h.greet(ctx, peer)
	if err != nil {
		return nil, fmt.Errorf("error during greeting: %s", err)
	}

	return h.openStream(ctx, knownPeer)
}

func (h *Handler) greet(ctx context.Context, peer *peers.Peer) (*peers.Peer, error) {
	grpcClient, err := client.InsecureConnect(ctx, h.logger, peer)
	if err != nil {
		return nil, fmt.Errorf("can't connect: %s", err)
	}

	selfProto := &greeter.Peer{}
	if err := selfProto.UnmarshalPeer(h.self); err != nil {
		return nil, fmt.Errorf("can't unmarshal self")
	}

	h.logger.Info("greeting %s", peer.ID)

	grpcPeer, err := grpcClient.Greet(ctx, selfProto)
	if err != nil {
		return nil, fmt.Errorf("greeting error: %s", err)
	}

	return grpcPeer.MarshalPeer()
}

func (h *Handler) openStream(ctx context.Context, peer *peers.Peer) (Stream, error) {
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
		if peer, ok := h.self.KnownPeers.Map()[peerID]; ok {
			return peer, nil
		}
		return nil, fmt.Errorf("unknown peer: %s", peerID)
	}
}
