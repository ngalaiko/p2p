package peers

import "net"

type addrsList struct {
	list []net.Addr
	byIP map[string]net.Addr
}

func newAddrsList() *addrsList {
	return &addrsList{
		byIP: map[string]net.Addr{},
	}
}

// Add adds a new address to the peer.
func (a *addrsList) Add(addr net.Addr) bool {
	if _, known := a.byIP[addr.String()]; known {
		return false
	}

	a.byIP[addr.String()] = addr
	a.list = append(a.list, addr)
	return true
}

// List returns known addres list.
func (a *addrsList) List() []net.Addr {
	return a.list
}
