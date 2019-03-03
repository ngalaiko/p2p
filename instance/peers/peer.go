package peers

import (
	"crypto/tls"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/rand"
	"time"
)

const idLen = 8

// Peer is an instance of the same app.
type Peer struct {
	ID   string `json:"id"`
	Name string `json:"name"`

	Port         string `json:"port"`
	InsecurePort string `json:"insecure_port"`

	KnownPeers *peersList `json:"known_peers"`

	Addrs *addrsList `json:"-"`

	PublicCrt   []byte           `json:"-"`
	Certificate *tls.Certificate `json:"-"`
}

// New is a peer constructor.
func New(
	r *rand.Rand,
	port string,
	insecurePort string,
	keySize int,
) (*Peer, error) {
	idBytes := make([]byte, idLen)
	_, err := r.Read(idBytes)
	if err != nil {
		return nil, fmt.Errorf("error reading random bytes: %s", err)
	}

	p := &Peer{
		ID:           hex.EncodeToString(idBytes),
		Name:         newName(r),
		KnownPeers:   newPeersList(),
		Addrs:        newAddrsList(),
		Port:         port,
		InsecurePort: insecurePort,
	}

	p.PublicCrt, p.Certificate, err = generateCertificate(p, r, keySize, time.Now().AddDate(1, 0, 0))
	if err != nil {
		return nil, fmt.Errorf("can't generate CA certificate: %s", err)
	}

	return p, nil
}

// NewBlank returns new empty peer.
func NewBlank() *Peer {
	return &Peer{
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
	p.Addrs = newAddrsList()
	p.KnownPeers = newPeersList()
	return json.Unmarshal(data, p)
}
