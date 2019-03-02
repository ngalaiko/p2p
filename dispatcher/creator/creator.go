package creator

import "github.com/ngalayko/p2p/instance/peers"

// Creator can create new peer instances.
type Creator interface {
	Create() (*peers.Peer, error)
}
