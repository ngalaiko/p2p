package discovery

import (
	"context"

	"github.com/ngalayko/p2p/instance/peers"
)

// Discovery used to search for peers.
type Discovery interface {
	Discover(context.Context) <-chan *peers.Peer
}
