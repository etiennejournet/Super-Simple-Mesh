package main

import (
	certManagerClient "github.com/jetstack/cert-manager/pkg/client/clientset/versioned"
	log "github.com/sirupsen/logrus"
	v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
	rest "k8s.io/client-go/rest"
	"strconv"
)

type webHook struct {
	Name       string
	Namespace  string
	Port       int
	RestConfig *rest.Config
	EnvoyUID   int
}

type webHookInterface interface {
	createCertManagerClientSet() certManagerClient.Interface
	defineSidecar(certificatesPath string) *v1.Container
	defineInitContainer() *v1.Container
	getName() string
	getEnvoyUID() int
}

func newWebHook(name string, namespace string, port int, envoyUID int, restConfig *rest.Config) webHook {
	log.Print("Starting " + name + "  webhook on port " + strconv.Itoa(port) + ", envoy User ID is " + strconv.Itoa(envoyUID))
	return webHook{
		Name:       name,
		Namespace:  namespace,
		Port:       port,
		EnvoyUID:   envoyUID,
		RestConfig: restConfig,
	}
}

func (wh *webHook) getName() string {
	return wh.Name
}

func (wh *webHook) getEnvoyUID() int {
	return wh.EnvoyUID
}

func (wh *webHook) defineSidecar(certificatesPath string) *v1.Container {
	return &v1.Container{
		Name:  wh.Name + "sidecar",
		Image: "etiennejournet/autoproxy:0.0.1",
		Env: []v1.EnvVar{
			{Name: "ENVOY_UID", Value: strconv.Itoa(wh.EnvoyUID)},
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
			"apk add iptables; iptables -t nat -A PREROUTING -p tcp -j REDIRECT --to-ports 10000; iptables -t nat -A OUTPUT -p tcp -m owner ! --uid-owner " + strconv.Itoa(wh.EnvoyUID) + " -j REDIRECT --to-ports 10001;",
		},
		SecurityContext: &v1.SecurityContext{
			Capabilities: &v1.Capabilities{
				Add: []v1.Capability{"NET_ADMIN"},
			},
		},
	}
}

func (wh *webHook) createKubernetesClientSet() kubernetes.Interface {
	clientSet, err := kubernetes.NewForConfig(wh.RestConfig)
	if err != nil {
		log.Print(err)
	}
	return clientSet
}

func (wh *webHook) createCertManagerClientSet() certManagerClient.Interface {
	clientSet, err := certManagerClient.NewForConfig(wh.RestConfig)
	if err != nil {
		log.Print(err)
	}
	return clientSet
}
