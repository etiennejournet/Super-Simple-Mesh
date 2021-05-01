package main

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"log"
	"math/big"
	"time"
)

func (wh *webHook) createCert() ([]byte, []byte) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	pemPrivateKey := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
	})

	if err != nil {
		log.Fatal(err)
	}

	template := x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			CommonName: wh.Name,
		},
		NotBefore: time.Now(),
		NotAfter:  time.Now().Add(time.Hour * 24 * 180),

		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
		DNSNames:              []string{wh.Name, wh.Name + ".default.svc"},
	}

	derBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, &privateKey.PublicKey, privateKey)
	cert := &bytes.Buffer{}
	pem.Encode(cert, &pem.Block{Type: "CERTIFICATE", Bytes: derBytes})

	if err != nil {
		log.Fatalf("Failed to create certificate: %s", err)
	}

	return cert.Bytes(), pemPrivateKey
}
