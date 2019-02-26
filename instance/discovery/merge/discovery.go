package merge

import (
	"context"
	"sync"

	"github.com/ngalayko/p2p/instance/discovery"
	"github.com/ngalayko/p2p/instance/peers"
)

// Discovery merges results from several discoveries.
type Discovery struct {
	dd []discovery.Discovery
}

// New is a discovery constructor.
func New(dd ...discovery.Discovery) *Discovery {
	return &Discovery{
		dd: dd,
	}
}

// Discover implements Discovery interface.
func (d *Discovery) Discover(ctx context.Context) <-chan *peers.Peer {
	out := make(chan *peers.Peer)
	wg := &sync.WaitGroup{}
	wg.Add(len(d.dd))
	for _, d := range d.dd {
		go func(c <-chan *peers.Peer) {
			for v := range c {
				out <- v
			}
			wg.Done()
		}(d.Discover(ctx))
	}
	go func() {
		wg.Wait()
		close(out)
	}()
	return out
}
