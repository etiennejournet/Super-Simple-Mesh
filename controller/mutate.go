package main

import (
	"encoding/json"
	"errors"
	log "github.com/sirupsen/logrus"

	admission "k8s.io/api/admission/v1"
	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	v1 "k8s.io/api/core/v1"
)

func parseAndResolveInjectionDemand(admissionReviewBody []byte, wh webHookInterface) (admissionReview admission.AdmissionReview) {
	json.Unmarshal(admissionReviewBody, &admissionReview)
	patch := []patchValue{}
	var patchType admission.PatchType = "JSONPatch"

	log.Print("Mutation request received for object ", admissionReview.Request.Resource.Resource, " ", admissionReview.Request.Name, " in namespace ", admissionReview.Request.Namespace)

	podTemplate, err := getPodTemplateFromAdmissionRequest(admissionReview.Request)
	if err != nil {
		log.Print(err)
	} else if podTemplate.Annotations["cert-manager.ssm.io/service-name"] != "" {
		log.Print("Patching demand of type cert-manager received")

		mutationConfig, err := newCertManagerMutationConfig(
			wh,
			admissionReview.Request.Name,
			admissionReview.Request.Namespace,
			&podTemplate,
		)
		if err == nil {
			patch = mutationConfig.createJSONPatch()
		} else {
			log.Print(err)
		}
	} else if podTemplate.Annotations["autosidecar.ssm.io/enabled"] == "true" {
		log.Print("Patching demand for autocert received, not implemented yet")
	}

	patchByte, err := json.Marshal(patch)
	if err != nil {
		log.Print(err)
	}

	admissionReview.Response = &admission.AdmissionResponse{
		UID:       admissionReview.Request.UID,
		Allowed:   true,
		PatchType: &patchType,
		Patch:     patchByte,
	}
	return
}

func getPodTemplateFromAdmissionRequest(admissionRequest *admission.AdmissionRequest) (v1.PodTemplateSpec, error) {
	switch admissionRequest.Kind.Kind {
	case "Deployment":
		var object appsv1.Deployment
		err := json.Unmarshal(admissionRequest.Object.Raw, &object)
		return object.Spec.Template, err
	case "DaemonSet":
		var object appsv1.DaemonSet
		err := json.Unmarshal(admissionRequest.Object.Raw, &object)
		return object.Spec.Template, err
	case "StatefulSet":
		var object appsv1.StatefulSet
		err := json.Unmarshal(admissionRequest.Object.Raw, &object)
		return object.Spec.Template, err
	case "Job":
		var object batchv1.Job
		err := json.Unmarshal(admissionRequest.Object.Raw, &object)
		return object.Spec.Template, err
	}
	err := errors.New("this object is neither a deployment, a daemonset or a stateful set: " + admissionRequest.Kind.Kind)
	return v1.PodTemplateSpec{}, err
}
