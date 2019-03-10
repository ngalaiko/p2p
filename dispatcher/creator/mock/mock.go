package mock

import (
	"context"
	"net/url"

	"github.com/ngalayko/p2p/instance/peers"
)

// Mock is a mock creator.
type Mock struct {
	create func(ctx context.Context) (*peers.Peer, *url.URL, error)
}

// New is a mock creator constructor.
func New(
	create func(ctx context.Context) (*peers.Peer, *url.URL, error),
) *Mock {
	return &Mock{
		create: create,
	}
}

// Create implements Creator.
func (m *Mock) Create(ctx context.Context) (*peers.Peer, *url.URL, error) {
	return m.create(ctx)
}
