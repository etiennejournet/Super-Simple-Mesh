package main

import (
	certManagerClient "github.com/jetstack/cert-manager/pkg/client/clientset/versioned"
	log "github.com/sirupsen/logrus"
	v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
	rest "k8s.io/client-go/rest"
	"os"
	"strconv"
)

const (
	defaultWebHookName       = "ssm"
	defaultWebHookPort       = 8443
	defaultEnvoyUserUID      = "777"
	defaultCertManagerIssuer = "caIssuer"
)

type webHook struct {
	Name              string
	Namespace         string
	Port              int
	RestConfig        *rest.Config
	EnvoyUID          string
	CertManagerIssuer string
}

type webHookInterface interface {
	createCertManagerClientSet() (certManagerClient.Interface, error)
	defineSidecar(certificatesPath string) *v1.Container
	defineInitContainer() *v1.Container
	getName() string
	getEnvoyUID() string
	getCertManagerIssuer() string
}

func newWebHook(kubClient *rest.Config) (webHook, error) {
	webHookName, ok := os.LookupEnv("WEBHOOK_NAME")
	if !ok {
		webHookName = defaultWebHookName
	}

	envoyUserUID, ok := os.LookupEnv("ENVOY_UID")
	if !ok {
		envoyUserUID = defaultEnvoyUserUID
	}

	certManagerIssuer, ok := os.LookupEnv("CERTMANAGER_ISSUER")
	if !ok {
		certManagerIssuer = defaultCertManagerIssuer
	}

	ns, err := getCurrentNamespace()
	if err != nil {
		return webHook{}, err
	}

	var webHookPort int
	webHookPortString, ok := os.LookupEnv("WEBHOOK_PORT")
	if !ok {
		webHookPort = defaultWebHookPort
	} else {
		webHookPort, err = strconv.Atoi(webHookPortString)
		if err != nil {
			return webHook{}, err
		}
	}

	log.Print("Starting " + webHookName + "  webhook on port " + strconv.Itoa(webHookPort) + ", envoy User ID is " + envoyUserUID)
	return webHook{
		Name:              webHookName,
		Namespace:         ns,
		Port:              webHookPort,
		EnvoyUID:          envoyUserUID,
		CertManagerIssuer: certManagerIssuer,
		RestConfig:        kubClient,
	}, nil
}

func (wh *webHook) getName() string {
	return wh.Name
}

func (wh *webHook) getEnvoyUID() string {
	return wh.EnvoyUID
}

func (wh *webHook) getCertManagerIssuer() string {
	return wh.CertManagerIssuer
}

func (wh *webHook) defineSidecar(certificatesPath string) *v1.Container {
	return &v1.Container{
		Name:  wh.Name + "sidecar",
		Image: "etiennejournet/autoproxy:0.0.1",
		Env: []v1.EnvVar{
			{Name: "ENVOY_UID", Value: wh.EnvoyUID},
			{Name: "CERTIFICATES_PATH", Value: certificatesPath},
		},
		ImagePullPolicy: v1.PullAlways,
	}
}

func (wh *webHook) defineInitContainer() *v1.Container {
	return &v1.Container{
		Name:    wh.Name + "init",
		Image:   "alpine:latest",
		Command: []string{"/bin/sh", "-c"},
		Args: []string{
			"apk add iptables; iptables -t nat -A PREROUTING -p tcp -j REDIRECT --to-ports 10000; iptables -t nat -A OUTPUT -p tcp -m owner ! --uid-owner " + wh.EnvoyUID + " -j REDIRECT --to-ports 10001;",
		},
		SecurityContext: &v1.SecurityContext{
			Capabilities: &v1.Capabilities{
				Add: []v1.Capability{"NET_ADMIN"},
			},
		},
	}
}

func (wh *webHook) createKubernetesClientSet() (kubernetes.Interface, error) {
	clientSet, err := kubernetes.NewForConfig(wh.RestConfig)
	if err != nil {
		log.Print(err)
	}
	return clientSet, err
}

func (wh *webHook) createCertManagerClientSet() (certManagerClient.Interface, error) {
	clientSet, err := certManagerClient.NewForConfig(wh.RestConfig)
	if err != nil {
		log.Print(err)
	}
	return clientSet, err
}
