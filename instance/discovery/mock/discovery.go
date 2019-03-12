package mock

import (
	"context"

	"github.com/ngalayko/p2p/instance/peers"
)

// Mock is a mock discrovery.
type Mock struct {
	items []*peers.Peer
}

// New is a mock discovery constructor.
func New(ii ...*peers.Peer) *Mock {
	return &Mock{
		items: ii,
	}
}

// Discover implements Discovery.
func (m *Mock) Discover(context.Context) <-chan *peers.Peer {
	out := make(chan *peers.Peer)
	go func() {
		for _, i := range m.items {
			out <- i
		}
		close(out)
	}()
	return out
}
