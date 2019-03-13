package swarm

import (
	"context"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/swarm"
	"github.com/docker/docker/client"
	consul "github.com/hashicorp/consul/api"

	"github.com/ngalayko/p2p/instance/peers"
	"github.com/ngalayko/p2p/logger"
)

// Swarm can create a new peer service in the docker swarm.
type Swarm struct {
	logger *logger.Logger

	swarmCli  *client.Client
	consulCli *consul.Client

	peerServiceName string

	creatingGuard *sync.Mutex

	newPeers chan *peers.Peer
}

// New is a swarm creator constructor.
func New(
	ctx context.Context,
	log *logger.Logger,
	peerServiceName string,
	consulAddr string,
) *Swarm {
	log = log.Prefix("swarm")

	swarmCli, err := client.NewEnvClient()
	if err != nil {
		log.Panic("failed to create docker client: %s", err)
	}

	consulCfg := consul.DefaultConfig()
	consulCfg.Address = consulAddr

	consulCli, err := consul.NewClient(consulCfg)
	if err != nil {
		log.Panic("failed to create consul client: %s", err)
	}

	s := &Swarm{
		logger:          log,
		swarmCli:        swarmCli,
		consulCli:       consulCli,
		peerServiceName: peerServiceName,
		newPeers:        make(chan *peers.Peer),
		creatingGuard:   &sync.Mutex{},
	}

	go func() {
		if err := s.watchPeers(ctx); err != nil {
			s.logger.Error("error watching peers: %s", err)
		}
	}()

	return s
}

// Create implements Creator.
// creates a new docker service in a swarm cluster.
func (s *Swarm) Create(ctx context.Context) (*peers.Peer, *url.URL, error) {
	peerService, err := s.getPeersService(ctx)
	if err != nil {
		return nil, nil, fmt.Errorf("can't get peer service: %s", err)
	}

	before := *peerService.Spec.Mode.Replicated.Replicas
	after := before + 1

	s.logger.Info("scaling %s from %d to %d", s.peerServiceName, before, after)

	*peerService.Spec.Mode.Replicated.Replicas++

	s.creatingGuard.Lock()
	resp, err := s.swarmCli.ServiceUpdate(
		ctx,
		peerService.ID,
		swarm.Version{
			Index: peerService.Version.Index,
		},
		peerService.Spec,
		types.ServiceUpdateOptions{},
	)
	s.creatingGuard.Unlock()

	if err != nil {
		return nil, nil, fmt.Errorf("failed to update service %s: %s", s.peerServiceName, err)
	}

	for _, w := range resp.Warnings {
		s.logger.Warning(w)
	}

	start := time.Now()
	newPeer := <-s.newPeers
	s.logger.Info("peer %s took %s to start", newPeer.ID, time.Since(start))

	for _, ip := range newPeer.Addrs.Map() {
		u, err := url.Parse(fmt.Sprintf("http://%s:%d", ip, newPeer.UIPort))
		if err != nil {
			return nil, nil, fmt.Errorf("can't parse url for %s: %s", newPeer.ID, err)
		}
		return newPeer, u, nil
	}
	return nil, nil, fmt.Errorf("there are no known addreses for %s", newPeer.ID)
}

func (s *Swarm) queryServices(ctx context.Context) (map[string]*consul.AgentService, error) {
	for {
		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("cancelled")
		case <-time.Tick(time.Second):
			ss, err := s.consulCli.Agent().Services()
			if err == nil {
				return ss, nil
			}
			s.logger.Error("failed to list consul services: %s", err)
		}
	}
}

func (s *Swarm) watchPeers(ctx context.Context) error {
	ss, err := s.queryServices(ctx)
	if err != nil {
		return fmt.Errorf("failed to list consul services: %s", err)
	}

	knownServices := make(map[string]bool, len(ss))
	for _, s := range ss {
		knownServices[s.ID] = true
	}

	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("cancelled")
		case <-time.Tick(time.Second):
			services, err := s.consulCli.Agent().Services()
			if err != nil {
				s.logger.Error("failed to list consul services: %s", err)
				continue
			}
			for _, service := range services {
				if knownServices[service.ID] {
					continue
				}

				peer, err := s.parsePeer(service)
				if err != nil {
					s.logger.Error("error parsing peer %s: %s", service.ID, err)
					continue
				}

				knownServices[service.ID] = true

				go func() {
					s.newPeers <- peer
				}()
			}
		}
	}
}

func (s *Swarm) parsePeer(service *consul.AgentService) (*peers.Peer, error) {
	response, err := http.Get(fmt.Sprintf("http://%s:%d/healthcheck", service.Address, service.Port))
	if err != nil {
		return nil, fmt.Errorf("can't get healthcheck response from %s: %s", service.ID, err)
	}

	bytes, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading %s response: %s", service.ID, err)
	}
	defer response.Body.Close()

	newPeer := &peers.Peer{}
	if err := newPeer.Unmarshal(bytes); err != nil {
		return nil, fmt.Errorf("can't unmarshal response from %s: %s", service.ID, err)
	}

	newPeer.Addrs.Add(net.IP(service.Address))
	return newPeer, nil

}

func (s *Swarm) getPeersService(ctx context.Context) (*swarm.Service, error) {
	for {
		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("cancelled")
		case <-time.Tick(time.Second):
			services, err := s.swarmCli.ServiceList(ctx, types.ServiceListOptions{})
			if err != nil {
				return nil, fmt.Errorf("failed to list services: %s", err)
			}

			for _, service := range services {
				if service.Spec.Name == s.peerServiceName {
					return &service, nil
				}
			}

			continue
		}
	}
}
