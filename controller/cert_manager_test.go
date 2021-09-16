package main

import (
	"context"
	"errors"
	certmanager "github.com/jetstack/cert-manager/pkg/apis/certmanager/v1"
	certManagerClient "github.com/jetstack/cert-manager/pkg/client/clientset/versioned"
	certManagerTesting "github.com/jetstack/cert-manager/pkg/client/clientset/versioned/fake"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
	"testing"
)

type webHookTest struct {
	webHookInterface
}

func (wh *webHookTest) createCertManagerClientSet() (certManagerClient.Interface, error) {
	clientSet := certManagerTesting.NewSimpleClientset()
	readyClusterIssuer := &certmanager.ClusterIssuer{
		ObjectMeta: metav1.ObjectMeta{Name: "test-issuer"},
		Spec: certmanager.IssuerSpec{
			IssuerConfig: certmanager.IssuerConfig{},
		},
		Status: certmanager.IssuerStatus{
			Conditions: []certmanager.IssuerCondition{
				{Type: "Ready", Status: "True"},
			},
		},
	}
	_, err := clientSet.CertmanagerV1().ClusterIssuers().Create(context.TODO(), readyClusterIssuer, metav1.CreateOptions{})
	if err != nil {
		return clientSet, err
	}

	if wh.getName() == "test-certificate-exists" {
		certificate := &certmanager.Certificate{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "my-test-object",
				Namespace: "my-test-namespace",
				Labels: map[string]string{
					"cert-manager.ssm.io/certificate-name": "my-test-object",
				},
			},
		}
		clientSet.CertmanagerV1().Certificates("my-test-namespace").Create(context.TODO(), certificate, metav1.CreateOptions{})
	}

	if wh.getName() == "test-certificate-exists-no-labels" {
		certificate := &certmanager.Certificate{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "my-test-object",
				Namespace: "my-test-namespace",
			},
		}
		clientSet.CertmanagerV1().Certificates("my-test-namespace").Create(context.TODO(), certificate, metav1.CreateOptions{})
	}

	if wh.getName() == "test-object-fails" {
		err := errors.New("normal error for testing purpose")
		return nil, err
	}
	return clientSet, nil
}

func TestNewCertManagerMutationConfig(t *testing.T) {
	whTest := &webHookTest{
		&webHook{
			Name:              "my-test-webhook",
			EnvoyUID:          "777",
			CertManagerIssuer: "test-issuer",
			RestConfig:        &rest.Config{},
		},
	}
	podTemplateSpec := &corev1.PodTemplateSpec{
		ObjectMeta: metav1.ObjectMeta{Name: "test-pod-template"},
	}
	_, err := newCertManagerMutationConfig(whTest, "my-test-object", "my-test-namespace", podTemplateSpec)
	if err != nil {
		t.Fatal(err)
	}

	whTest = &webHookTest{
		&webHook{
			Name:       "my-test-webhook",
			EnvoyUID:   "777",
			RestConfig: &rest.Config{},
		},
	}
	podTemplateSpec = &corev1.PodTemplateSpec{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-pod-template",
		},
	}
	_, err = newCertManagerMutationConfig(whTest, "my-test-object", "my-test-namespace", podTemplateSpec)
	if err == nil {
		t.Fatal("There should be an error - cluster issuer doesnt exist")
	}

	whTest = &webHookTest{
		&webHook{
			Name:              "test-object-fails",
			EnvoyUID:          "777",
			CertManagerIssuer: "test-issuer",
			RestConfig:        &rest.Config{},
		},
	}
	_, err = newCertManagerMutationConfig(whTest, "my-test-object", "my-test-namespace", podTemplateSpec)
	if err == nil {
		t.Fatal("Error should be returned here, none was")
	}
}

