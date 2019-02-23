package p2p

import (
	"context"
	"time"

	"github.com/ngalayko/p2p/app/discovery"
	"github.com/ngalayko/p2p/app/discovery/merge"
	"github.com/ngalayko/p2p/app/discovery/udp4"
	"github.com/ngalayko/p2p/app/discovery/udp6"
	"github.com/ngalayko/p2p/app/logger"
	"github.com/ngalayko/p2p/app/peers"
)

// Messenger is a single instance of a p2p messenger.
type Messenger struct {
	self *peers.Peer

	logger    *logger.Logger
	discovery discovery.Discovery
}

// New is a messenger constructor.
func New(
	logLevel logger.Level,
	udp4Multicast string,
	udp6Multicast string,
	discroverInterval time.Duration,
) *Messenger {
	l := logger.New(logLevel)
	self := peers.New()
	return &Messenger{
		self:   self,
		logger: l,
		discovery: merge.New(
			udp6.New(l, udp6Multicast, discroverInterval, self),
			udp4.New(l, udp4Multicast, discroverInterval, self),
		),
	}
}

// Start starts a messanger instance.
func (m *Messenger) Start() error {
	go m.watchPeers()
	for {
	}
	return nil
}

func (m *Messenger) watchPeers() {
	for peer := range m.discovery.Discover(context.Background()) {
		if peer.ID != m.self.ID {
			m.self.KnownPeers.Add(peer)
			continue
		}

		for _, addr := range peer.Addrs.List() {
			m.self.Addrs.Add(addr)
		}
	}
}
