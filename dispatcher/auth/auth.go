package auth

import (
	"net/http"

	"github.com/ngalayko/p2p/instance/peers"
)

// Authorizer returns peer information from http request.
type Authorizer interface {
	Get(*http.Request) (*peers.Peer, error)
	Store(http.ResponseWriter, *peers.Peer) error
}
