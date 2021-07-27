package test

import (
	//"fmt"
	//"path/filepath"
	//"strings"
	//
	"testing"
	"time"
	//"github.com/stretchr/testify/require"

	"github.com/gruntwork-io/terratest/modules/k8s"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	//"github.com/gruntwork-io/terratest/modules/random"
)

func TestSimpleMeshCommunications(t *testing.T) {
	ssmNamespace := "super-simple-mesh"
	kubeconfig := "/etc/rancher/k3s/k3s.yaml"
	options := k8s.NewKubectlOptions("", kubeconfig, ssmNamespace)

	listOptions := metav1.ListOptions{
		LabelSelector: "app=ssm-injector",
	}
	k8s.WaitUntilNumPodsCreated(t, options, listOptions, 1, 5, 2)
	for i := 0; k8s.ListPods(t, options, listOptions)[0].Status.Phase != "Running" && i < 5; i++ {
		time.Sleep(2 * time.Second)
	}
	if k8s.ListPods(t, options, listOptions)[0].Status.Phase != "Running" {
		t.Log(k8s.ListPods(t, options, listOptions))
		t.Fatal("SSM not properly launched")
	}

	options = k8s.NewKubectlOptions("", kubeconfig, "default")
	k8s.KubectlApply(t, options, "manifest/nginx.yml")
	k8s.KubectlApply(t, options, "manifest/test-simple-mtls.yml")
  retriesDuration, _ := time.ParseDuration("2s")
  k8s.WaitUntilJobSucceed(t, kubectlOptions, "test-simple-mtls", 10, retriesDuration)
}
