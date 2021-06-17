package main

import (
	"testing"
)

func TestGetCurrentNamespace(t *testing.T) {
	_, err := getCurrentNamespace()
	if err == nil {
		t.Fatal("Should not be able to find current Namespace")
	}
}
