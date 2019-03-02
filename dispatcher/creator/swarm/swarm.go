package swarm

import "github.com/ngalayko/p2p/instance/peers"

// Swarm can create a new peer service in the docker swarm.
type Swarm struct{}

// New is a swarm creator constructor.
func New() *Swarm {
	return &Swarm{}
}

// Create implements Creator.
func (s *Swarm) Create() (*peers.Peer, error) {
	return &peers.Peer{
		ID: "test",
	}, nil
	return nil, nil
}
