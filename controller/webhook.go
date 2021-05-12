package main

import (
	v1 "k8s.io/api/core/v1"
	restclient "k8s.io/client-go/rest"
	"strconv"
)

type webHook struct {
	Name                       string
	Port                       int
	Client                     *restclient.Config
	SidecarConfiguration       *v1.Container
	InitContainerConfiguration *v1.Container
}

func newWebHook(name string, port int, envoyUID int) webHook {
	return webHook{
		Name:                       name,
		Port:                       port,
		Client:                     kubClient(),
		SidecarConfiguration:       defineSidecar(name, envoyUID),
		InitContainerConfiguration: defineInitContainer(name, envoyUID),
	}
}

func defineSidecar(name string, envoyUID int) *v1.Container {
	return &v1.Container{
		Name:  name + "sidecar",
		Image: "etiennejournet/autoproxy:0.0.1",
		Env: []v1.EnvVar{
			{Name: "ENVOY_UID", Value: strconv.Itoa(envoyUID)},
		},
		ImagePullPolicy: v1.PullAlways,
	}
}

func defineInitContainer(name string, envoyUID int) *v1.Container {
	return &v1.Container{
		Name:    name + "init",
		Image:   "alpine:latest",
		Command: []string{"/bin/sh", "-c"},
		Args: []string{
			"apk add iptables; iptables -t nat -A PREROUTING -p tcp -j REDIRECT --to-ports 10000; iptables -t nat -A OUTPUT -p tcp -m owner ! --uid-owner " + strconv.Itoa(envoyUID) + " -j REDIRECT --to-ports 10001;",
		},
		SecurityContext: &v1.SecurityContext{
			Capabilities: &v1.Capabilities{
				Add: []v1.Capability{"NET_ADMIN"},
			},
		},
	}
}
