package instance

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"github.com/ngalayko/p2p/instance/discovery"
	"github.com/ngalayko/p2p/instance/discovery/merge"
	"github.com/ngalayko/p2p/instance/discovery/udp4"
	"github.com/ngalayko/p2p/instance/discovery/udp6"
	"github.com/ngalayko/p2p/instance/messages"
	"github.com/ngalayko/p2p/instance/peers"
	"github.com/ngalayko/p2p/logger"
)

// Instance is a single instance of a p2p messenger.
type Instance struct {
	*messages.Handler

	*peers.Peer

	logger    *logger.Logger
	discovery discovery.Discovery
}

// New is a messenger constructor.
func New(
	log *logger.Logger,
	udp4Multicast string,
	udp6Multicast string,
	discoveryPort string,
	port string,
	insecurePort string,
	discroverInterval time.Duration,
	keySize int,
) *Instance {
	r := rand.New(rand.NewSource(time.Now().Unix()))

	self, err := peers.New(r, port, insecurePort, keySize)
	if err != nil {
		log.Panic("can't initialize peer: %s", err)
	}

	log.Info("peer id: %s", self.ID)

	msgHandler := messages.NewHandler(r, log, self)

	dd := []discovery.Discovery{}
	if udp6Multicast != "" {
		dd = append(dd, udp6.New(log, fmt.Sprintf("%s:%s", udp6Multicast, discoveryPort), discroverInterval, self))
	}
	if udp4Multicast != "" {
		dd = append(dd, udp4.New(log, fmt.Sprintf("%s:%s", udp4Multicast, discoveryPort), discroverInterval, self))
	}

	return &Instance{
		Handler:   msgHandler,
		Peer:      self,
		logger:    log,
		discovery: merge.New(dd...),
	}
}

// Start starts a messanger instance.
func (i *Instance) Start(ctx context.Context) error {
	go i.watchPeers(ctx)
	return i.Handler.Start(ctx)
}

func (i *Instance) watchPeers(ctx context.Context) {
	newPeers := i.discovery.Discover(ctx)

	for {
		select {
		case <-ctx.Done():
			return
		case peer := <-newPeers:
			if peer.ID != i.ID {
				i.KnownPeers.Add(peer)
				continue
			}

			for _, addr := range peer.Addrs.Map() {
				i.Addrs.Add(addr)
			}
		}
	}
}
