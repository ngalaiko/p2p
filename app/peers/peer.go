package peers

import (
	"encoding/hex"
	"encoding/json"
	"math/rand"
	"time"
)

const idLen = 32

var r = rand.New(rand.NewSource(time.Now().Unix()))

// Peer is an instance of the same app.
type Peer struct {
	ID         string     `json:"id"`
	Name       string     `json:"name"`
	KnownPeers *peersList `json:"known_peers"`
	Addrs      *addrsList `json:"addrs"`
}

// New is a peer constructor.
func New() *Peer {
	idBytes := make([]byte, idLen)
	_, _ = r.Read(idBytes)
	return &Peer{
		ID:         hex.EncodeToString(idBytes),
		Name:       newName(r),
		KnownPeers: newPeersList(),
		Addrs:      newAddrsList(),
	}
}

// Marshal is a marshal function to use when sending peer info over a network.
func (p *Peer) Marshal() ([]byte, error) {
	return json.Marshal(p)
}

// Unmarshal is a marshal function to use when sending peer info over a network.
func (p *Peer) Unmarshal(data []byte) error {
	return json.Unmarshal(data, p)
}
