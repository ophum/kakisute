package main

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"log/slog"
	"math/big"
	"os"
	"time"

	"github.com/go-faster/errors"
)

func main() {
	if err := run(); err != nil {
		slog.Error("failed to generate client certificate", "error", err)
		os.Exit(1)
	}
}

func run() error {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return errors.Wrap(err, "failed to generate key")
	}

	keyUsage := x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment
	notBefore := time.Now()
	notAfter := notBefore.Add(time.Hour)

	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)
	if err != nil {
		return errors.Wrap(err, "failed to generate serial numbner")
	}

	template := x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			Organization: []string{"Example"},
		},
		NotBefore:             notBefore,
		NotAfter:              notAfter,
		KeyUsage:              keyUsage,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
		BasicConstraintsValid: true,
		DNSNames:              []string{"localhost"},
	}

	caCert, err := tls.LoadX509KeyPair("cert.pem", "key.pem")
	if err != nil {
		return errors.Wrap(err, "failed to load x509 key pair")
	}
	derBytes, err := x509.CreateCertificate(rand.Reader, &template, caCert.Leaf, &privateKey.PublicKey, caCert.PrivateKey)
	if err != nil {
		return errors.Wrap(err, "failed to create certificate")
	}

	certOut, err := os.Create("client-cert.pem")
	if err != nil {
		return errors.Wrap(err, "failed to create client-cert.pem")
	}
	defer certOut.Close()
	if err := pem.Encode(certOut, &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: derBytes,
	}); err != nil {
		return errors.Wrap(err, "failed to encode certificate pem")
	}

	keyOut, err := os.Create("client-key.pem")
	if err != nil {
		return errors.Wrap(err, "failed to create client-key.pem")
	}
	defer keyOut.Close()
	privBytes, err := x509.MarshalPKCS8PrivateKey(privateKey)
	if err != nil {
		return errors.Wrap(err, "failed to MarshalPKCS8PrivateKey")
	}
	if err := pem.Encode(keyOut, &pem.Block{
		Type:  "PRIVATE KEY",
		Bytes: privBytes,
	}); err != nil {
		return errors.Wrap(err, "failed to encode private key pem")
	}

	return nil
}
