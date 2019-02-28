package resolver

import (
	"fmt"
	"net"
	"sync"

	"google.golang.org/grpc/resolver"

	"github.com/ngalayko/p2p/instance/peers"
)

// Builder returns a resolver for a client addrs.
type Builder struct {
	secure bool

	guard      *sync.RWMutex
	addrsStore map[string][]string
}

// New returns a new resolver for the peer.
func New(secure bool) *Builder {
	return &Builder{
		guard:      &sync.RWMutex{},
		addrsStore: map[string][]string{},
		secure:     secure,
	}
}

// Add adds peer resolution rules.
func (b *Builder) Add(peer *peers.Peer) {
	m := peer.Addrs.Map()
	addrs := make([]string, 0, len(m))
	for _, addr := range m {
		if b.secure {
			addrs = append(addrs, makeAddr(addr, peer.Port))
		} else {
			addrs = append(addrs, makeAddr(addr, peer.InsecurePort))
		}
	}

	b.guard.Lock()
	b.addrsStore[peer.ID] = addrs
	b.guard.Unlock()
}

func makeAddr(ip net.IP, port string) string {
	if ip.To4() == nil {
		return fmt.Sprintf("[%s]:%s", ip.String(), port)
	}
	return fmt.Sprintf("%s:%s", ip.String(), port)
}

// Build implements grpc.Builder.
func (b *Builder) Build(target resolver.Target, cc resolver.ClientConn, opts resolver.BuildOption) (resolver.Resolver, error) {
	b.guard.RLock()
	defer b.guard.RUnlock()

	r := &Resolver{
		target:     target,
		cc:         cc,
		addrsStore: b.addrsStore,
	}
	r.start()
	return r, nil
}

// Scheme implements grpc.Builder.
func (b *Builder) Scheme() string {
	if b.secure {
		return "peer"
	}
	return "greet"
}

// Resolver implements grpc.Resolver.
type Resolver struct {
	target     resolver.Target
	cc         resolver.ClientConn
	addrsStore map[string][]string
}

func (r *Resolver) start() {
	addrStrs := r.addrsStore[r.target.Endpoint]
	addrs := make([]resolver.Address, len(addrStrs))
	for i, s := range addrStrs {
		addrs[i] = resolver.Address{Addr: s}
	}
	r.cc.NewAddress(addrs)
}

// ResolveNow implements grpc.Resolver.
func (*Resolver) ResolveNow(o resolver.ResolveNowOption) {}

// Close implements grpc.Resolver.
func (*Resolver) Close() {}
