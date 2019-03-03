package client

import (
	"context"
	"net/http"

	"github.com/ngalayko/p2p/client/ws"
	"github.com/ngalayko/p2p/instance"
	"github.com/ngalayko/p2p/logger"
)

// UI serves user interface.
type UI struct {
	logger   *logger.Logger
	server   *http.Server
	ws       *ws.WebSocket
	instance *instance.Instance
}

// New is a ui constructor.
func New(
	log *logger.Logger,
	addr string,
	instance *instance.Instance,
	staticPath string,
) *UI {
	log = log.Prefix("ui")

	u := &UI{
		logger: log,
		server: &http.Server{
			Addr: addr,
		},
		ws:       ws.New(log, instance),
		instance: instance,
	}
	u.server.Handler = u.handler(staticPath)
	return u
}

func (u *UI) handler(staticPath string) http.Handler {
	m := http.NewServeMux()
	m.Handle("/", http.FileServer(http.Dir(staticPath)))
	m.Handle("/healthcheck", healthcheckHandler(u.instance.Peer))
	m.Handle("/ws", u.ws)
	return m
}

// Start servers ui.
func (u *UI) Start(ctx context.Context) error {
	go func() {
		u.logger.Info("serving ui at %s", u.server.Addr)
		defer u.logger.Info("stopping ui")
		if err := u.server.ListenAndServe(); err != nil {
			u.logger.Error("error serving ui: %s", err)
		}
	}()

	<-ctx.Done()

	u.server.Shutdown(ctx)
	return nil
}
