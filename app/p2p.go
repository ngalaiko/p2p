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
	"github.com/ngalayko/p2p/app/messages"
	"github.com/ngalayko/p2p/app/peers"
	"github.com/ngalayko/p2p/app/ui"
)

// Messenger is a single instance of a p2p messenger.
type Messenger struct {
	self *peers.Peer

	logger     *logger.Logger
	discovery  discovery.Discovery
	msgHandler *messages.Handler
	ui         *ui.UI
}

// New is a messenger constructor.
func New(
	logLevel logger.Level,
	udp4Multicast string,
	udp6Multicast string,
	discoveryPort string,
	port string,
	uiPort string,
	discroverInterval time.Duration,
) *Messenger {
	log := logger.New(logLevel)
	self, err := peers.New()
	if err != nil {
		log.Panic("can't initialize peer: %s", err)
	}
	return &Messenger{
		self:   self,
		logger: log,
		discovery: merge.New(
			udp6.New(log, fmt.Sprintf("%s:%s", udp6Multicast, discoveryPort), discroverInterval, self),
			udp4.New(log, fmt.Sprintf("%s:%s", udp4Multicast, discoveryPort), discroverInterval, self),
		),
		ui:         ui.New(log, self, fmt.Sprintf("127.0.0.1:%s", uiPort)),
		msgHandler: messages.New(log, self, port),
	}
}

// Start starts a messanger instance.
func (m *Messenger) Start() error {
	go m.watchPeers()
	go m.msgHandler.ListenAndServe()
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
