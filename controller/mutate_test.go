package main

import (
	"encoding/json"
	"reflect"
	"testing"

	admission "k8s.io/api/admission/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func TestGetPodTemplateFromAdmissionRequest(t *testing.T) {
	podTemplateSpec := v1.PodTemplateSpec{
		ObjectMeta: metav1.ObjectMeta{Name: "test-pod-template"},
	}
	rawPodTemplateSpec, _ := json.Marshal(podTemplateSpec)
	testObjectRawExtension := runtime.RawExtension{
		Raw: rawPodTemplateSpec,
	}

	admissionRequestForDeployment := &admission.AdmissionRequest{
		UID: "1",
		Kind: metav1.GroupVersionKind{
			Group:   "apps",
			Version: "v1",
			Kind:    "Deployment",
		},
		Operation: "CREATE",
		Namespace: "test-namespace",
		Object:    testObjectRawExtension,
	}
	podTemplate, err := getPodTemplateFromAdmissionRequest(admissionRequestForDeployment)

	if err != nil {
		t.Fatal(err)
	}

	if reflect.TypeOf(podTemplate).String() != "v1.PodTemplateSpec" {
		t.Fatal("Wrong return type")
	}

	admissionRequestForStatefulSet := &admission.AdmissionRequest{
		UID: "1",
		Kind: metav1.GroupVersionKind{
			Group:   "apps",
			Version: "v1",
			Kind:    "StatefulSet",
		},
		Operation: "CREATE",
		Namespace: "test-namespace",
		Object:    testObjectRawExtension,
	}
	podTemplate, err = getPodTemplateFromAdmissionRequest(admissionRequestForStatefulSet)

	if err != nil {
		t.Fatal(err)
	}

	if reflect.TypeOf(podTemplate).String() != "v1.PodTemplateSpec" {
		t.Fatal("Wrong return type")
	}

	admissionRequestForDaemonSet := &admission.AdmissionRequest{
		UID: "1",
		Kind: metav1.GroupVersionKind{
			Group:   "apps",
			Version: "v1",
			Kind:    "DaemonSet",
		},
		Operation: "CREATE",
		Namespace: "test-namespace",
		Object:    testObjectRawExtension,
	}
	podTemplate, err = getPodTemplateFromAdmissionRequest(admissionRequestForDaemonSet)

	if err != nil {
		t.Fatal(err)
	}

	if reflect.TypeOf(podTemplate).String() != "v1.PodTemplateSpec" {
		t.Fatal("Wrong return type")
	}

	admissionRequestForOther := &admission.AdmissionRequest{
		UID: "1",
		Kind: metav1.GroupVersionKind{
			Group:   "apps",
			Version: "v1",
			Kind:    "ReplicaSet",
		},
		Operation: "CREATE",
		Namespace: "test-namespace",
		Object:    testObjectRawExtension,
	}
	podTemplate, err = getPodTemplateFromAdmissionRequest(admissionRequestForOther)

	if err == nil {
		t.Fatal("No error returned for incompatible replicaset")
	}
}
