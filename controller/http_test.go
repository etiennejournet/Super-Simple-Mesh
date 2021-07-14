package main

import (
	"bytes"
	"encoding/json"
	admission "k8s.io/api/admission/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"net/http/httptest"
	"testing"
)

func TestHttpServer(t *testing.T) {
	wh := &webHook{
		Name:      "my-test-webhook",
		Namespace: "my-test-namespace",
		EnvoyUID:  "777",
	}
	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	wh.server(w, req)
	if statusCode := w.Result().StatusCode; statusCode != 405 {
		t.Fatal("Wrong return code for GET method : ", statusCode)
	}

	req = httptest.NewRequest("POST", "/", nil)
	w = httptest.NewRecorder()
	wh.server(w, req)
	if statusCode := w.Result().StatusCode; statusCode != 400 {
		t.Fatal("Wrong return code for empty payload : ", statusCode)
	}

	req = httptest.NewRequest("POST", "/", nil)
	w = httptest.NewRecorder()
	wh.server(w, req)
	if statusCode := w.Result().StatusCode; statusCode != 400 {
		t.Fatal("Wrong return code for empty payload : ", statusCode)
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
		},
	}
	admissionReviewMarshaled, _ := json.Marshal(admissionReview)
	admissionReviewBytes := bytes.NewBuffer(admissionReviewMarshaled)
	req = httptest.NewRequest("POST", "/", admissionReviewBytes)
	w = httptest.NewRecorder()
	wh.server(w, req)
}
