package messages

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"sync"
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
	port     string
	upgrader *websocket.Upgrader
	dialer   *websocket.Dialer

	connectionsGuard *sync.RWMutex
	connections      map[string]*websocket.Conn
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
		port:   port,
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
			NetDial: func(network, addr string) (net.Conn, error) {
				return net.Dial("udp", addr)
			},
			TLSClientConfig:   tlsConf,
			EnableCompression: true,
		},
		connectionsGuard: &sync.RWMutex{},
		connections:      map[string]*websocket.Conn{},
	}
}

// ServeHTTP implements http.Server.
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	origin := r.Header.Get("origin")

	h.logger.Info("new connection from %s", origin)
	defer h.logger.Info("connection from %s closed", origin)

	conn, err := h.upgrader.Upgrade(w, r, nil)
	if err != nil {
		h.logger.Error("error upgrading a connection from %s: %s", origin, err)
		return
	}

	for {
		_, data, err := conn.ReadMessage()
		if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
			fmt.Printf("\nnikitag: %+v\n\n", data)
			h.logger.Error("error reading message from %s: %s", origin, err)
			continue
		}

		fmt.Printf("\nnikitag: %+v\n\n", string(data))
	}
}

// ListenAndServe starts server.
func (h *Handler) ListenAndServe() error {
	h.logger.Info("starting message server on %s", h.srv.Server.Addr)
	return h.srv.ListenAndServe()
}

// Send sends the message to recipients.
func (h *Handler) Send(m *Message) map[string]error {
	h.enrichMessage(m)

	errs := make(map[string]error, len(m.To))
	for _, r := range m.To {
		if err := h.send(m, r); err != nil {
			errs[r.ID] = err
		}
	}
	return errs
}

func (h *Handler) enrichMessage(m *Message) {
	m.From = h.self
	for i := range m.To {
		if m.To[i].ID == h.self.ID {
			m.To[i] = h.self
			continue
		}
		m.To[i] = h.self.KnownPeers.Get(m.To[i].ID)
	}
}

func (h *Handler) send(m *Message, r *peers.Peer) error {
	conn, err := h.connection(r)
	if err != nil {
		return fmt.Errorf("connection error: %s", err)
	}

	data, err := json.Marshal(m)
	if err != nil {
		return fmt.Errorf("error marshaling message: %s", err)
	}

	if err := conn.WriteMessage(websocket.TextMessage, data); err != nil {
		return fmt.Errorf("error writing message: %s", err)
	}

	h.logger.Info("message sent to %s", r.ID)

	return nil
}

func (h *Handler) connection(p *peers.Peer) (*websocket.Conn, error) {
	h.connectionsGuard.RLock()
	conn, ok := h.connections[p.ID]
	h.connectionsGuard.RUnlock()
	if ok {
		return conn, nil
	}

	for _, addr := range p.Addrs.List() {
		var err error

		a := fmt.Sprintf("ws://%s:%s", addr.String(), h.port)
		if addr.To4() == nil {
			a = fmt.Sprintf("ws://[%s]:%s", addr.String(), h.port)
		}

		conn, _, err = h.dialer.Dial(a, nil)
		if err == nil {
			h.logger.Info("opened connection with %s", a)
			break
		}
		h.logger.Error("can't dial %s: %s", addr, err)
	}
	if conn == nil {
		return nil, fmt.Errorf("can't reach recipient: %s", p.ID)
	}
	return conn, nil
}
