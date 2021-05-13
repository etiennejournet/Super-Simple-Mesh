package main

import (
	"encoding/json"
	"log"

	admission "k8s.io/api/admission/v1"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
)

func injectSidecar(admissionReviewBody []byte, wh *webHook) (admissionReview admission.AdmissionReview) {
	json.Unmarshal(admissionReviewBody, &admissionReview)
	patch := []patchValue{}
	log.Print("Mutation request received for object ", admissionReview.Request.Resource.Resource, " ", admissionReview.Request.Name, " in namespace ", admissionReview.Request.Namespace)

	podTemplate := getPodTemplateFromAdmissionRequest(admissionReview.Request)
	if podTemplate.Annotations["cert-manager.ssm.io/service-name"] != "" {
		log.Print("Patching demand for cert-manager received")

		mutationConfig := newCertManagerMutationConfig(
			wh,
			admissionReview.Request.Name,
			admissionReview.Request.Namespace,
			podTemplate,
		)
		patch = mutationConfig.createJSONPatch()
	} else if podTemplate.Annotations["autosidecar.ssm.io/enabled"] == "true" {
		log.Print("Patching demand for autocert received, not implemented yet")
		//patch = wh.autocertMutation(podTemplate)
	}

	patchByte, _ := json.Marshal(patch)
	admissionResponse := admission.AdmissionResponse{
		UID:     admissionReview.Request.UID,
		Allowed: true,
		Patch:   patchByte,
	}
	admissionReview.Response = &admissionResponse
	return
}

func getPodTemplateFromAdmissionRequest(admissionRequest *admission.AdmissionRequest) v1.PodTemplateSpec {
	switch admissionRequest.Resource.Resource {
	case "deployments":
		var object appsv1.Deployment
		if err := json.Unmarshal(admissionRequest.Object.Raw, &object); err != nil {
			log.Print("Error unmarshaling pod")
		}
		return object.Spec.Template
	case "daemonsets":
		var object appsv1.DaemonSet
		if err := json.Unmarshal(admissionRequest.Object.Raw, &object); err != nil {
			log.Print("Error unmarshaling pod")
		}
		return object.Spec.Template
	case "statefulsets":
		var object appsv1.StatefulSet
		if err := json.Unmarshal(admissionRequest.Object.Raw, &object); err != nil {
			log.Print("Error unmarshaling pod")
		}
		return object.Spec.Template
	}
	return v1.PodTemplateSpec{}
}
