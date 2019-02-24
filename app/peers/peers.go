package peers

import (
	"encoding/json"
	"sync"
)

// peer is a list of peers.
type peersList struct {
	guard *sync.RWMutex

	list    []*Peer
	byID    map[string]*Peer
	updated chan *Peer
}

func newPeersList() *peersList {
	return &peersList{
		guard:   &sync.RWMutex{},
		byID:    map[string]*Peer{},
		updated: make(chan *Peer),
	}
}

// MarshalJSON returns json.
func (p peersList) MarshalJSON() ([]byte, error) {
	return json.Marshal(p.Map())
}

// Updated returns a chan that closes every time list gets updated.
func (p *peersList) Updated() <-chan *Peer {
	p.guard.RLock()
	defer p.guard.RUnlock()
	return p.updated
}

// Add adds a new peer to list.
func (p *peersList) Add(peer *Peer) {
	p.guard.Lock()
	defer p.guard.Unlock()

	known, ok := p.byID[peer.ID]
	if !ok {
		p.list = append(p.list, peer)
		p.byID[peer.ID] = peer
		p.updated <- peer
		return
	}

	wasUpdated := false
	for _, addr := range peer.Addrs.List() {
		wasUpdated = wasUpdated || known.Addrs.Add(addr)
	}

	if wasUpdated {
		p.updated <- peer
	}
}

// Map retuns a map of peers.
func (p *peersList) Map() map[string]*Peer {
	p.guard.Lock()
	defer p.guard.Unlock()
	return p.byID
}

// List retuns a list of peers.
func (p *peersList) List() []*Peer {
	p.guard.Lock()
	defer p.guard.Unlock()
	return p.list
}
