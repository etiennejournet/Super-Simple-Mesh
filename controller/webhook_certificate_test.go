package main

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"reflect"
	"testing"
)

func TestCreateSelfSignedCert(t *testing.T) {
	wh := &webHook{
		Name:      "my-test-certificate",
		Namespace: "my-test-namespace",
	}
	cert, key := createSelfSignedCert(wh)

	_, err := tls.X509KeyPair(cert, key)
	if err != nil {
		t.Fatal(err)
	}

	pemBlock, _ := pem.Decode(cert)
	if pemBlock == nil {
		t.Fatal("certificate is not a pem encoded string")
	}

	certificate, err := x509.ParseCertificate(pemBlock.Bytes)
	if err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(certificate.DNSNames, []string{wh.Name, wh.Name + "." + wh.Namespace + ".svc", wh.Name + "." + wh.Namespace + ".svc.cluster.local"}) {
		t.Fatal("Certificate not signed for the proper DNS Names")
	}

	if err = certificate.CheckSignatureFrom(certificate); err != nil {
		t.Fatal(err)
	}
}
