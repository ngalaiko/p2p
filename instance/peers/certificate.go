package peers

import (
	"crypto/tls"
	"fmt"
	"math/rand"
	"time"

	"github.com/square/certstrap/pkix"
)

func generateCertificate(peer *Peer, r *rand.Rand, keySize int, expires time.Time) ([]byte, *tls.Certificate, error) {
	key, err := pkix.CreateRSAKey(keySize)
	if err != nil {
		return nil, nil, err
	}

	crt, err := pkix.CreateCertificateAuthority(
		key,
		"",
		expires,
		"",
		"",
		"",
		"",
		peer.ID,
	)
	if err != nil {
		return nil, nil, fmt.Errorf("error generating .crt: %s", err)
	}

	tlsCert, err := makeTLSCertificate(key, crt)
	if err != nil {
		return nil, nil, fmt.Errorf("can't make TLS certificate: %s", err)
	}

	publicCrt, err := crt.Export()
	if err != nil {
		return nil, nil, fmt.Errorf("can't export public .crt: %s", err)
	}

	return publicCrt, tlsCert, nil
}

func makeTLSCertificate(key *pkix.Key, crt *pkix.Certificate) (*tls.Certificate, error) {
	keyBytes, err := key.ExportPrivate()
	if err != nil {
		return nil, fmt.Errorf("failed to export .key: %s", err)
	}

	crtBytes, err := crt.Export()
	if err != nil {
		return nil, fmt.Errorf("failed to export .crt: %s", err)
	}

	cert, err := tls.X509KeyPair(crtBytes, keyBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to make x509 key pair: %s", err)
	}

	return &cert, nil
}
