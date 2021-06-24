package main

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	admissionRegistration "k8s.io/api/admissionregistration/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
	"os"
	"reflect"
	"testing"
)

func TestCreateSelfSignedCert(t *testing.T) {
	wh := &webHook{
		Name:      "my-test-webhook",
		Namespace: "my-test-namespace",
	}
	cert, key, err := createSelfSignedCert(wh)
	if err != nil {
		t.Fatal(err)
	}

	_, err = tls.X509KeyPair(cert, key)
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

func TestCAInjection(t *testing.T) {
	wh := &webHook{
		Name:      "my-test-webhook",
		Namespace: "my-test-namespace",
	}
	clientSet := fake.NewSimpleClientset()
	cert, _, err := createSelfSignedCert(wh)
	if err != nil {
		t.Fatal(err)
	}
	err = injectCAInMutatingWebhook(clientSet, wh.Name, cert)
	if err == nil {
		t.Fatal("No error raised but mutating webhook shouldn't have been found")
	}

	mutatingWebHookConfiguration := &admissionRegistration.MutatingWebhookConfiguration{
		ObjectMeta: metav1.ObjectMeta{Name: wh.Name},
		Webhooks: []admissionRegistration.MutatingWebhook{{
			Name: wh.Name,
		}},
	}

	_, err = clientSet.AdmissionregistrationV1().MutatingWebhookConfigurations().Create(context.TODO(), mutatingWebHookConfiguration, metav1.CreateOptions{})
	if err != nil {
		t.Fatal(err)
	}
	injectCAInMutatingWebhook(clientSet, wh.Name, cert)

	result, _ := clientSet.AdmissionregistrationV1().MutatingWebhookConfigurations().Get(context.TODO(), mutatingWebHookConfiguration.ObjectMeta.Name, metav1.GetOptions{})
	if !reflect.DeepEqual(result.Webhooks[0].ClientConfig.CABundle, cert) {
		t.Fatal("Certificate not properly injected")
	}
}

func TestWriteCertsToHomeFolder(t *testing.T) {
	certPath, keyPath := writeCertsToHomeFolder([]byte{}, []byte{})
	defer os.Remove(certPath)
	defer os.Remove(keyPath)
	if _, err := os.Stat(certPath); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(keyPath); err != nil {
		t.Fatal(err)
	}
}
