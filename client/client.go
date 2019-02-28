package client

import (
	"net/http"

	"github.com/ngalayko/p2p/client/ws"
	"github.com/ngalayko/p2p/instance"
	"github.com/ngalayko/p2p/instance/logger"
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
	addr string,
	instance *instance.Instance,
) *UI {
	log = log.Prefix("ui")

	u := &UI{
		log: log,
		server: &http.Server{
			Addr: addr,
		},
		ws: ws.New(log, instance),
	}
	u.server.Handler = u.handler()
	return u
}

func (u *UI) handler() http.Handler {
	m := http.NewServeMux()
	m.Handle("/", http.FileServer(http.Dir("./client/public")))
	m.Handle("/ws", u.ws)
	return m
}

// ListenAndServe servers ui.
func (u *UI) ListenAndServe() error {
	u.log.Info("serving ui at %s", u.server.Addr)
	return u.server.ListenAndServe()
}
