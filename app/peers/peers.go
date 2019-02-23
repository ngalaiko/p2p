package peers

// peer is a list of peers.
type peersList struct {
	list []*Peer
	byID map[string]*Peer
}

func newPeersList() *peersList {
	return &peersList{
		byID: map[string]*Peer{},
	}
}

// Add adds a new peer to list.
func (p *peersList) Add(peer *Peer) {
	known, ok := p.byID[peer.ID]
	if !ok {
		p.list = append(p.list, peer)
		p.byID[peer.ID] = peer
		return
	}

	for _, addr := range peer.Addrs.List() {
		known.Addrs.Add(addr)
	}
}

// List retuns a list of peers.
func (p *peersList) List() []*Peer {
	return p.list
}
