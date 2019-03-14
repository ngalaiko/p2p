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

	peersGuard *sync.RWMutex
	knownPeers map[string]bool
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

	if _, err := consulCli.Status().Leader(); err != nil {
		log.Panic("can't connect to consul: %s", err)
	}

	swarm := &Swarm{
		logger:          log,
		swarmCli:        swarmCli,
		consulCli:       consulCli,
		peerServiceName: peerServiceName,
		peersGuard:      &sync.RWMutex{},
	}

	ss, err := swarm.queryServices(ctx)
	if err != nil {
		log.Panic("failed to list consul services: %s", err)
	}

	swarm.knownPeers = make(map[string]bool, len(ss))
	for _, s := range ss {
		swarm.knownPeers[s.ID] = false
	}

	return swarm
}

// Create implements Creator.
// creates a new docker service in a swarm cluster.
func (s *Swarm) Create(ctx context.Context) (*peers.Peer, *url.URL, error) {
	if err := s.createPeer(ctx); err != nil {
		return nil, nil, fmt.Errorf("error creating peer: %s", err)
	}

	start := time.Now()

	newPeer, u, err := s.waitForNewPeer(ctx)
	if err != nil {
		return nil, nil, fmt.Errorf("error while waiting for a new peer: %s", err)
	}

	s.logger.Info("peer %s took %s to start", newPeer.ID, time.Since(start))
	return newPeer, u, nil
}

func (s *Swarm) createPeer(ctx context.Context) error {
	s.peersGuard.RLock()

	s.logger.Debug("looking for unused peers")
	for _, inUse := range s.knownPeers {
		if inUse {
			continue
		}

		s.logger.Debug("found unused peer")

		s.peersGuard.RUnlock()

		return nil
	}
	s.peersGuard.RUnlock()

	s.logger.Debug("no unused peers")

	peerService, err := s.getPeersService(ctx)
	if err != nil {
		return fmt.Errorf("can't get peer service: %s", err)
	}

	before := *peerService.Spec.Mode.Replicated.Replicas
	after := before + 1

	s.logger.Info("scaling %s from %d to %d", s.peerServiceName, before, after)

	*peerService.Spec.Mode.Replicated.Replicas++

	resp, err := s.swarmCli.ServiceUpdate(
		ctx,
		peerService.ID,
		swarm.Version{
			Index: peerService.Version.Index,
		},
		peerService.Spec,
		types.ServiceUpdateOptions{},
	)
	if err != nil {
		return fmt.Errorf("failed to update service %s: %s", s.peerServiceName, err)
	}

	for _, w := range resp.Warnings {
		s.logger.Warning(w)
	}

	return nil
}

func (s *Swarm) waitForNewPeer(ctx context.Context) (*peers.Peer, *url.URL, error) {
	s.logger.Info("waiting for a new peer")

	getNew := func(ctx context.Context) (*peers.Peer, *url.URL, error) {
		services, err := s.consulCli.Agent().Services()
		if err != nil {
			return nil, nil, fmt.Errorf("failed to list consul services: %s", err)
		}
		for _, service := range services {
			p, u, err := s.processPeer(service)
			if err != nil {
				s.logger.Error("%s: %s", service.ID, err)
				continue
			}
			return p, u, nil
		}
		return nil, nil, fmt.Errorf("no valid peers found")
	}

	if p, u, err := getNew(ctx); err == nil {
		return p, u, nil
	}

	for {
		select {
		case <-ctx.Done():
			return nil, nil, fmt.Errorf("cancelled")
		case <-time.Tick(time.Second):
			p, u, err := getNew(ctx)
			if err == nil {
				return p, u, nil
			}
			s.logger.Error("failed to get new peer: %s", err)
		}
	}
}

func (s *Swarm) processPeer(service *consul.AgentService) (*peers.Peer, *url.URL, error) {
	s.peersGuard.Lock()
	defer s.peersGuard.Unlock()

	if s.knownPeers[service.ID] {
		return nil, nil, fmt.Errorf("already in use")
	}

	peer, u, err := s.parsePeer(service)
	if err != nil {
		return nil, nil, fmt.Errorf("error parsing peer %s: %s", service.ID, err)
	}

	s.knownPeers[service.ID] = true

	return peer, u, nil
}

func (s *Swarm) queryServices(ctx context.Context) (map[string]*consul.AgentService, error) {
	ss, err := s.consulCli.Agent().Services()
	if err == nil {
		return ss, nil
	}

	s.logger.Info("waiting for consul")

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

func (s *Swarm) parsePeer(service *consul.AgentService) (*peers.Peer, *url.URL, error) {
	host := fmt.Sprintf("http://%s:%d", service.Address, service.Port)
	response, err := http.Get(fmt.Sprintf("%s/healthcheck", host))
	if err != nil {
		return nil, nil, fmt.Errorf("can't get healthcheck response from %s: %s", service.ID, err)
	}

	bytes, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, nil, fmt.Errorf("error reading %s response: %s", service.ID, err)
	}
	defer response.Body.Close()

	newPeer := &peers.Peer{}
	if err := newPeer.Unmarshal(bytes); err != nil {
		return nil, nil, fmt.Errorf("can't unmarshal response from %s: %s", service.ID, err)
	}

	newPeer.Addrs.Add(net.IP(service.Address))

	u, err := url.Parse(host)
	if err != nil {
		return nil, nil, fmt.Errorf("can't parse url: %s", err)
	}

	return newPeer, u, nil
}

func (s *Swarm) getPeersService(ctx context.Context) (*swarm.Service, error) {
	get := func(ctx context.Context) (*swarm.Service, error) {
		services, err := s.swarmCli.ServiceList(ctx, types.ServiceListOptions{})
		if err != nil {
			return nil, fmt.Errorf("failed to list services: %s", err)
		}

		for _, service := range services {
			if service.Spec.Name == s.peerServiceName {
				return &service, nil
			}
		}
		return nil, fmt.Errorf("service %s not found", s.peerServiceName)
	}

	if s, err := get(ctx); err == nil {
		return s, nil
	}

	for {
		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("cancelled")
		case <-time.Tick(time.Second):
			service, err := get(ctx)
			if err != nil {
				s.logger.Error("failed to get service: %s", s.peerServiceName)
				continue
			}
			return service, nil
		}
	}
}
