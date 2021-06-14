package main

import (
	"context"
	certmanager "github.com/jetstack/cert-manager/pkg/apis/certmanager/v1"
	certManagerTesting "github.com/jetstack/cert-manager/pkg/client/clientset/versioned/fake"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"testing"
)

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
		t.Fatal("Found Ready cluster issuer - Should be not Ready")
	}
}
