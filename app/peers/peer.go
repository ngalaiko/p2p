package peers

import (
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/hex"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"math/big"
	"math/rand"
	"time"
)

const idLen = 32

// Peer is an instance of the same app.
type Peer struct {
	ID          string          `json:"id"`
	Name        string          `json:"name"`
	KnownPeers  *peersList      `json:"known_peers"`
	Addrs       *addrsList      `json:"-"`
	Certificate tls.Certificate `json:"-"`
}

// New is a peer constructor.
func New() (*Peer, error) {
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
	}
	p.Certificate, err = generateCertificate(p, r)
	if err != nil {
		return nil, fmt.Errorf("error generating certificate: %s", err)
	}
	return p, nil
}

func generateCertificate(p *Peer, r *rand.Rand) (tls.Certificate, error) {
	key, err := rsa.GenerateKey(r, 2048)
	if err != nil {
		return tls.Certificate{}, fmt.Errorf("private key cannot be created: %s", err)
	}

	keyPem := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(key),
	})

	tml := x509.Certificate{
		NotBefore:    time.Now(),
		NotAfter:     time.Now().AddDate(5, 0, 0),
		SerialNumber: big.NewInt(r.Int63()),
		Subject: pkix.Name{
			CommonName: p.Name,
		},
		BasicConstraintsValid: true,
	}
	cert, err := x509.CreateCertificate(r, &tml, &tml, &key.PublicKey, key)
	if err != nil {
		return tls.Certificate{}, fmt.Errorf("certificate cannot be created:%s", err)
	}

	certPem := pem.EncodeToMemory(&pem.Block{
		Type:  "CERTIFICATE",
		Bytes: cert,
	})

	return tls.X509KeyPair(certPem, keyPem)
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
