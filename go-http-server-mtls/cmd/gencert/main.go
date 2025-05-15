package main

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"net"
	"os"
	"time"
)

func main() {
	if err := run(); err != nil {
		panic(err)
	}
}

func run() error {
	ca := &x509.Certificate{
		SerialNumber: big.NewInt(0),
		Subject: pkix.Name{
			Organization:  []string{"Test"},
			Country:       []string{"JP"},
			Province:      []string{""},
			Locality:      []string{"Osaka"},
			StreetAddress: []string{""},
			PostalCode:    []string{""},
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().AddDate(10, 0, 0),
		IsCA:                  true,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		BasicConstraintsValid: true,
	}
	caPrivKey, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		return nil
	}

	caBytes, err := x509.CreateCertificate(rand.Reader, ca, ca, &caPrivKey.PublicKey, caPrivKey)
	if err != nil {
		return err
	}

	caFile, err := os.Create("ca.crt")
	if err != nil {
		return err
	}
	defer caFile.Close()

	if err := pem.Encode(caFile, &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: caBytes,
	}); err != nil {
		return err
	}

	privKeyFile, err := os.Create("ca.key")
	if err != nil {
		return err
	}
	defer privKeyFile.Close()

	if err := pem.Encode(privKeyFile, &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(caPrivKey),
	}); err != nil {
		return err
	}

	//
	serverCert := &x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			Organization:  []string{"Test"},
			Country:       []string{"JP"},
			Province:      []string{""},
			Locality:      []string{"Osaka"},
			StreetAddress: []string{""},
			PostalCode:    []string{""},
			CommonName:    "localhost",
		},

		IPAddresses: []net.IP{net.IPv4(127, 0, 0, 1)},
		DNSNames:    []string{"localhost"},

		NotBefore:   time.Now(),
		NotAfter:    time.Now().Add(time.Hour),
		ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		KeyUsage:    x509.KeyUsageDigitalSignature,
	}

	serverPrivKey, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		return err
	}

	serverCertBytes, err := x509.CreateCertificate(rand.Reader, serverCert, ca, &serverPrivKey.PublicKey, caPrivKey)
	if err != nil {
		return err
	}

	serverCertFile, err := os.Create("server.crt")
	if err != nil {
		return err
	}
	defer serverCertFile.Close()

	if err := pem.Encode(serverCertFile, &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: serverCertBytes,
	}); err != nil {
		return err
	}

	serverPrivKeyFile, err := os.Create("server.key")
	if err != nil {
		return err
	}
	defer serverPrivKeyFile.Close()

	if err := pem.Encode(serverPrivKeyFile, &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(serverPrivKey),
	}); err != nil {
		return err
	}
	//
	clientCert := &x509.Certificate{
		SerialNumber: big.NewInt(2),
		Subject: pkix.Name{
			Organization:  []string{"Test"},
			Country:       []string{"JP"},
			Province:      []string{""},
			Locality:      []string{"Osaka"},
			StreetAddress: []string{""},
			PostalCode:    []string{""},
			CommonName:    "test-client",
		},
		IPAddresses: []net.IP{net.IPv4(127, 0, 0, 1)},
		NotBefore:   time.Now(),
		NotAfter:    time.Now().Add(time.Hour),
		ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
		KeyUsage:    x509.KeyUsageDigitalSignature,
	}

	clientPrivKey, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		return err
	}

	clientCertBytes, err := x509.CreateCertificate(rand.Reader, clientCert, ca, &clientPrivKey.PublicKey, caPrivKey)
	if err != nil {
		return err
	}

	clientCertFile, err := os.Create("client.crt")
	if err != nil {
		return err
	}
	defer clientCertFile.Close()

	if err := pem.Encode(clientCertFile, &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: clientCertBytes,
	}); err != nil {
		return err
	}

	clientPrivKeyFile, err := os.Create("client.key")
	if err != nil {
		return err
	}
	defer clientPrivKeyFile.Close()

	if err := pem.Encode(clientPrivKeyFile, &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(clientPrivKey),
	}); err != nil {
		return err
	}
	return nil
}
