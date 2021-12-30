package main

import (
	"os"
	"testing"
)

func TestGetCurrentNamespace(t *testing.T) {
	_, err := getCurrentNamespace()
	if err == nil {
		t.Fatal("Should not be able to find current Namespace")
	}

	os.Setenv("POD_NAMESPACE", "test")
	defer os.Unsetenv("POD_NAMESPACE")
	ns, err := getCurrentNamespace()
	if err != nil {
		t.Fatal(err)
	}
	if ns != "test" {
		t.Fatal("Didn't found proper namespace name. Found: ", ns)
	}
}

func TestGetRestClient(t *testing.T) {
	_, err := kubClient()
	if err == nil {
		t.Fatal("kubClient responded with no error but it should have")
	}
}