func TestNewCreateCertificateRequest(t *testing.T) {
	whTest := &webHookTest{
		&webHook{
			Name:              "my-test-webhook",
			EnvoyUID:          "777",
			CertManagerIssuer: "test-issuer",
			RestConfig:        &rest.Config{},
		},
	}
	podTemplateSpec := &corev1.PodTemplateSpec{
		ObjectMeta: metav1.ObjectMeta{Name: "test-pod-template"},
	}

	mutationConfig, err := newCertManagerMutationConfig(whTest, "my-test-object-not-existing-cert", "my-test-namespace", podTemplateSpec)
	if err != nil {
		t.Fatal(err)
	}
	err = mutationConfig.createCertificateRequest()
	if err != nil {
		t.Fatal(err)
	}

	whTest = &webHookTest{
		&webHook{
			Name:              "test-certificate-exists",
			EnvoyUID:          "777",
			CertManagerIssuer: "test-issuer",
			RestConfig:        &rest.Config{},
		},
	}
	podTemplateSpec = &corev1.PodTemplateSpec{
		ObjectMeta: metav1.ObjectMeta{Name: "test-pod-template"},
	}

	mutationConfig, err = newCertManagerMutationConfig(whTest, "my-test-object", "my-test-namespace", podTemplateSpec)
	if err != nil {
		t.Fatal(err)
	}
	err = mutationConfig.createCertificateRequest()
	if err != nil {
		t.Fatal(err)
	}

	whTest = &webHookTest{
		&webHook{
			Name:              "test-certificate-exists-no-labels",
			EnvoyUID:          "777",
			RestConfig:        &rest.Config{},
			CertManagerIssuer: "test-issuer",
		},
	}
	podTemplateSpec = &corev1.PodTemplateSpec{
		ObjectMeta: metav1.ObjectMeta{Name: "test-pod-template"},
	}

	mutationConfig, err = newCertManagerMutationConfig(whTest, "my-test-object", "my-test-namespace", podTemplateSpec)
	if err != nil {
		t.Fatal(err)
	}
	err = mutationConfig.createCertificateRequest()
	if err == nil {
		t.Fatal("There should be an error about certificate already existing here")
	}
}

func TestCheckClusterIssuerExistsAndReady(t *testing.T) {
	clientSet := certManagerTesting.NewSimpleClientset()
	err := checkClusterIssuerExistsAndReady(clientSet, "test")
	if err == nil {
		t.Fatal("Found absent cluster issuer")
	}

	notReadyClusterIssuer := &certmanager.ClusterIssuer{
		ObjectMeta: metav1.ObjectMeta{Name: "test"},
		Spec: certmanager.IssuerSpec{
			IssuerConfig: certmanager.IssuerConfig{},
		},
		Status: certmanager.IssuerStatus{
			Conditions: []certmanager.IssuerCondition{
				{Type: "Ready", Status: "False"},
			},
		},
	}
	clientSet.CertmanagerV1().ClusterIssuers().Create(context.TODO(), notReadyClusterIssuer, metav1.CreateOptions{})

	err = checkClusterIssuerExistsAndReady(clientSet, "test")
	if err == nil {
		t.Fatal("Found Ready cluster issuer - Should be not Ready")
	}

	readyClusterIssuer := &certmanager.ClusterIssuer{
		ObjectMeta: metav1.ObjectMeta{Name: "test"},
		Spec: certmanager.IssuerSpec{
			IssuerConfig: certmanager.IssuerConfig{},
		},
		Status: certmanager.IssuerStatus{
			Conditions: []certmanager.IssuerCondition{
				{Type: "Ready", Status: "True"},
			},
		},
	}
	clientSet.CertmanagerV1().ClusterIssuers().Update(context.TODO(), readyClusterIssuer, metav1.UpdateOptions{})

	err = checkClusterIssuerExistsAndReady(clientSet, "test")
	if err != nil {
		t.Fatal(err)
	}
}

