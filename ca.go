package main

// http://golang.org/src/pkg/crypto/tls/generate_cert.go

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"log"
	"math/big"
	"os"
	"strings"
	"time"
)

type CA struct {
	orgName string
	rsaBits int
}

func NewCA(orgName string, rsaBits int) *CA {
	return &CA{orgName, rsaBits}
}

func (c *CA) Issue(isCA bool, host string, vaildFor time.Duration) {
	priv, err := rsa.GenerateKey(rand.Reader, c.rsaBits)
	if err != nil {
		log.Fatalf("failed to generate private key: %s", err)
	}

	notBefore := time.Now().Add(-time.Duration(time.Hour))
	notAfter := time.Now().Add(vaildFor)

	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)
	if err != nil {
		log.Fatalf("failed to generate serial number: %s", err)
	}

	template := x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			Organization: []string{c.orgName},
		},
		NotBefore: notBefore,
		NotAfter:  notAfter,

		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
	}

	hosts := strings.Split(host, ",")
	for _, h := range hosts {
		template.DNSNames = append(template.DNSNames, h)
	}

	if isCA {
		template.IsCA = true
		template.KeyUsage |= x509.KeyUsageCertSign
	}

	derBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, &priv.PublicKey, priv)
	if err != nil {
		log.Fatalf("Failed to create certificate: %s", err)
	}

	outFile, err := os.Create("cert.crt")
	defer outFile.Close()
	if err != nil {
		log.Fatalf("failed to open cert.crt for writing: %s", err)
	}
	pem.Encode(outFile, &pem.Block{Type: "CERTIFICATE", Bytes: derBytes})
	pem.Encode(outFile, &pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(priv)})
}
