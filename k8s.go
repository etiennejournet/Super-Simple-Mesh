package main

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"flag"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	types "k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"log"
	"os"
)

func kubClient() *rest.Config {
	kubeconfig := flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	flag.Parse()

	if *kubeconfig == "" {
		log.Print("Detected in cluster launch")
		// creates the in-cluster config
		config, err := rest.InClusterConfig()
		if err != nil {
			panic(err.Error())
		}
		return config
	}

	// use the current context in kubeconfig
	log.Print("Detected Kubeconfig flag")
	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		panic(err.Error())
	}
	return config
}

//func (wh *webHook) alterMutatingWebhook(clientSet *kubernetes.Clientset, ca []byte) {
func (wh *webHook) alterMutatingWebhook(ca []byte) {
	var hashedCA = make([]byte, base64.StdEncoding.EncodedLen(len(ca)))
	// creates the clientset
	clientSet, err := kubernetes.NewForConfig(wh.Client)
	if err != nil {
		panic(err.Error())
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

	log.Print("Upgrading new certificate in ", wh.Name, " webhook")
	_, err = clientSet.AdmissionregistrationV1().MutatingWebhookConfigurations().Patch(context.TODO(), wh.Name, types.JSONPatchType, newCAByte, metav1.PatchOptions{})
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}

}

// This is not usable for now because no Kubernetes Signer permits creating server certs
//func getKubernetesCert(clientSet *kubernetes.Clientset) ([]byte, []byte) {
//	var keyUsage = []certificates.KeyUsage{"server auth"}
//	var certName = "mutating-webhook-secret"
//
//	log.Print("First, let's check for existing CSR")
//	csr, err := clientSet.CertificatesV1().CertificateSigningRequests().Get(context.TODO(), certName, v1.GetOptions{})
//	if err != nil {
//		log.Print(err)
//		os.Exit(1)
//	}
//
//	if csr != nil {
//		log.Print("Legacy CSR was found, deleting it")
//		err = clientSet.CertificatesV1().CertificateSigningRequests().Delete(context.TODO(), certName, v1.DeleteOptions{})
//	}
//
//	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
//	pemPrivateKey := pem.EncodeToMemory(&pem.Block{
//		Type:  "RSA PRIVATE KEY",
//		Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
//	})
//
//	if err != nil {
//		log.Print("Cannot generate RSA key\n")
//		os.Exit(1)
//	}
//
//	subject := pkix.Name{
//		CommonName: "autosidecar-hook",
//	}
//
//	asn1, err := asn1.Marshal(subject.ToRDNSequence())
//	csrReq := x509.CertificateRequest{
//		RawSubject:         asn1,
//		SignatureAlgorithm: x509.SHA256WithRSA,
//    DNSNames: [
//	}
//	bytes, err := x509.CreateCertificateRequest(rand.Reader, &csrReq, privateKey)
//	if err != nil {
//		fmt.Println(err)
//		os.Exit(1)
//	}
//
//	csr = &certificates.CertificateSigningRequest{
//		ObjectMeta: v1.ObjectMeta{
//			Name: certName,
//		},
//		Spec: certificates.CertificateSigningRequestSpec{
//			Groups: []string{
//				"system:authenticated",
//			},
//			Request:    pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE REQUEST", Bytes: bytes}),
//			Usages:     keyUsage,
//			SignerName: "kubernetes.io/kube-apiserver-client",
//		},
//	}
//
//	log.Print("Creating New CSR")
//	csr, err = clientSet.CertificatesV1().CertificateSigningRequests().Create(context.TODO(), csr, v1.CreateOptions{})
//	if err != nil {
//		fmt.Println(err)
//		os.Exit(1)
//	}
//
//	csr, err = clientSet.CertificatesV1().CertificateSigningRequests().Get(context.TODO(), certName, v1.GetOptions{})
//	if err != nil {
//		fmt.Println(err)
//		os.Exit(1)
//	}
//
//	csr.Status.Conditions = append(csr.Status.Conditions, certificates.CertificateSigningRequestCondition{
//		Type:           certificates.CertificateApproved,
//		Status:         "True",
//		Message:        "This CSR was approved",
//		LastUpdateTime: v1.Now(),
//	})
//
//	log.Print("Validating CSR")
//	csr, err = clientSet.CertificatesV1().CertificateSigningRequests().UpdateApproval(context.TODO(), certName, csr, v1.UpdateOptions{})
//	if err != nil {
//		fmt.Println(err)
//		os.Exit(1)
//	}
//
//	csr, err = clientSet.CertificatesV1().CertificateSigningRequests().Get(context.TODO(), certName, v1.GetOptions{})
//
//	return pemPrivateKey, csr.Status.Certificate
//}
