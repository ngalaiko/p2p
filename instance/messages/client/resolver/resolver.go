package resolver

import (
	"fmt"
	"net"
	"sync"

	"github.com/ngalayko/p2p/instance/peers"
	"google.golang.org/grpc/resolver"
)

// Builder returns a resolver for a client addrs.
type Builder struct {
	port string

	guard      *sync.RWMutex
	addrsStore map[string][]string
}

// New returns a new resolver for the peer.
func New(port string) *Builder {
	return &Builder{
		port:       port,
		guard:      &sync.RWMutex{},
		addrsStore: map[string][]string{},
	}
}

// Add adds peer resolution rules.
func (b *Builder) Add(peer *peers.Peer) {
	list := peer.Addrs.List()
	addrs := make([]string, 0, len(list))
	for _, addr := range list {
		addrs = append(addrs, makeAddr(addr, b.port))
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
func (*Builder) Scheme() string { return "client" }

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
