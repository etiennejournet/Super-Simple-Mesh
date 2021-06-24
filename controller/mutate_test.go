package main

import (
	"encoding/json"
	"reflect"
	"testing"

	admission "k8s.io/api/admission/v1"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func TestParseAndResolveInjectionDemand(t *testing.T) {
	whTest := &webHookTest{
		&webHook{
			Name:     "my-test-webhook",
			EnvoyUID: 777,
		},
	}
	podTemplateSpec := v1.PodTemplateSpec{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-pod-template",
			Annotations: map[string]string{
				"cert-manager.ssm.io/service-name": "Test",
			},
		},
	}
	rawPodTemplateSpec, _ := json.Marshal(podTemplateSpec)
	testObjectRawExtension := runtime.RawExtension{
		Raw: rawPodTemplateSpec,
	}
	admissionReview := admission.AdmissionReview{
		Request: &admission.AdmissionRequest{
			UID: "1",
			Kind: metav1.GroupVersionKind{
				Group:   "apps",
				Version: "v1",
				Kind:    "ReplicaSet",
			},
			Operation: "CREATE",
			Namespace: "test-namespace",
			Object:    testObjectRawExtension,
		},
	}
	admissionReviewMarshaled, _ := json.Marshal(admissionReview)

	admissionResponse := parseAndResolveInjectionDemand(admissionReviewMarshaled, whTest)
	if !admissionResponse.Response.Allowed {
		t.Fatal("Error building Admission Response with non-mutable object")
	}

	deployTemplateSpec := appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-deploy-template",
			Annotations: map[string]string{
				"cert-manager.ssm.io/service-name": "Test",
			},
		},
		Spec: appsv1.DeploymentSpec{
			Template: v1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-pod-template",
					Annotations: map[string]string{
						"cert-manager.ssm.io/service-name": "Test",
					},
				},
			},
		},
	}
	rawDeployTemplateSpec, _ := json.Marshal(deployTemplateSpec)
	testObjectRawExtension = runtime.RawExtension{
		Raw: rawDeployTemplateSpec,
	}
	admissionReview = admission.AdmissionReview{
		Request: &admission.AdmissionRequest{
			UID: "1",
			Kind: metav1.GroupVersionKind{
				Group:   "apps",
				Version: "v1",
				Kind:    "Deployment",
			},
			Operation: "CREATE",
			Namespace: "test-namespace",
			Object:    testObjectRawExtension,
		},
	}
	admissionReviewMarshaled, _ = json.Marshal(admissionReview)

	admissionResponse = parseAndResolveInjectionDemand(admissionReviewMarshaled, whTest)
	if admissionResponse.Response.Allowed != true {
		t.Fatal("Problem with admission response not allowed")
		t.Fatal(admissionResponse.Response)
	}
}

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
