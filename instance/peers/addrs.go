package peers

import (
	"net"
	"sync"
)

type addrsList struct {
	guard *sync.RWMutex
	list  []net.IP
	byIP  map[string]net.IP
}

func newAddrsList() *addrsList {
	return &addrsList{
		guard: &sync.RWMutex{},
		byIP:  map[string]net.IP{},
	}
}

// Add adds a new address to the peer.
func (a *addrsList) Add(addr net.IP) bool {
	a.guard.Lock()
	defer a.guard.Unlock()

	if _, known := a.byIP[addr.String()]; known {
		return false
	}

	a.byIP[addr.String()] = addr
	a.list = append(a.list, addr)
	return true
}

// List returns known addres list.
func (a *addrsList) List() []net.IP {
	a.guard.RLock()
	defer a.guard.RUnlock()
	return a.list
}
