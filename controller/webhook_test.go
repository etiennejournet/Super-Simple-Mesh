package main

import (
	"k8s.io/client-go/rest"
	"reflect"
	"testing"
)

func TestNewWebHook(t *testing.T) {
	wh := &webHook{
		Name:       "my-test-webhook",
		Namespace:  "my-test-namespace",
		Port:       8443,
		RestConfig: &rest.Config{},
		EnvoyUID:   777,
	}
	testWebHook := newWebHook(wh.Name, wh.Namespace, wh.Port, wh.EnvoyUID, &rest.Config{})
	if !reflect.DeepEqual(testWebHook, *wh) {
		t.Fatal("Unexpected error creating webhook object")
	}
}

func TestDefineSidecar(t *testing.T) {
	wh := &webHook{
		Name:      "my-test-webhook",
		Namespace: "my-test-namespace",
		EnvoyUID:  777,
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
		EnvoyUID:  777,
	}
	container := wh.defineInitContainer()
	if reflect.TypeOf(container).String() != "*v1.Container" {
		t.Fatal("Type problem in sidecar creation, type found: " + reflect.TypeOf(container).String())
	}
}

func TestGetEnvoyUID(t *testing.T) {
	wh := &webHook{
		EnvoyUID: 777,
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
