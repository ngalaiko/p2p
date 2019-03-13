package pool

import (
	"context"
	"fmt"
	"math/rand"
	"net/url"
	"sync"
	"time"

	"github.com/ngalayko/p2p/dispatcher/creator"
	"github.com/ngalayko/p2p/instance/peers"
	"github.com/ngalayko/p2p/logger"
)

// Pool wraps over creator.Creator and creates peers in advance in order
// to decrease start time.
type Pool struct {
	logger *logger.Logger

	c creator.Creator

	size int

	poolGuard *sync.RWMutex
	pool      map[string]*instance
}

type instance struct {
	Peer *peers.Peer
	URL  *url.URL
}

// New is a pool constructor.
func New(
	log *logger.Logger,
	c creator.Creator,
	size int,
) *Pool {
	p := &Pool{
		logger:    log.Prefix("pool"),
		c:         c,
		size:      size,
		poolGuard: &sync.RWMutex{},
		pool:      map[string]*instance{},
	}
	rand.Seed(time.Now().UnixNano())
	for i := 0; i < size; i++ {
		time.Sleep(time.Duration(rand.Intn(1000)) * time.Millisecond)
		go p.push()
	}
	return p
}

// Create implements Creator.
func (p *Pool) Create(ctx context.Context) (*peers.Peer, *url.URL, error) {
	p.poolGuard.RLock()
	l := len(p.pool)
	p.poolGuard.RUnlock()

	defer func() {
		go p.push()
	}()

	if l != 0 {
		return p.pop()
	}
	p.logger.Debug("pool is empty, fallback")
	return p.c.Create(ctx)
}

// Len returns number of peers in the pool.
func (p *Pool) Len() int {
	p.poolGuard.RLock()
	defer p.poolGuard.RUnlock()
	return len(p.pool)
}

func (p *Pool) pop() (*peers.Peer, *url.URL, error) {
	p.poolGuard.Lock()
	defer p.poolGuard.Unlock()

	for k := range p.pool {
		i := p.pool[k]
		delete(p.pool, k)

		p.logger.Debug("pop %s on %s from pool", i.Peer.ID, i.URL)

		p.logger.Debug("pool size: %d", len(p.pool))
		return i.Peer, i.URL, nil
	}

	return nil, nil, fmt.Errorf("pool is empty")
}

func (p *Pool) push() {
	p.logger.Debug("adding new instance to pool")

	peer, u, err := p.c.Create(context.Background())
	if err != nil {
		p.logger.Error("error creating new peer: %s", err)
		return
	}

	p.poolGuard.Lock()
	p.pool[peer.ID] = &instance{
		Peer: peer,
		URL:  u,
	}
	p.poolGuard.Unlock()
}
