package main

import (
	"context"
	certmanager "github.com/jetstack/cert-manager/pkg/apis/certmanager/v1"
	certManagerClient "github.com/jetstack/cert-manager/pkg/client/clientset/versioned"
	certManagerTesting "github.com/jetstack/cert-manager/pkg/client/clientset/versioned/fake"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/rest"
	"testing"
)

type webHookTest struct {
	webHookInterface
}

func (wh *webHookTest) createKubernetesClientSet() (kubernetes.Interface, error) {
	return fake.NewSimpleClientset(), nil
}

func (wh *webHookTest) createCertManagerClientSet() (certManagerClient.Interface, error) {
	clientSet := certManagerTesting.NewSimpleClientset()
	readyClusterIssuer := &certmanager.ClusterIssuer{
		ObjectMeta: metav1.ObjectMeta{Name: "ca-issuer"},
		Spec: certmanager.IssuerSpec{
			IssuerConfig: certmanager.IssuerConfig{},
		},
		Status: certmanager.IssuerStatus{
			Conditions: []certmanager.IssuerCondition{
				{Type: "Ready", Status: "True"},
			},
		},
	}
	_, err := clientSet.CertmanagerV1().ClusterIssuers().Create(context.TODO(), readyClusterIssuer, metav1.CreateOptions{})
	if err != nil {
		return clientSet, err
	}

	return clientSet, nil
}

func TestNewCertManagerMutationConfig(t *testing.T) {
	whTest := &webHookTest{
		&webHook{
			Name:       "my-test-webhook",
			EnvoyUID:   777,
			RestConfig: &rest.Config{},
		},
	}
	podTemplateSpec := corev1.PodTemplateSpec{
		ObjectMeta: metav1.ObjectMeta{Name: "test-pod-template"},
	}

	_, err := newCertManagerMutationConfig(whTest, "my-test-object", "my-test-namespace", podTemplateSpec)
	if err != nil {
		t.Fatal(err)
	}
}

func TestNewCreateCertificateRequest(t *testing.T) {
	whTest := &webHookTest{
		&webHook{
			Name:       "my-test-webhook",
			EnvoyUID:   777,
			RestConfig: &rest.Config{},
		},
	}
	podTemplateSpec := corev1.PodTemplateSpec{
		ObjectMeta: metav1.ObjectMeta{Name: "test-pod-template"},
	}

	mutationConfig, err := newCertManagerMutationConfig(whTest, "my-test-object-not-existing-cert", "my-test-namespace", podTemplateSpec)
	if err != nil {
		t.Fatal(err)
	}
	err = mutationConfig.createCertificateRequest()
	if err != nil {
		t.Fatal(err)
	}
}

func TestCheckClusterIssuerExistsAndReady(t *testing.T) {
	clientSet := certManagerTesting.NewSimpleClientset()
	err := checkClusterIssuerExistsAndReady(clientSet, "test")
	if err == nil {
		t.Fatal("Found absent cluster issuer")
	}

	notReadyClusterIssuer := &certmanager.ClusterIssuer{
		ObjectMeta: metav1.ObjectMeta{Name: "test"},
		Spec: certmanager.IssuerSpec{
			IssuerConfig: certmanager.IssuerConfig{},
		},
		Status: certmanager.IssuerStatus{
			Conditions: []certmanager.IssuerCondition{
				{Type: "Ready", Status: "False"},
			},
		},
	}
	clientSet.CertmanagerV1().ClusterIssuers().Create(context.TODO(), notReadyClusterIssuer, metav1.CreateOptions{})

	err = checkClusterIssuerExistsAndReady(clientSet, "test")
	if err == nil {
		t.Fatal("Found Ready cluster issuer - Should be not Ready")
	}

	readyClusterIssuer := &certmanager.ClusterIssuer{
		ObjectMeta: metav1.ObjectMeta{Name: "test"},
		Spec: certmanager.IssuerSpec{
			IssuerConfig: certmanager.IssuerConfig{},
		},
		Status: certmanager.IssuerStatus{
			Conditions: []certmanager.IssuerCondition{
				{Type: "Ready", Status: "True"},
			},
		},
	}
	clientSet.CertmanagerV1().ClusterIssuers().Update(context.TODO(), readyClusterIssuer, metav1.UpdateOptions{})

	err = checkClusterIssuerExistsAndReady(clientSet, "test")
	if err != nil {
		t.Fatal(err)
	}
}
