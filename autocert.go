package main

import (
	"log"

	v1 "k8s.io/api/core/v1"
)

func (wh *webHook) autocertMutation(podSpec v1.Pod) (patch []patchValue) {
	log.Print("Checking if the pod bears our annotation")
	if podSpec.ObjectMeta.Annotations["autosidecar.io/enabled"] == "true" {
		log.Print("Mounting certs in the autosidecar")
		for _, volumeMount := range podSpec.Spec.Containers[0].VolumeMounts {
			if volumeMount.Name == "certs" {
				wh.SidecarConfiguration.VolumeMounts = []v1.VolumeMount{volumeMount}
			}
		}
		patch = append(patch, patchValue{Op: "add", Path: "/spec/containers/-", Value: wh.SidecarConfiguration})
	} else {
		return
	}

	log.Print("Checking if there already is a sidecar defined")
	if len(podSpec.Spec.InitContainers) == 0 {
		patch = append(patch, patchValue{Op: "add", Path: "/spec/initContainers", Value: []v1.Container{*wh.InitContainerConfiguration}})
	} else {
		patch = append(patch, patchValue{Op: "add", Path: "/spec/initContainers/-", Value: wh.InitContainerConfiguration})
	}

	return
}
