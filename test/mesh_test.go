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

// An example of how to test the Kubernetes resource config in examples/kubernetes-basic-example using Terratest.
func TestKubernetesBasicExample(t *testing.T) {
	ssmNamespace := "super-simple-mesh"
	options := k8s.NewKubectlOptions("", "kubeconfig", ssmNamespace)
	//k8s.CreateNamespace(t, options, ssmNamespace)

	k8s.KubectlApply(t, options, "./manifest/clusterissuer.yml")
 // defer k8s.KubectlDelete(t, options, "./manifest/clusterissuer.yml")
	k8s.KubectlApply(t, options, "../deploy/manifest")
 // defer k8s.KubectlDelete(t, options, "../deploy/manifest")
	listOptions := metav1.ListOptions{
		LabelSelector: "app=ssm-injector",
	}
	k8s.WaitUntilNumPodsCreated(t, options, listOptions, 1, 5, 2)
	for i := 0; k8s.ListPods(t, options, listOptions)[0].Status.Phase != "Running" && i < 5; i++ {
		time.Sleep(2 * time.Second)
	}

	options = k8s.NewKubectlOptions("", "kubeconfig", "default")
	k8s.KubectlApply(t, options, "manifest/deploy.yml")
	//k8s.KubectlDelete(t, options, "manifest/deploy.yml")
}
