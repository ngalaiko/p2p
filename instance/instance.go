package instance

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"github.com/ngalayko/p2p/instance/discovery"
	"github.com/ngalayko/p2p/instance/discovery/merge"
	"github.com/ngalayko/p2p/instance/discovery/udp4"
	"github.com/ngalayko/p2p/instance/logger"
	"github.com/ngalayko/p2p/instance/messages"
	"github.com/ngalayko/p2p/instance/peers"
	"github.com/ngalayko/p2p/instance/ui"
)

// Instance is a single instance of a p2p messenger.
type Instance struct {
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
	insecurePort string,
	uiPort string,
	discroverInterval time.Duration,
) *Instance {
	log := logger.New(logLevel)

	r := rand.New(rand.NewSource(time.Now().Unix()))

	self, err := peers.New(r, port, insecurePort)
	if err != nil {
		log.Panic("can't initialize peer: %s", err)
	}

	log.Info("peer id: %s", self.ID)

	msgHandler := messages.NewHandler(r, log, self)

	return &Instance{
		self:   self,
		logger: log,
		discovery: merge.New(
			//udp6.New(log, fmt.Sprintf("%s:%s", udp6Multicast, discoveryPort), discroverInterval, self),
			udp4.New(log, fmt.Sprintf("%s:%s", udp4Multicast, discoveryPort), discroverInterval, self),
		),
		ui:         ui.New(log, self, fmt.Sprintf("127.0.0.1:%s", uiPort), msgHandler),
		msgHandler: msgHandler,
	}
}

// Start starts a messanger instance.
func (i *Instance) Start(ctx context.Context) error {
	go i.watchPeers(ctx)
	go func() {
		if err := i.msgHandler.Start(ctx); err != nil {
			i.logger.Panic("message server: %s", err)
		}
	}()
	return i.ui.ListenAndServe()
}

func (i *Instance) watchPeers(ctx context.Context) {
	newPeers := i.discovery.Discover(ctx)

	for {
		select {
		case <-ctx.Done():
			return
		case peer := <-newPeers:
			if peer.ID != i.self.ID {
				i.self.KnownPeers.Add(peer)
				continue
			}

			for _, addr := range peer.Addrs.Map() {
				i.self.Addrs.Add(addr)
			}
		}
	}
}
