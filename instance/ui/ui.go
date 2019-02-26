package ui

import (
	"net/http"

	"github.com/ngalayko/p2p/instance/logger"
	"github.com/ngalayko/p2p/instance/messages"
	"github.com/ngalayko/p2p/instance/peers"
	"github.com/ngalayko/p2p/instance/ui/ws"
)

// UI serves user interface.
type UI struct {
	log    *logger.Logger
	server *http.Server
	ws     *ws.WebSocket
}

// New is a ui constructor.
func New(
	log *logger.Logger,
	self *peers.Peer,
	addr string,
	msgHandler *messages.Handler,
) *UI {
	log = log.Prefix("ui")

	u := &UI{
		log: log,
		server: &http.Server{
			Addr: addr,
		},
		ws: ws.New(log, self, msgHandler),
	}
	u.server.Handler = u.handler()
	return u
}

func (u *UI) handler() http.Handler {
	m := http.NewServeMux()
	m.Handle("/", http.FileServer(http.Dir("./ui/public")))
	m.Handle("/ws", u.ws)
	return m
}

// ListenAndServe servers ui.
func (u *UI) ListenAndServe() error {
	u.log.Info("serving ui at %s", u.server.Addr)
	return u.server.ListenAndServe()
}
