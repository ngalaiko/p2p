package dispatcher

import (
	"context"
	"fmt"
	"net"
	"net/http"

	"github.com/ngalayko/p2p/dispatcher/auth"
	"github.com/ngalayko/p2p/dispatcher/auth/jwt"
	"github.com/ngalayko/p2p/dispatcher/creator"
	"github.com/ngalayko/p2p/dispatcher/creator/swarm"
	"github.com/ngalayko/p2p/dispatcher/registrar"
	"github.com/ngalayko/p2p/dispatcher/registrar/traefik"
	"github.com/ngalayko/p2p/logger"
)

// Dispatcher redirect clients to it's peers.
type Dispatcher struct {
	logger *logger.Logger

	host string

	creator    creator.Creator
	authorizer auth.Authorizer
	registrar  registrar.Registrar
}

// New is a discpatcher constructor.
func New(
	ctx context.Context,
	log *logger.Logger,
	jwtSecret string,
	host string,
	peerImageName string,
	peerNetworkName string,
	consulURL string,
) *Dispatcher {
	return &Dispatcher{
		logger: log.Prefix("dispatcher"),

		host: host,

		creator:    swarm.New(ctx, log, peerImageName, peerNetworkName),
		authorizer: jwt.New(jwtSecret),
		registrar:  traefik.New(log, consulURL),
	}
}

// Start servers the dispatcher.
func (d *Dispatcher) Start(ctx context.Context, port string) error {
	addr := net.JoinHostPort("0.0.0.0", port)
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("failed to listen to :%s", port)
	}

	mux := http.NewServeMux()
	mux.Handle("/", d.mainHandler())

	srv := &http.Server{
		Handler: mux,
	}
	go func() {
		d.logger.Info("starting server on %s", addr)
		defer d.logger.Info("shutdown server")
		if err := srv.Serve(lis); err != nil {
			d.logger.Error("serve error: %s", err)
		}
	}()

	<-ctx.Done()

	return srv.Shutdown(ctx)
}

// ServeHTTP implements http.Handler.
func (d *Dispatcher) mainHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		peer, _ := d.authorizer.Get(r)
		if peer != nil {
			w.Write([]byte("found " + peer.ID))
			return
		}

		peer, url, err := d.creator.Create(r.Context())
		if err != nil {
			d.responseError(w, fmt.Errorf("error creating a peer: %s", err))
			return
		}

		if err := d.registrar.Register(peer, url, d.host); err != nil {
			d.responseError(w, fmt.Errorf("error registring a peer: %s", err))
			return
		}

		if err := d.authorizer.Store(w, peer); err != nil {
			d.responseError(w, fmt.Errorf("error setting token: %s", err))
			return
		}

		w.Write([]byte("new " + peer.ID))
	}
}

func (d *Dispatcher) responseError(w http.ResponseWriter, err error) {
	d.logger.Error("serving error: %s", err)

	w.WriteHeader(http.StatusInternalServerError)
	_, _ = w.Write([]byte(err.Error()))
}
