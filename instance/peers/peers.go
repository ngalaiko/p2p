package peers

import (
	"encoding/json"
	"sync"
)

// peer is a list of peers.
type peersList struct {
	guard   *sync.RWMutex
	byID    map[string]*Peer
	updated chan bool
}

func newPeersList() *peersList {
	return &peersList{
		guard:   &sync.RWMutex{},
		byID:    map[string]*Peer{},
		updated: make(chan bool),
	}
}

// MarshalJSON returns json.
func (p peersList) MarshalJSON() ([]byte, error) {
	p.guard.RLock()
	defer p.guard.RUnlock()
	return json.Marshal(p.byID)
}

// Updated returns a chan that closes every time list gets updated.
func (p *peersList) Updated() <-chan bool {
	p.guard.RLock()
	defer p.guard.RUnlock()
	return p.updated
}

// Add adds a new peer to list.
func (p *peersList) Add(peer *Peer) {
	p.guard.Lock()
	defer p.guard.Unlock()

	if _, ok := p.byID[peer.ID]; ok {
		return
	}

	p.byID[peer.ID] = peer

	close(p.updated)
	p.updated = make(chan bool)

	return
}

// Map returns peers map.
func (p *peersList) Map() map[string]*Peer {
	p.guard.RLock()
	defer p.guard.RUnlock()
	return p.byID
}
