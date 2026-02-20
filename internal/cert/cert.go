package cert

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"os"
	"path/filepath"
	"time"
)

func certsDir() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".port0", "certs")
}

func Generate(name string) (certPEM, keyPEM []byte, err error) {
	priv, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, nil, fmt.Errorf("cert: generate key: %w", err)
	}

	serial, err := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))
	if err != nil {
		return nil, nil, fmt.Errorf("cert: generate serial: %w", err)
	}

	now := time.Now()
	template := &x509.Certificate{
		SerialNumber: serial,
		Subject: pkix.Name{
			CommonName: name + ".localhost",
		},
		DNSNames: []string{
			name + ".localhost",
			name + ".web",
			name + ".local",
		},
		NotBefore:             now,
		NotAfter:              now.Add(365 * 24 * time.Hour),
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
		IsCA:                  true,
	}

	certDER, err := x509.CreateCertificate(rand.Reader, template, template, &priv.PublicKey, priv)
	if err != nil {
		return nil, nil, fmt.Errorf("cert: create certificate: %w", err)
	}

	certPEM = pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certDER})

	keyDER, err := x509.MarshalECPrivateKey(priv)
	if err != nil {
		return nil, nil, fmt.Errorf("cert: marshal key: %w", err)
	}
	keyPEM = pem.EncodeToMemory(&pem.Block{Type: "EC PRIVATE KEY", Bytes: keyDER})

	dir := certsDir()
	os.MkdirAll(dir, 0755)

	certPath := filepath.Join(dir, name+".pem")
	keyPath := filepath.Join(dir, name+"-key.pem")

	if err := os.WriteFile(certPath, certPEM, 0644); err != nil {
		return nil, nil, fmt.Errorf("cert: write cert: %w", err)
	}
	if err := os.WriteFile(keyPath, keyPEM, 0600); err != nil {
		return nil, nil, fmt.Errorf("cert: write key: %w", err)
	}

	return certPEM, keyPEM, nil
}

func Load(name string) (tls.Certificate, error) {
	dir := certsDir()
	certPath := filepath.Join(dir, name+".pem")
	keyPath := filepath.Join(dir, name+"-key.pem")
	cert, err := tls.LoadX509KeyPair(certPath, keyPath)
	if err != nil {
		return tls.Certificate{}, fmt.Errorf("cert: load %s: %w", name, err)
	}
	return cert, nil
}

func Exists(name string) bool {
	dir := certsDir()
	certPath := filepath.Join(dir, name+".pem")
	_, err := os.Stat(certPath)
	return err == nil
}

func EnsureGenerated(name string) error {
	if Exists(name) {
		return nil
	}
	_, _, err := Generate(name)
	return err
}
