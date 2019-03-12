package consul

import (
	"context"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"strings"
	"time"

	consul "github.com/hashicorp/consul/api"

	"github.com/ngalayko/p2p/instance/peers"
	"github.com/ngalayko/p2p/logger"
)

// Discovery registers peer in consul catalog
// and returns other peers.
type Discovery struct {
	logger   *logger.Logger
	host     string
	self     *peers.Peer
	client   *consul.Client
	interval time.Duration

	httpClient *http.Client
}

// New returns new consul discovery.
func New(
	log *logger.Logger,
	self *peers.Peer,
	consulHost string,
	interval time.Duration,
) *Discovery {
	log = log.Prefix("consul")

	cfg := consul.DefaultConfig()
	cfg.Address = consulHost

	client, err := consul.NewClient(cfg)
	if err != nil {
		log.Panic("can't connect to consul: %s", err)
	}

	return &Discovery{
		logger:     log,
		host:       consulHost,
		client:     client,
		self:       self,
		interval:   interval,
		httpClient: &http.Client{},
	}
}

// Discover implements Discovery.
func (d *Discovery) Discover(ctx context.Context) <-chan *peers.Peer {
	go func() {
		if err := d.register(); err != nil {
			d.logger.Error("failed to register: %s", err)
		}
	}()

	go func() {
		<-ctx.Done()
		if err := d.deregister(); err != nil {
			d.logger.Error("failed to unregister: %s", err)
		}
	}()

	return d.discover(ctx)
}

func (d *Discovery) discover(ctx context.Context) <-chan *peers.Peer {
	out := make(chan *peers.Peer)
	go func() {
		d.logger.Info("running discovery")
		defer d.logger.Info("stopping discovery")

		for {
			select {
			case <-ctx.Done():
				close(out)
				return
			case <-time.Tick(d.interval):
				ss, err := d.client.Agent().Services()
				if err != nil {
					d.logger.Error("can't list services: %s", err)
					continue
				}
				for _, s := range ss {
					if _, ok := s.Meta["peer"]; !ok {
						continue
					}
					p, err := d.getPeer(fmt.Sprintf("http://%s:%d", s.Address, s.Port))
					if err != nil {
						d.logger.Error("can't get peer %s info: %s", s.Service, err)
						continue
					}
					p.Addrs.Add(net.ParseIP(s.Address))
					out <- p
				}
			}
		}
	}()
	return out
}

func (d *Discovery) getPeer(url string) (*peers.Peer, error) {
	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/healthcheck", url), nil)
	if err != nil {
		return nil, fmt.Errorf("error forming a request: %s", err)
	}

	response, err := d.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error making a request: %s", err)
	}
	defer response.Body.Close()

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading a body: %s", err)
	}

	peer := &peers.Peer{}
	if err := peer.Unmarshal(body); err != nil {
		return nil, err
	}

	return peer, nil
}

func (d *Discovery) register() error {
	d.logger.Info("registrating")

	iface, err := net.InterfaceByName("eth0")
	if err != nil {
		return fmt.Errorf("unable to access eth0: %s", err)
	}

	addr, err := iface.Addrs()
	if err != nil {
		return fmt.Errorf("unable to determine local addr: %s", err)
	}
	localAddr := strings.Split(addr[0].String(), "/")[0]

	reg := &consul.AgentServiceRegistration{
		Name:    d.self.ID,
		Port:    d.self.UIPort,
		Address: localAddr,
		Meta: map[string]string{
			"peer": "true",
		},
		Tags: []string{
			"traefik.backend.loadbalancer.stickiness=true",
		},
		Check: &consul.AgentServiceCheck{
			CheckID:  "ui",
			Name:     "UI access",
			HTTP:     fmt.Sprintf("http://%s:%d/", localAddr, d.self.UIPort),
			Method:   http.MethodGet,
			Interval: fmt.Sprintf("%s", 10*time.Second),
			Timeout:  fmt.Sprintf("%s", 1*time.Second),
		},
	}

	return d.client.Agent().ServiceRegister(reg)
}

func (d *Discovery) deregister() error {
	return d.client.Agent().ServiceDeregister(d.self.ID)
}
