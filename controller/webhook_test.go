package main

import (
	"k8s.io/client-go/rest"
	"reflect"
	"testing"
)

func TestNewWebHook(t *testing.T) {
	wh := &webHook{
		Name:             "my-test-webhook",
		Namespace:        "my-test-namespace",
		Port:             8443,
		KubernetesClient: &rest.Config{},
		EnvoyUID:         777,
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

