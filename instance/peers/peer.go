package peers

import (
	"crypto/tls"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/rand"
	"time"
)

const idLen = 32

// Peer is an instance of the same app.
type Peer struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Port string `json:"port"`

	KnownPeers *peersList `json:"-"`

	Addrs *addrsList `json:"-"`

	PublicCrt   []byte           `json:"public_key"`
	Certificate *tls.Certificate `json:"-"`
}

// New is a peer constructor.
func New(port string) (*Peer, error) {
	r := rand.New(rand.NewSource(time.Now().Unix()))

	idBytes := make([]byte, idLen)
	_, err := r.Read(idBytes)
	if err != nil {
		return nil, fmt.Errorf("error reading random bytes: %s", err)
	}

	p := &Peer{
		ID:         hex.EncodeToString(idBytes),
		Name:       newName(r),
		KnownPeers: newPeersList(),
		Addrs:      newAddrsList(),
		Port:       port,
	}

	p.PublicCrt, p.Certificate, err = generateCertificate(p, r, 4096, time.Now().AddDate(1, 0, 0))
	if err != nil {
		return nil, fmt.Errorf("can't generate CA certificate: %s", err)
	}

	return p, nil
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
