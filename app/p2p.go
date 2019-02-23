package p2p

import (
	"context"
	"fmt"
	"time"

	"github.com/ngalayko/p2p/app/discovery"
	"github.com/ngalayko/p2p/app/discovery/merge"
	"github.com/ngalayko/p2p/app/discovery/udp4"
	"github.com/ngalayko/p2p/app/discovery/udp6"
	"github.com/ngalayko/p2p/app/logger"
	"github.com/ngalayko/p2p/app/peers"
	"github.com/ngalayko/p2p/app/ui"
)

// Messenger is a single instance of a p2p messenger.
type Messenger struct {
	self *peers.Peer

	logger    *logger.Logger
	discovery discovery.Discovery
	ui        *ui.UI
}

// New is a messenger constructor.
func New(
	logLevel logger.Level,
	udp4Multicast string,
	udp6Multicast string,
	port string,
	uiPort string,
	discroverInterval time.Duration,
) *Messenger {
	l := logger.New(logLevel)
	self := peers.New()
	return &Messenger{
		self:   self,
		logger: l,
		discovery: merge.New(
			udp6.New(l, fmt.Sprintf("%s:%s", udp6Multicast, port), discroverInterval, self),
			udp4.New(l, fmt.Sprintf("%s:%s", udp4Multicast, port), discroverInterval, self),
		),
		ui: ui.New(l, self, fmt.Sprintf("127.0.0.1:%s", uiPort)),
	}
}

// Start starts a messanger instance.
func (m *Messenger) Start() error {
	go m.watchPeers()
	return m.ui.ListenAndServe()
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
