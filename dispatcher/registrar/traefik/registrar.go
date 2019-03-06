package traefik

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"

	"golang.org/x/sync/errgroup"

	"github.com/ngalayko/p2p/instance/peers"
	"github.com/ngalayko/p2p/logger"
)

// Registrar allows to register redirect rules in traefik.
type Registrar struct {
	logger    *logger.Logger
	client    *http.Client
	consulURL string
}

// New is a registrar constructor.
func New(
	log *logger.Logger,
	consulURL string,
) *Registrar {
	return &Registrar{
		logger:    log.Prefix("traefik"),
		client:    &http.Client{},
		consulURL: consulURL,
	}
}

// Register implements Registrar.
func (r *Registrar) Register(peer *peers.Peer, target *url.URL, host string) error {
	kv := map[string]string{
		fmt.Sprintf("traefik/backends/%s/servers/server1/url", peer.ID): target.String(),
		fmt.Sprintf("traefik/frontends/%s/backend", peer.ID):            peer.ID,
		fmt.Sprintf("traefik/frontends/%s/routes/main/rule", peer.ID):   fmt.Sprintf("Host:%s.%s", peer.ID, host),
	}

	wg := &errgroup.Group{}
	for key, value := range kv {
		k, v := key, value
		wg.Go(func() error {
			return r.consulSet(k, v)
		})
	}

	defer r.logger.Info("registered %s on %s", peer.ID, target)

	return wg.Wait()
}

func (r *Registrar) consulSet(key, value string) error {
	fullURL := fmt.Sprintf("%s/v1/kv/%s", r.consulURL, key)

	r.logger.Debug("putting %s to %s", value, fullURL)

	req, err := http.NewRequest(
		http.MethodPut,
		fullURL,
		bytes.NewBuffer([]byte(value)),
	)
	if err != nil {
		return fmt.Errorf("error making a request: %s", err)
	}

	resp, err := r.client.Do(req)
	if err != nil {
		return fmt.Errorf("error sending a request: %s", err)
	}

	if resp.StatusCode == http.StatusOK {
		return nil
	}

	body, err := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()
	if err != nil {
		return fmt.Errorf("error reading response body: %s", body)
	}

	return fmt.Errorf("error from consul: %s, status code %d", string(body), resp.StatusCode)
}
