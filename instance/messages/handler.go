package messages

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
	"github.com/ngalayko/p2p/instance/messages/client"
	"github.com/ngalayko/p2p/instance/messages/client/resolver"
	"github.com/ngalayko/p2p/instance/messages/proto/chat"
	"github.com/ngalayko/p2p/instance/messages/server"
	"github.com/ngalayko/p2p/instance/peers"
)

// Handler runs message server and holds connection to known peers.
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

// NewHandler returns new messages handler.
func NewHandler(
	log *logger.Logger,
	self *peers.Peer,
	port string,
) *Handler {

	creds := credentials.NewServerTLSFromCert(self.Certificate)

	grpcServer := grpc.NewServer(
		grpc.Creds(creds),
	)

	messagesServer := server.New(log)
	customResolver := resolver.New()

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

// Listen starts server.
func (h *Handler) Listen(ctx context.Context) error {
	addr := net.JoinHostPort("0.0.0.0", h.port)
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	h.logger.Info("starting message server on %s", addr)
	return h.srv.Serve(lis)
}

// SendText sends the text message.
func (h *Handler) SendText(ctx context.Context, text string, toID string) error {
	to, err := h.getPeer(toID)
	if err != nil {
		return err
	}

	grpcClient, err := h.getClient(ctx, to)
	if err != nil {
		return err
	}

	return grpcClient.SendText(ctx, text)
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
