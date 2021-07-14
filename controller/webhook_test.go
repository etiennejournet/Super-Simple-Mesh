package main

import (
	"k8s.io/client-go/rest"
	"os"
	"reflect"
	"testing"
)

func TestNewWebHook(t *testing.T) {
	//	wh := &webHook{
	//		Name:       defaultWebHookName,
	//		Namespace:  "test-namespace",
	//		Port:       defaultWebHookPort,
	//		RestConfig: &rest.Config{},
	//		EnvoyUID:   defaultEnvoyUserUID,
	//	}
	_, err := newWebHook(&rest.Config{})
	if err == nil {
		t.Fatal("Should throw error when looking for namespace")
	}
	os.Setenv("POD_NAMESPACE", "test-namespace")
	defer os.Unsetenv("POD_NAMESPACE")
	_, err = newWebHook(&rest.Config{})
	if err != nil {
		t.Fatal("Unexpected error -", err)
	}
	os.Setenv("WEBHOOK_PORT", "8080")
	defer os.Unsetenv("WEBHOOK_PORT")
	_, err = newWebHook(&rest.Config{})
	if err != nil {
		t.Fatal("Unexpected error -", err)
	}
	os.Setenv("WEBHOOK_PORT", "test")
	_, err = newWebHook(&rest.Config{})
	if err == nil {
		t.Fatal("Should have returned error, port is not an int")
	}
}

func TestDefineSidecar(t *testing.T) {
	wh := &webHook{
		Name:      "my-test-webhook",
		Namespace: "my-test-namespace",
		EnvoyUID:  "777",
	}
	container := wh.defineSidecar("test-cert-path")
	if reflect.TypeOf(container).String() != "*v1.Container" {
		t.Fatal("Type problem in sidecar creation, type found: " + reflect.TypeOf(container).String())
	}
}

func TestInitContainer(t *testing.T) {
	wh := &webHook{
		Name:      "my-test-webhook",
		Namespace: "my-test-namespace",
		EnvoyUID:  "777",
	}
	container := wh.defineInitContainer()
	if reflect.TypeOf(container).String() != "*v1.Container" {
		t.Fatal("Type problem in sidecar creation, type found: " + reflect.TypeOf(container).String())
	}
}

func TestGetEnvoyUID(t *testing.T) {
	wh := &webHook{
		EnvoyUID: "777",
	}
	if wh.getEnvoyUID() != wh.EnvoyUID {
		t.Fatal("Error getting EnvoyUID")
	}
}

func TestKubernetesClient(t *testing.T) {
	wh := &webHook{
		RestConfig: &rest.Config{},
	}
	_, err := wh.createKubernetesClientSet()
	if err != nil {
		t.Fatal(err)
	}

	wh.RestConfig = &rest.Config{
		QPS:   1,
		Burst: -1,
	}
	_, err = wh.createKubernetesClientSet()
	if err == nil {
		t.Fatal("Parameters of rest config are supposed to throw an error")
	}
}

func TestCertManagerClient(t *testing.T) {
	wh := &webHook{
		RestConfig: &rest.Config{},
	}
	_, err := wh.createCertManagerClientSet()
	if err != nil {
		t.Fatal(err)
	}

	wh.RestConfig = &rest.Config{
		QPS:   1,
		Burst: -1,
	}
	_, err = wh.createCertManagerClientSet()
	if err == nil {
		t.Fatal("Parameters of rest config are supposed to throw an error")
	}
}
