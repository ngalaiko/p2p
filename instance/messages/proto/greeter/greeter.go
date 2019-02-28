package greeter

import (
	fmt "fmt"
	"net"

	"github.com/ngalayko/p2p/instance/peers"
)

// UnmarshalPeer unmarshals gRPC peer from peer.
func (g *Peer) UnmarshalPeer(p *peers.Peer) error {
	g.ID = p.ID
	g.Name = p.Name
	g.PublicKey = p.PublicCrt

	for _, ip := range p.Addrs.Map() {
		g.IPs = append(g.IPs, ip.String())
	}
	for _, p := range p.KnownPeers.Map() {
		peer := &Peer{}
		if err := peer.UnmarshalPeer(p); err != nil {
			return fmt.Errorf("error unmarshaling known peer: %s", err)
		}
		g.KnownPeers = append(g.KnownPeers, peer)
	}
	return nil
}

// MarshalPeer returns peer from gRPC peer.
func (g Peer) MarshalPeer() (*peers.Peer, error) {
	peer := peers.NewBlank()
	peer.ID = g.ID
	peer.Name = g.Name
	peer.PublicCrt = g.PublicKey
	for _, ip := range g.IPs {
		peer.Addrs.Add(net.ParseIP(ip))
	}
	for _, p := range g.KnownPeers {
		peer, err := p.MarshalPeer()
		if err != nil {
			return nil, fmt.Errorf("error marshaling known peer %s", p.ID)
		}

		peer.KnownPeers.Add(peer)
	}
	return peer, nil
}
