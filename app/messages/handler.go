package messages

import (
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
	quic "github.com/lucas-clemente/quic-go"
	"github.com/lucas-clemente/quic-go/h2quic"

	"github.com/ngalayko/p2p/app/logger"
	"github.com/ngalayko/p2p/app/peers"
)

// Handler used to send and receive messages.
type Handler struct {
	logger *logger.Logger
	self   *peers.Peer

	srv      *h2quic.Server
	upgrader *websocket.Upgrader
	dialer   *websocket.Dialer
}

// New returns new messages handler.
func New(
	log *logger.Logger,
	self *peers.Peer,
	port string,
) *Handler {
	tlsConf := &tls.Config{
		Certificates: []tls.Certificate{self.Certificate},
		ClientAuth:   tls.RequireAndVerifyClientCert,
		MinVersion:   tls.VersionTLS12,
	}
	quicConf := &quic.Config{
		AcceptCookie: func(clientAddr net.Addr, cookie *quic.Cookie) bool {
			return false
		},
		HandshakeTimeout:      30 * time.Second,
		IdleTimeout:           30 * time.Second,
		MaxIncomingUniStreams: -1,
		KeepAlive:             true,
	}
	return &Handler{
		logger: log.Prefix("messages"),
		self:   self,
		srv: &h2quic.Server{
			Server: &http.Server{
				Addr:      fmt.Sprintf("0.0.0.0:%s", port),
				TLSConfig: tlsConf,
			},
			QuicConfig: quicConf,
		},
		upgrader: &websocket.Upgrader{
			EnableCompression: true,
		},
		dialer: &websocket.Dialer{
			TLSClientConfig:   tlsConf,
			EnableCompression: true,
		},
	}
}

// ServeHTTP implements http.Server.
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	_, err := h.upgrader.Upgrade(w, r, nil)
	if err != nil {
		h.logger.Error("error upgrading a connection: %s", err)
		return
	}
}

// ListenAndServe starts server.
func (h *Handler) ListenAndServe() error {
	h.logger.Info("starting message server on %s", h.srv.Server.Addr)
	return h.srv.ListenAndServe()
}
