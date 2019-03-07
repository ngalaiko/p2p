package dispatcher

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/url"

	"github.com/ngalayko/p2p/dispatcher/auth"
	"github.com/ngalayko/p2p/dispatcher/auth/jwt"
	"github.com/ngalayko/p2p/dispatcher/creator"
	"github.com/ngalayko/p2p/dispatcher/creator/swarm"
	"github.com/ngalayko/p2p/dispatcher/registrar"
	"github.com/ngalayko/p2p/dispatcher/registrar/traefik"
	"github.com/ngalayko/p2p/instance/peers"
	"github.com/ngalayko/p2p/logger"
)

// Dispatcher redirect clients to it's peers.
type Dispatcher struct {
	logger *logger.Logger

	creator    creator.Creator
	authorizer auth.Authorizer
	registrar  registrar.Registrar
}

// New is a discpatcher constructor.
func New(
	ctx context.Context,
	log *logger.Logger,
	jwtSecret string,
	peerImageName string,
	peerNetworkName string,
	consulURL string,
) *Dispatcher {
	return &Dispatcher{
		logger: log.Prefix("dispatcher"),

		creator:    swarm.New(ctx, log, peerImageName, peerNetworkName),
		authorizer: jwt.New(log, jwtSecret),
		registrar:  traefik.New(log, consulURL),
	}
}

// Start servers the dispatcher.
func (d *Dispatcher) Start(ctx context.Context, port string, staticPath string) error {
	addr := net.JoinHostPort("0.0.0.0", port)
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("failed to listen to :%s", port)
	}

	mux := http.NewServeMux()
	mux.Handle("/", http.FileServer(http.Dir(staticPath)))
	mux.Handle("/dispatch", d.dispatchHandler())

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

func (d *Dispatcher) dispatchHandler() http.HandlerFunc {
	type response struct {
		Error    error  `json:"error,omitempty"`
		Hostname string `json:"hostname,omitempty"`
	}

	responseError := func(w http.ResponseWriter, err error) {
		d.logger.Error("serving error: %s", err)

		bytes, _ := json.Marshal(&response{
			Error: err,
		})

		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write(bytes)
	}

	responseRedirect := func(w http.ResponseWriter, r *http.Request, u *url.URL) {
		d.logger.Debug("redirecting to %s", u)
		http.Redirect(w, r, u.String(), http.StatusSeeOther)
	}

	uiURL := func(ref *url.URL, p *peers.Peer) *url.URL {
		u, err := url.Parse(fmt.Sprintf("%s://%s.%s", ref.Scheme, p.ID, ref.Host))
		if err != nil {
			d.logger.Error("error making redirect link: %s", err)
		}
		return u
	}

	return func(w http.ResponseWriter, r *http.Request) {
		referer, err := url.Parse(r.Referer())
		if err != nil {
			responseError(w, fmt.Errorf("can't parse referer: %s", err))
			return
		}

		peer, err := d.authorizer.Get(r)
		if err == nil {
			responseRedirect(w, r, uiURL(referer, peer))
			return
		}

		peer, url, err := d.creator.Create(r.Context())
		if err != nil {
			responseError(w, fmt.Errorf("error creating a peer: %s", err))
			return
		}

		if err := d.registrar.Register(peer, url, r.URL.Host); err != nil {
			responseError(w, fmt.Errorf("error registring a peer: %s", err))
			return
		}

		if err := d.authorizer.Store(w, peer); err != nil {
			responseError(w, fmt.Errorf("error setting token: %s", err))
			return
		}

		responseRedirect(w, r, uiURL(referer, peer))
	}
}
