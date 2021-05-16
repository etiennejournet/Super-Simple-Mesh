package main

import (
	"encoding/json"
	"log"
  "errors"

	admission "k8s.io/api/admission/v1"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
)

func injectSidecar(admissionReviewBody []byte, wh *webHook) (admissionReview admission.AdmissionReview) {
	json.Unmarshal(admissionReviewBody, &admissionReview)
	patch := []patchValue{}
	log.Print("Mutation request received for object ", admissionReview.Request.Resource.Resource, " ", admissionReview.Request.Name, " in namespace ", admissionReview.Request.Namespace)

	err, podTemplate := getPodTemplateFromAdmissionRequest(admissionReview.Request)
  if err != nil {
    log.Print(err)
  } else if podTemplate.Annotations["cert-manager.ssm.io/service-name"] != "" {
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

	patchByte, err := json.Marshal(patch)
  if err != nil {
    log.Print(err)
  }

	admissionReview.Response = &admission.AdmissionResponse{
		UID:     admissionReview.Request.UID,
		Allowed: true,
		Patch:   patchByte,
	}
	return
}

func getPodTemplateFromAdmissionRequest(admissionRequest *admission.AdmissionRequest) (error, v1.PodTemplateSpec) {
	switch admissionRequest.Resource.Resource {
	case "deployments":
		var object appsv1.Deployment
    err := json.Unmarshal(admissionRequest.Object.Raw, &object)
		return err, object.Spec.Template
	case "daemonsets":
		var object appsv1.DaemonSet
    err := json.Unmarshal(admissionRequest.Object.Raw, &object)
		return err, object.Spec.Template
	case "statefulsets":
		var object appsv1.StatefulSet
    err := json.Unmarshal(admissionRequest.Object.Raw, &object)
		return err, object.Spec.Template
	}
  err := errors.New("This object is neither a deployment, a daemonset or a stateful set")
	return err, v1.PodTemplateSpec{}
}
