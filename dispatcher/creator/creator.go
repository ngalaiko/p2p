package creator

import (
	"context"
	"net/url"

	"github.com/ngalayko/p2p/instance/peers"
)

// Creator can create new peer instances.
type Creator interface {
	Create(context.Context) (*peers.Peer, *url.URL, error)
}
