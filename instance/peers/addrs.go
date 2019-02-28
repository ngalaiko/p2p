package peers

import (
	"net"
	"sync"
)

type addrsList struct {
	guard *sync.RWMutex
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
	return true
}

// Map returns map of ips.
func (a *addrsList) Map() map[string]net.IP {
	a.guard.RLock()
	defer a.guard.RUnlock()
	return a.byIP
}
