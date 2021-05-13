package main

import (
	"context"
	"log"
	"strconv"

	certmanager "github.com/jetstack/cert-manager/pkg/apis/certmanager/v1"
	metacertmanager "github.com/jetstack/cert-manager/pkg/apis/meta/v1"
	"github.com/jetstack/cert-manager/pkg/client/clientset/versioned"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	v1 "k8s.io/api/core/v1"
	"time"
)

type certManagerMutationConfig struct {
	ObjectName           string
	ObjectNamespace      string
	PodTemplate          *v1.PodTemplateSpec
	Certificate          *certmanager.Certificate
	Volume               *v1.Volume
	VolumeMount          *v1.VolumeMount
	WebHookConfiguration *webHook
}

func newCertManagerMutationConfig(wh *webHook, objectName string, objectNamespace string, podTemplate v1.PodTemplateSpec) *certManagerMutationConfig {
	return &certManagerMutationConfig{
		ObjectName:      objectName,
		ObjectNamespace: objectNamespace,
		PodTemplate:     &podTemplate,
		Certificate: &certmanager.Certificate{
			ObjectMeta: metav1.ObjectMeta{
				Name: objectName,
			},
			Spec: certmanager.CertificateSpec{
				CommonName: podTemplate.Annotations["cert-manager.ssm.io/service-name"],
				DNSNames:   []string{podTemplate.Annotations["cert-manager.ssm.io/service-name"]},
				//TODO : find a way to make this variable
				Duration: &metav1.Duration{5 * time.Hour},
				//TODO : find a way to make this variable
				RenewBefore: &metav1.Duration{4 * time.Hour},
				SecretName:  wh.Name + "-cert-" + objectName,
				IssuerRef: metacertmanager.ObjectReference{
					//TODO : find a way to make this variable
					Name: "ca-issuer",
					Kind: "ClusterIssuer",
				},
			},
		},
		Volume: &v1.Volume{
			Name: wh.Name + "-volume",
			VolumeSource: v1.VolumeSource{
				Secret: &v1.SecretVolumeSource{
					SecretName: wh.Name + "-cert-" + objectName,
					Items: []v1.KeyToPath{
						// TODO : replace this mapping by a templating in the docker image"
						{Key: "ca.crt", Path: "root.crt"},
						{Key: "tls.crt", Path: "site.crt"},
						{Key: "tls.key", Path: "site.key"},
					},
				},
			},
		},
		VolumeMount: &v1.VolumeMount{
			Name: wh.Name + "-volume",
			// TODO : replace this hardcoded var
			MountPath: "/var/run/autocert.step.sm/",
		},
		// TODO : inherit only needed fields from WebHook
		WebHookConfiguration: wh,
	}
}

func (mutation *certManagerMutationConfig) certificateRequest() (err error) {
	// We create a new cert
	clientSet, err := versioned.NewForConfig(mutation.WebHookConfiguration.Client)
	if err != nil {
		log.Fatal(err)
	}

	existingCert, err := clientSet.CertmanagerV1().Certificates(mutation.ObjectNamespace).Get(context.TODO(), mutation.ObjectName, metav1.GetOptions{})
	if err != nil {
		log.Print(err)
	}
	if existingCert != nil {
		log.Print("Cert already exists by the same name, patching it")
      // Find a way of managing this more switfly
		clientSet.CertmanagerV1().Certificates(mutation.ObjectNamespace).Delete(context.TODO(), mutation.ObjectName, metav1.DeleteOptions{})
	}
	_, err = clientSet.CertmanagerV1().Certificates(mutation.ObjectNamespace).Create(context.TODO(), mutation.Certificate, metav1.CreateOptions{})

	return
}

func (mutation *certManagerMutationConfig) certManagerMutation() []patchValue {
	log.Print("Creating secret: cert-", mutation.ObjectName, " in namespace default with certificateRequest")
	err := mutation.certificateRequest()
	if err != nil {
		log.Print(err)
	}

	// Check if there already is a Volume, adding it as a new json array if there isn't
	volumePatch := patchValue{
		Op:    "add",
		Path:  "/spec/template/spec/volumes",
		Value: []v1.Volume{*mutation.Volume},
	}
	if len(mutation.PodTemplate.Spec.Volumes) != 0 {
		volumePatch.Value = mutation.Volume
		for index, volume := range mutation.PodTemplate.Spec.Volumes {
			if volume.Name == mutation.Volume.Name {
				volumePatch.Op = "replace"
				volumePatch.Path = "/spec/template/spec/volumes/" + strconv.Itoa(index)
			}
		}
		if volumePatch.Op == "add" {
			volumePatch.Path = "/spec/template/spec/volumes/-"
		}
	}

	// Check if there already is an IniContainer , adding it as a new json array if there isn't
	initPatch := patchValue{
		Op:    "add",
		Path:  "/spec/template/spec/initContainers",
		Value: []v1.Container{*mutation.WebHookConfiguration.InitContainerConfiguration},
	}
	if len(mutation.PodTemplate.Spec.InitContainers) != 0 {
		initPatch.Value = mutation.WebHookConfiguration.InitContainerConfiguration
		for index, init := range mutation.PodTemplate.Spec.InitContainers {
			if init.Name == mutation.WebHookConfiguration.InitContainerConfiguration.Name {
				initPatch.Op = "replace"
				initPatch.Path = "/spec/template/spec/initContainers/" + strconv.Itoa(index)
			}
		}
		if initPatch.Op == "add" {
			initPatch.Path = "/spec/template/spec/initContainers/-"
		}
	}

	// Add or replace Sidecar
	sidecarPatch := patchValue{
		Op:    "add",
		Path:  "/spec/template/spec/containers/-",
		Value: mutation.WebHookConfiguration.SidecarConfiguration,
	}
	// Lets configure the volumeMount at the same time
	mountPatch := patchValue{
		Op:    "add",
		Path:  "/spec/template/spec/containers/" + strconv.Itoa(len(mutation.PodTemplate.Spec.Containers)) + "/volumeMounts",
		Value: []v1.VolumeMount{*mutation.VolumeMount},
	}
	for index, sidecar := range mutation.PodTemplate.Spec.Containers {
		if sidecar.Name == mutation.WebHookConfiguration.SidecarConfiguration.Name {
			sidecarPatch.Op = "replace"
			sidecarPatch.Path = "/spec/template/spec/containers/" + strconv.Itoa(index)
			mountPatch.Path = sidecarPatch.Path + "/volumeMounts"
		}
	}

	return []patchValue{
    sidecarPatch,
    volumePatch,
    initPatch,
    mountPatch}
}
