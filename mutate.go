package main

import (
	"encoding/json"
	"log"

	admission "k8s.io/api/admission/v1"
	//	v1 "k8s.io/api/core/v1"
	appsv1 "k8s.io/api/apps/v1"
)

func injectSidecar(admissionReviewBody []byte, wh *webHook) (admissionReview admission.AdmissionReview) {
	json.Unmarshal(admissionReviewBody, &admissionReview)
	patch := []patchValue{}
	log.Print("Mutation request received for object ", admissionReview.Request.Resource.Resource, " ", admissionReview.Request.Name, " in namespace ", admissionReview.Request.Namespace)

	//  switch admissionReview.Request.Resource.Resource {
	//  case "deployments":
	var object appsv1.Deployment
	if err := json.Unmarshal(admissionReview.Request.Object.Raw, &object); err != nil {
		log.Fatal("Error unmarshaling pod")
	}
	mutationConfig := newCertManagerMutationConfig(
		wh,
		admissionReview.Request.Name,
		admissionReview.Request.Namespace,
		object.Spec.Template,
	)
	if mutationConfig.PodTemplate.Annotations["cert-manager.ssm.io/service-name"] != "" {
		log.Print("Patching demand for cert-manager received")
		patch = mutationConfig.certManagerMutation()
	}
	//  }

	//  if object.Annotations["autosidecar.ssm.io/enabled"] == "true" {
	//	  patch = wh.autocertMutation(podTemplate)
	//  }
	//  } else

	patchByte, _ := json.Marshal(patch)
	admissionResponse := admission.AdmissionResponse{
		UID:     admissionReview.Request.UID,
		Allowed: true,
		Patch:   patchByte,
	}
	admissionReview.Response = &admissionResponse
	return
}
