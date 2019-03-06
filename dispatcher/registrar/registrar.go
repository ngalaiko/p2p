package registrar

import (
	"net/url"

	"github.com/ngalayko/p2p/instance/peers"
)

// Registrar register new redirection rules.
type Registrar interface {
	Register(peer *peers.Peer, target *url.URL, host string) error
}
