package main

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	types "k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
	"math/big"
	"os"
	"time"
)

func createSelfSignedCert(wh *webHook) ([]byte, []byte) {
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

		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
		IsCA:                  true,
		DNSNames:              []string{wh.Name, wh.Name + "." + wh.Namespace + ".svc", wh.Name + "." + wh.Namespace + ".svc.cluster.local"},
	}

	derBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, &privateKey.PublicKey, privateKey)
	cert := &bytes.Buffer{}
	pem.Encode(cert, &pem.Block{Type: "CERTIFICATE", Bytes: derBytes})

	if err != nil {
		log.Fatal(err)
	}

	return cert.Bytes(), pemPrivateKey
}

func injectCAInMutatingWebhook(clientSet kubernetes.Interface, webHookName string, ca []byte) {
	var hashedCA = make([]byte, base64.StdEncoding.EncodedLen(len(ca)))

	_, err := clientSet.AdmissionregistrationV1().MutatingWebhookConfigurations().Get(context.TODO(), webHookName, metav1.GetOptions{})
	if err != nil {
		log.Fatal(err)
	}

	base64.StdEncoding.Encode(hashedCA, ca)
	newCA := []patchValue{{
		Op:    "replace",
		Path:  "/webhooks/0/clientConfig/caBundle",
		Value: string(hashedCA),
	},
	}
	newCAByte, _ := json.Marshal(newCA)

	InfoLogger.Print("Installing new certificate in ", webHookName, " mutating webhook configuration")
	_, err = clientSet.AdmissionregistrationV1().MutatingWebhookConfigurations().Patch(context.TODO(), webHookName, types.JSONPatchType, newCAByte, metav1.PatchOptions{})
	if err != nil {
		log.Fatal(err)
	}
}

func writeCertsToHomeFolder(cert []byte, key []byte) (string, string) {
	userHomeDir, err := os.UserHomeDir()
	if err != nil {
		log.Fatal(err)
	}

	certPath := userHomeDir + "/tls.crt"
	keyPath := userHomeDir + "/tls.key"

	err = ioutil.WriteFile(certPath, cert, 0644)
	if err != nil {
		log.Fatal(err)
	}
	err = ioutil.WriteFile(keyPath, key, 0644)
	if err != nil {
		log.Fatal(err)
	}

	return certPath, keyPath
}

// This is not usable for now because no Kubernetes Signer permits creating server certs
//func createCAUsingKubernetesAPI(clientSet *kubernetes.Clientset) ([]byte, []byte) {
//     var keyUsage = []certificates.KeyUsage{"server auth"}
//     var certName = "mutating-webhook-secret"
//
//     log.Print("First, let's check for existing CSR")
//     csr, err := clientSet.CertificatesV1().CertificateSigningRequests().Get(context.TODO(), certName, v1.GetOptions{})
//     if err != nil {
//             log.Print(err)
//             os.Exit(1)
//     }
//
//     if csr != nil {
//             log.Print("Legacy CSR was found, deleting it")
//             err = clientSet.CertificatesV1().CertificateSigningRequests().Delete(context.TODO(), certName, v1.DeleteOptions{})
//     }
//
//     privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
//     pemPrivateKey := pem.EncodeToMemory(&pem.Block{
//             Type:  "RSA PRIVATE KEY",
//             Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
//     })
//
//     if err != nil {
//             log.Print("Cannot generate RSA key\n")
//             os.Exit(1)
//     }
//
//     subject := pkix.Name{
//             CommonName: "autosidecar-hook",
//     }
//
//     asn1, err := asn1.Marshal(subject.ToRDNSequence())
//     csrReq := x509.CertificateRequest{
//             RawSubject:         asn1,
//             SignatureAlgorithm: x509.SHA256WithRSA,
//    DNSNames: [
//     }
//     bytes, err := x509.CreateCertificateRequest(rand.Reader, &csrReq, privateKey)
//     if err != nil {
//             fmt.Println(err)
//             os.Exit(1)
//     }
//
//     csr = &certificates.CertificateSigningRequest{
//             ObjectMeta: v1.ObjectMeta{
//                     Name: certName,
//             },
//             Spec: certificates.CertificateSigningRequestSpec{
//                     Groups: []string{
//                             "system:authenticated",
//                     },
//                     Request:    pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE REQUEST", Bytes: bytes}),
//                     Usages:     keyUsage,
//                     SignerName: "kubernetes.io/kube-apiserver-client",
//             },
//     }
//
//     log.Print("Creating New CSR")
//     csr, err = clientSet.CertificatesV1().CertificateSigningRequests().Create(context.TODO(), csr, v1.CreateOptions{})
//     if err != nil {
//             fmt.Println(err)
//             os.Exit(1)
//     }
//
//     csr, err = clientSet.CertificatesV1().CertificateSigningRequests().Get(context.TODO(), certName, v1.GetOptions{})
//     if err != nil {
//             fmt.Println(err)
//             os.Exit(1)
//     }
//
//     csr.Status.Conditions = append(csr.Status.Conditions, certificates.CertificateSigningRequestCondition{
//             Type:           certificates.CertificateApproved,
//             Status:         "True",
//             Message:        "This CSR was approved",
//             LastUpdateTime: v1.Now(),
//     })
//
//     log.Print("Validating CSR")
//     csr, err = clientSet.CertificatesV1().CertificateSigningRequests().UpdateApproval(context.TODO(), certName, csr, v1.UpdateOptions{})
//     if err != nil {
//             fmt.Println(err)
//             os.Exit(1)
//     }
//
//     csr, err = clientSet.CertificatesV1().CertificateSigningRequests().Get(context.TODO(), certName, v1.GetOptions{})
//
//     return pemPrivateKey, csr.Status.Certificate
//}