func TestCreateJSONPatch(t *testing.T) {
	whTest := &webHookTest{
		&webHook{
			Name:              "my-test-webhook",
			EnvoyUID:          "777",
			RestConfig:        &rest.Config{},
			CertManagerIssuer: "test-issuer",
		},
	}
	podTemplateSpec := &corev1.PodTemplateSpec{
		ObjectMeta: metav1.ObjectMeta{Name: "test-pod-template"},
	}

	certManagerMutationConfiguration, err := newCertManagerMutationConfig(whTest, "my-test-object", "my-test-namespace", podTemplateSpec)
	if err != nil {
		t.Fatal(err)
	}

	patch := certManagerMutationConfiguration.createJSONPatch()
	if len(patch) != 4 {
		t.Fatal("Number of patches on empty object should be 4")
	}

	podTemplateSpecWithVolumes := &corev1.PodTemplateSpec{
		ObjectMeta: metav1.ObjectMeta{Name: "test-pod-template"},
		Spec: corev1.PodSpec{
			Volumes: []corev1.Volume{{
				Name: "my-test-volume",
			}},
		},
	}

	certManagerMutationConfiguration, err = newCertManagerMutationConfig(whTest, "my-test-object", "my-test-namespace", podTemplateSpecWithVolumes)
	if err != nil {
		t.Fatal(err)
	}
	patch = certManagerMutationConfiguration.createJSONPatch()
	if patch[1].Op != "add" {
		t.Fatal("Problem in patch type for Volumes, should be add")
	}

	podTemplateSpecWithVolumes = &corev1.PodTemplateSpec{
		ObjectMeta: metav1.ObjectMeta{Name: "test-pod-template"},
		Spec: corev1.PodSpec{
			Volumes: []corev1.Volume{{
				Name: whTest.getName() + "-volume",
			}},
		},
	}
	certManagerMutationConfiguration, err = newCertManagerMutationConfig(whTest, "my-test-object", "my-test-namespace", podTemplateSpecWithVolumes)
	if err != nil {
		t.Fatal(err)
	}
	patch = certManagerMutationConfiguration.createJSONPatch()
	if patch[1].Op != "replace" {
		t.Fatal("Problem in patch type for Volumes, should be replace")
	}

	podTemplateSpecWithInitContainer := &corev1.PodTemplateSpec{
		ObjectMeta: metav1.ObjectMeta{Name: "test-pod-template"},
		Spec: corev1.PodSpec{
			InitContainers: []corev1.Container{{
				Name: "test",
			}},
		},
	}
	certManagerMutationConfiguration, err = newCertManagerMutationConfig(whTest, "my-test-object", "my-test-namespace", podTemplateSpecWithInitContainer)
	if err != nil {
		t.Fatal(err)
	}
	patch = certManagerMutationConfiguration.createJSONPatch()
	if patch[2].Op != "add" {
		t.Fatal("Problem in patch type for InitContainer, should be add")
	}

	podTemplateSpecWithInitContainer = &corev1.PodTemplateSpec{
		ObjectMeta: metav1.ObjectMeta{Name: "test-pod-template"},
		Spec: corev1.PodSpec{
			InitContainers: []corev1.Container{{
				Name: certManagerMutationConfiguration.InitContainerConfiguration.Name,
			}},
		},
	}
	certManagerMutationConfiguration, err = newCertManagerMutationConfig(whTest, "my-test-object", "my-test-namespace", podTemplateSpecWithInitContainer)
	if err != nil {
		t.Fatal(err)
	}
	patch = certManagerMutationConfiguration.createJSONPatch()
	if patch[2].Op != "replace" {
		t.Fatal("Problem in patch type for InitContainer, should be replace")
	}

	podTemplateSpecWithSidecarContainer := &corev1.PodTemplateSpec{
		ObjectMeta: metav1.ObjectMeta{Name: "test-pod-template"},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{{
				Name: "test",
			}},
		},
	}
	certManagerMutationConfiguration, err = newCertManagerMutationConfig(whTest, "my-test-object", "my-test-namespace", podTemplateSpecWithSidecarContainer)
	if err != nil {
		t.Fatal(err)
	}
	patch = certManagerMutationConfiguration.createJSONPatch()
	if patch[0].Op != "add" {
		t.Fatal("Problem in patch type for SidecarContainer, should be add")
	}

	podTemplateSpecWithSidecarContainer = &corev1.PodTemplateSpec{
		ObjectMeta: metav1.ObjectMeta{Name: "test-pod-template"},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{{
				Name: certManagerMutationConfiguration.SidecarConfiguration.Name,
			}},
		},
	}
	certManagerMutationConfiguration, err = newCertManagerMutationConfig(whTest, "my-test-object", "my-test-namespace", podTemplateSpecWithSidecarContainer)
	if err != nil {
		t.Fatal(err)
	}
	patch = certManagerMutationConfiguration.createJSONPatch()
	if patch[0].Op != "replace" {
		t.Fatal("Problem in patch type for SidecarContainer, should be replace")
	}
}
