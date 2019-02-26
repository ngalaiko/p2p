package handler

import (
	"context"
	"crypto/x509"
	"fmt"
	"net"
	"sync"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	grpc_resolver "google.golang.org/grpc/resolver"

	"github.com/ngalayko/p2p/instance/logger"
	"github.com/ngalayko/p2p/instance/messages"
	"github.com/ngalayko/p2p/instance/messages/client"
	"github.com/ngalayko/p2p/instance/messages/client/resolver"
	"github.com/ngalayko/p2p/instance/messages/proto/chat"
	"github.com/ngalayko/p2p/instance/messages/server"
	"github.com/ngalayko/p2p/instance/peers"
)

// Handler used to send and receive messages.
type Handler struct {
	logger *logger.Logger
	self   *peers.Peer

	srv  *grpc.Server
	port string

	messagesServer *server.Server
	creds          credentials.TransportCredentials

	clientsGuard    *sync.RWMutex
	clients         map[string]*client.Client
	clientsResolver *resolver.Builder
}

// New returns new messages handler.
func New(
	log *logger.Logger,
	self *peers.Peer,
	port string,
) *Handler {

	creds := credentials.NewServerTLSFromCert(self.Certificate)

	grpcServer := grpc.NewServer(
		grpc.Creds(creds),
	)

	messagesServer := server.New(log)
	customResolver := resolver.New(port)

	chat.RegisterChatServer(grpcServer, messagesServer)
	grpc_resolver.Register(customResolver)

	return &Handler{
		logger:          log.Prefix("messages"),
		self:            self,
		port:            port,
		srv:             grpcServer,
		creds:           creds,
		messagesServer:  messagesServer,
		clientsGuard:    &sync.RWMutex{},
		clients:         map[string]*client.Client{},
		clientsResolver: customResolver,
	}
}

// ListenAndServe starts server.
func (h *Handler) Listen(ctx context.Context) error {
	addr := fmt.Sprintf("0.0.0.0:%s", h.port)
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	h.logger.Info("starting message server on %s", addr)
	return h.srv.Serve(lis)
}

// Send sends the message to recipients.
func (h *Handler) Send(ctx context.Context, m *messages.Message, to *peers.Peer) error {
	to = h.enrich(to)

	grpcClient, err := h.getClient(ctx, to)
	if err != nil {
		return err
	}

	return grpcClient.Send(ctx, m)
}

func (h *Handler) getClient(ctx context.Context, peer *peers.Peer) (*client.Client, error) {
	h.clientsGuard.RLock()
	grpcClient, found := h.clients[peer.ID]
	h.clientsGuard.RUnlock()
	if found {
		return grpcClient, nil
	}

	h.clientsResolver.Add(peer)

	cp := x509.NewCertPool()
	if !cp.AppendCertsFromPEM(peer.PublicCrt) {
		return nil, fmt.Errorf("failed to append certificates")
	}
	creds := credentials.NewClientTLSFromCert(cp, "")

	grpcClient, err := client.Connect(ctx, h.logger, creds, peer)
	if err != nil {
		return nil, fmt.Errorf("can't create client: %s", err)
	}

	h.clientsGuard.Lock()
	h.clients[peer.ID] = grpcClient
	h.clientsGuard.Unlock()

	return grpcClient, nil
}

func (h *Handler) enrich(peer *peers.Peer) *peers.Peer {
	switch peer.ID {
	case h.self.ID:
		return h.self
	default:
		return h.self.KnownPeers.Get(peer.ID)
	}
}
