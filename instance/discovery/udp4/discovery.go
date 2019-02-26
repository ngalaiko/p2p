package udp4

import (
	"context"
	"net"
	"time"

	"golang.org/x/net/ipv4"

	"github.com/ngalayko/p2p/instance/logger"
	"github.com/ngalayko/p2p/instance/peers"
)

const (
	maxDatagramSize = 66507
)

// Discovery is a udp4 multicast discovery.
type Discovery struct {
	log      *logger.Logger
	addr     *net.UDPAddr
	interval time.Duration
	payload  []byte
}

// New returns a new discovery instance.
func New(
	log *logger.Logger,
	addr string,
	interval time.Duration,
	self *peers.Peer,
) *Discovery {

	log = log.Prefix("udp4-discovery")

	a, err := net.ResolveUDPAddr("udp4", addr)
	if err != nil {
		log.Panic("can't resolve %s: %s", addr, err)
	}

	payload, err := self.Marshal()
	if err != nil {
		log.Panic("can't marshal payload: %s", err)
	}

	return &Discovery{
		log:      log,
		addr:     a,
		interval: interval,
		payload:  payload,
	}
}

// Discover implements Discovery interface.
func (d *Discovery) Discover(ctx context.Context) <-chan *peers.Peer {
	out := make(chan *peers.Peer)
	go d.broadcast(ctx, d.interval)
	go d.listen(ctx, out)
	return out
}

func (d *Discovery) broadcast(ctx context.Context, interval time.Duration) {
	c, err := net.ListenUDP("udp4", d.addr)
	if err != nil {
		d.log.Error("can't create listen packet: %s", err)
		return
	}
	defer c.Close()

	pconn := ipv4.NewPacketConn(c)

	d.log.Info("start broadcasting to %s", d.addr)

	for {
		select {
		case <-ctx.Done():
			d.log.Info("stop broadcasting")
			return
		case <-time.Tick(interval):
			ifaces, err := net.Interfaces()
			if err != nil {
				d.log.Error("can't list interfaces", err)
				return
			}

			for i := range ifaces {
				pconn.JoinGroup(&ifaces[i], d.addr)

				if err := pconn.SetMulticastInterface(&ifaces[i]); err != nil {
					// d.log.Error("can't set multicast interface %s: %s", ifaces[i].Name, err)
					continue
				}

				pconn.SetMulticastTTL(2)

				if _, err := pconn.WriteTo(d.payload, nil, d.addr); err != nil {
					// d.log.Error("can't send multicast message: %s", err)
					continue
				}

				d.log.Debug("sent broadcast to %s", ifaces[i].Name)

				pconn.LeaveGroup(&ifaces[i], d.addr)
			}
		}
	}
}

func (d *Discovery) listen(ctx context.Context, out chan *peers.Peer) {
	conn, err := net.ListenUDP("udp4", d.addr)
	if err != nil {
		return
	}
	defer conn.Close()

	pconn := ipv4.NewPacketConn(conn)

	buf := make([]byte, maxDatagramSize)
	for {
		select {
		case <-ctx.Done():
			d.log.Info("stop listening")
			return

		default:
			ifaces, err := net.Interfaces()
			if err != nil {
				d.log.Error("can't list interfaces")
				return
			}

			for i := range ifaces {
				pconn.JoinGroup(&ifaces[i], d.addr)
			}

			n, _, src, err := pconn.ReadFrom(buf)
			if err != nil {
				d.log.Error("error reading from socket: %s", err)
				continue
			}

			if n == 0 {
				continue
			}

			peer := &peers.Peer{}
			if err := peer.Unmarshal(buf[:n]); err != nil {
				d.log.Error("can't unmarshal peer data: %s", err)
				continue
			}

			ip := src.(*net.UDPAddr).IP
			peer.Addrs.Add(ip)

			out <- peer

			d.log.Debug("found a peer at %s", ip)
		}
	}
}
