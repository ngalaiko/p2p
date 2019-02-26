package peers

import (
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"math/rand"
	"time"
)

//func initCert(peer *Peer, r *rand.Rand) (*depot.CertificateRevocationList, error) {
//formattedName := peer.ID

//key, err = pkix.CreateRSAKey(4096)
//if err != nil {
//return nil, err
//}

//crt, err := pkix.CreateCertificateAuthority(
//key,
//"",
//365*time.Day,
//"",
//"",
//"",
//"",
//peer.ID,
//)
//if err != nil {
//return nil, fmt.Errorf("error generating .crt: %s", err)
//}

//crl, err := pkix.CreateCertificateRevocationList(key, crt, expiresTime)
//if err != nil {
//return nil, fmt.Errorf("error generating .crl: %s", err)
//}
//}

//func requestCertFor(peer *Peer) error {
//name := peer.ID

//var formattedName = formatName(name)

//key, err := pkix.CreateRSAKey(4096)
//if err != nil {
//return fmt.Errorf("can't create .key: %s", err)
//}
//if len(passphrase) > 0 {
//fmt.Printf("Created %s/%s.key (encrypted by passphrase)\n", depotDir, formattedName)
//} else {
//fmt.Printf("Created %s/%s.key\n", depotDir, formattedName)
//}

//csr, err := pkix.CreateCertificateSigningRequest(key, c.String("organizational-unit"), ips, domains, uris, c.String("organization"), c.String("country"), c.String("province"), c.String("locality"), name)
//if err != nil {
//fmt.Fprintln(os.Stderr, "Create certificate request error:", err)
//os.Exit(1)
//} else {
//fmt.Printf("Created %s/%s.csr\n", depotDir, formattedName)
//}

//if c.Bool("stdout") {
//csrBytes, err := csr.Export()
//if err != nil {
//fmt.Fprintln(os.Stderr, "Print certificate request error:", err)
//os.Exit(1)
//} else {
//fmt.Printf(string(csrBytes))
//}
//}

//if err = depot.PutCertificateSigningRequest(d, formattedName, csr); err != nil {
//fmt.Fprintln(os.Stderr, "Save certificate request error:", err)
//}
//if len(passphrase) > 0 {
//if err = depot.PutEncryptedPrivateKey(d, formattedName, key, passphrase); err != nil {
//fmt.Fprintln(os.Stderr, "Save encrypted private key error:", err)
//}
//} else {
//if err = depot.PutPrivateKey(d, formattedName, key); err != nil {
//fmt.Fprintln(os.Stderr, "Save private key error:", err)
//}
//}
//}

func generateCertificate(p *Peer, r *rand.Rand) (tls.Certificate, error) {
	key, err := rsa.GenerateKey(r, 4096)
	if err != nil {
		return tls.Certificate{}, fmt.Errorf("private key cannot be created: %s", err)
	}

	tml := x509.Certificate{
		SerialNumber: big.NewInt(r.Int63()),
		Subject: pkix.Name{
			CommonName: p.ID,
		},
		Issuer: pkix.Name{
			CommonName: p.ID,
		},
		SignatureAlgorithm:    x509.SHA512WithRSA,
		PublicKeyAlgorithm:    x509.ECDSA,
		NotBefore:             time.Now(),
		NotAfter:              time.Now().AddDate(0, 0, 7),
		BasicConstraintsValid: true,
		IsCA:                  true,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
	}

	keyPem := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(key),
	})

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
