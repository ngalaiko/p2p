package creator

import (
	"context"

	"github.com/ngalayko/p2p/instance/peers"
)

// Creator can create new peer instances.
type Creator interface {
	Create(context.Context) (*peers.Peer, error)
}
