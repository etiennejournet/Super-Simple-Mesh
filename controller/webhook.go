package main

import (
	v1 "k8s.io/api/core/v1"
   log "github.com/sirupsen/logrus"
	restclient "k8s.io/client-go/rest"
	"strconv"
)

type webHook struct {
	Name             string
	Namespace        string
	Port             int
	KubernetesClient *restclient.Config
	EnvoyUID         int
}

func newWebHook(name string, port int, envoyUID int) webHook {
	log.Print("Starting " + name + "  webhook on port " + strconv.Itoa(port) + ", envoy User ID is " + strconv.Itoa(envoyUID))
	return webHook{
		Name:             name,
		Namespace:        getNamespace(),
		Port:             port,
		EnvoyUID:         envoyUID,
		KubernetesClient: kubClient(),
	}
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
