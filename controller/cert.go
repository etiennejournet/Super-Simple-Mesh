package main

import (
	"encoding/json"
	types "k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
  "os"
	"context"
	"encoding/base64"
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"encoding/pem"
	"log"
	"math/big"
	"time"
)

func (wh *webHook) createSelfSignedCert() ([]byte, []byte) {
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
		DNSNames:              []string{wh.Name, wh.Name + "." + wh.Namespace + ".svc",  wh.Name + "." + wh.Namespace + ".svc.cluster.local"},
	}

	derBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, &privateKey.PublicKey, privateKey)
	cert := &bytes.Buffer{}
	pem.Encode(cert, &pem.Block{Type: "CERTIFICATE", Bytes: derBytes})

	if err != nil {
		log.Fatalf("Failed to create certificate: %s", err)
	}

	return cert.Bytes(), pemPrivateKey
}

func injectCAInMutatingWebhook(wh *webHook, ca []byte) {
	var hashedCA = make([]byte, base64.StdEncoding.EncodedLen(len(ca)))
	clientSet, err := kubernetes.NewForConfig(wh.Client)
	if err != nil {
		log.Print(err)
	}

	_, err = clientSet.AdmissionregistrationV1().MutatingWebhookConfigurations().Get(context.TODO(), wh.Name, metav1.GetOptions{})
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}

	base64.StdEncoding.Encode(hashedCA, ca)
	newCA := []patchValue{{
		Op:    "replace",
		Path:  "/webhooks/0/clientConfig/caBundle",
		Value: string(hashedCA),
	  },
	}
	newCAByte, _ := json.Marshal(newCA)

	log.Print("Installing new certificate in ", wh.Name, " mutating webhook configuration")
	_, err = clientSet.AdmissionregistrationV1().MutatingWebhookConfigurations().Patch(context.TODO(), wh.Name, types.JSONPatchType, newCAByte, metav1.PatchOptions{})
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}
}

//func watchMutatingWebHookObject(wh *WebHook) {}


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
