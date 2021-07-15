package main

import (
	"context"
	"errors"
	log "github.com/sirupsen/logrus"
	"strconv"
	"time"

	certmanager "github.com/jetstack/cert-manager/pkg/apis/certmanager/v1"
	metacertmanager "github.com/jetstack/cert-manager/pkg/apis/meta/v1"
	certManagerClient "github.com/jetstack/cert-manager/pkg/client/clientset/versioned"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type certManagerMutationConfig struct {
	ObjectName                 string
	ObjectNamespace            string
	PodTemplate                *v1.PodTemplateSpec
	Certificate                *certmanager.Certificate
	SidecarConfiguration       *v1.Container
	InitContainerConfiguration *v1.Container
	Volume                     *v1.Volume
	VolumeMount                *v1.VolumeMount
	KubernetesClient           certManagerClient.Interface
}

const (
	defaultCertificatesPath  = "/var/run/ssm"
	defaultCertDuration      = "24h"
	defaultCertRenewalPeriod = "20h"
)

func newCertManagerMutationConfig(wh webHookInterface, objectName string, objectNamespace string, podTemplate v1.PodTemplateSpec) (*certManagerMutationConfig, error) {
	if objectName == "" || objectNamespace == "" {
		err := errors.New("Unable to read object or namespace names from admission request")
		return &certManagerMutationConfig{}, err
	}

	certificatesPath := defaultCertificatesPath

	kubernetesClient, err := wh.createCertManagerClientSet()
	if err != nil {
		return &certManagerMutationConfig{}, err
	}

	certManagerIssuer := wh.getCertManagerIssuer()
	err = checkClusterIssuerExistsAndReady(kubernetesClient, certManagerIssuer)
	if err != nil {
		return &certManagerMutationConfig{}, err
	}

	// Define the Certificate and Key Path
	certPath := podTemplate.Annotations["cert-manager.ssm.io/cert-path"]
	if certPath == "" {
		certPath = defaultCertificatesPath
	}

	// Define the Certificate Duration
	certDuration := podTemplate.Annotations["cert-manager.ssm.io/cert-duration"]
	renewBefore := defaultCertRenewalPeriod
	if certDuration == "" {
		certDuration = defaultCertDuration
	}
	certDurationParsed, _ := time.ParseDuration(certDuration)
	renewBeforeParsed, _ := time.ParseDuration(renewBefore)

	return &certManagerMutationConfig{
		ObjectName:      objectName,
		ObjectNamespace: objectNamespace,
		PodTemplate:     &podTemplate,
		Certificate: &certmanager.Certificate{
			ObjectMeta: metav1.ObjectMeta{
				Name:      objectName,
				Namespace: objectNamespace,
				Labels: map[string]string{
					"cert-manager.ssm.io/certificate-name": objectName,
				},
			},
			Spec: certmanager.CertificateSpec{
				CommonName:  podTemplate.Annotations["cert-manager.ssm.io/service-name"],
				DNSNames:    []string{podTemplate.Annotations["cert-manager.ssm.io/service-name"]},
				Duration:    &metav1.Duration{certDurationParsed},
				RenewBefore: &metav1.Duration{renewBeforeParsed},
				SecretName:  wh.getName() + "-cert-" + objectName,
				IssuerRef: metacertmanager.ObjectReference{
					Name: certManagerIssuer,
					// SSM only support one mesh for now
					Kind: "ClusterIssuer",
				},
			},
		},
		SidecarConfiguration:       wh.defineSidecar(certificatesPath),
		InitContainerConfiguration: wh.defineInitContainer(),
		Volume: &v1.Volume{
			Name: wh.getName() + "-volume",
			VolumeSource: v1.VolumeSource{
				Secret: &v1.SecretVolumeSource{
					SecretName: wh.getName() + "-cert-" + objectName,
				},
			},
		},
		VolumeMount: &v1.VolumeMount{
			Name:      wh.getName() + "-volume",
			MountPath: certificatesPath,
		},
		KubernetesClient: kubernetesClient,
	}, err
}

func (mutation *certManagerMutationConfig) createCertificateRequest() error {
	clientSet := mutation.KubernetesClient
	existingCert, err := clientSet.CertmanagerV1().Certificates(mutation.ObjectNamespace).List(
		context.TODO(),
		metav1.ListOptions{
			LabelSelector: "cert-manager.ssm.io/certificate-name",
		},
	)
	if err != nil {
		return err
	}

	for _, cert := range existingCert.Items {
		if cert.Name == mutation.ObjectName {
			log.Print("Cert " + mutation.ObjectName + " already exists in namespace " + mutation.ObjectNamespace + ", patching it")
			//TODO: Find a way of managing this more switfly
			err := clientSet.CertmanagerV1().Certificates(mutation.ObjectNamespace).Delete(context.TODO(), mutation.ObjectName, metav1.DeleteOptions{})
			if err != nil {
				return err
			}
		}
	}
	_, err = clientSet.CertmanagerV1().Certificates(mutation.ObjectNamespace).Create(context.TODO(), mutation.Certificate, metav1.CreateOptions{})

	return err
}

func (mutation *certManagerMutationConfig) createJSONPatch() []patchValue {
	//TODO: this function should implement idempotency
	log.Print("Creating secret: cert-", mutation.ObjectName, " in namespace default with certificateRequest")
	err := mutation.createCertificateRequest()
	if err != nil {
		log.Print(err)
		return []patchValue{}
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

	// Check if there already is an InitContainer , adding it as a new json array if there isn't
	initPatch := patchValue{
		Op:    "add",
		Path:  "/spec/template/spec/initContainers",
		Value: []v1.Container{*mutation.InitContainerConfiguration},
	}
	if len(mutation.PodTemplate.Spec.InitContainers) != 0 {
		initPatch.Value = mutation.InitContainerConfiguration
		for index, init := range mutation.PodTemplate.Spec.InitContainers {
			if init.Name == mutation.InitContainerConfiguration.Name {
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
		Value: mutation.SidecarConfiguration,
	}
	// Lets configure the volumeMount at the same time
	mountPatch := patchValue{
		Op:    "add",
		Path:  "/spec/template/spec/containers/" + strconv.Itoa(len(mutation.PodTemplate.Spec.Containers)) + "/volumeMounts",
		Value: []v1.VolumeMount{*mutation.VolumeMount},
	}
	for index, sidecar := range mutation.PodTemplate.Spec.Containers {
		if sidecar.Name == mutation.SidecarConfiguration.Name {
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

func checkClusterIssuerExistsAndReady(clientSet certManagerClient.Interface, clusterIssuerName string) error {
	clusterIssuer, err := clientSet.CertmanagerV1().ClusterIssuers().Get(context.TODO(), clusterIssuerName, metav1.GetOptions{})
	if err == nil && clusterIssuer.Status.Conditions[0].Status == "False" {
		err = errors.New("ClusterIssuer " + clusterIssuerName + " is not Ready")
	}
	return err
}
