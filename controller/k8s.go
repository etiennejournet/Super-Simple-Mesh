package main

import (
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"k8s.io/client-go/rest"
	"os"
	"strings"
)

func kubClient() (*rest.Config, error) {
	config, err := rest.InClusterConfig()
	if err != nil {
		return nil, err
	}
	return config, nil
}

func getCurrentNamespace() (string, error) {
	var ns string
	if ns, ok := os.LookupEnv("POD_NAMESPACE"); ok {
		return ns, nil
	}

	if data, err := ioutil.ReadFile("/var/run/secrets/kubernetes.io/serviceaccount/namespace"); err == nil {
		if ns = strings.TrimSpace(string(data)); ns == "" {
			log.Print("Unable to read current Namespace. Exiting")
			return ns, err
		}
	} else {
		log.Print("Unable to read current Namespace. Exiting")
		return ns, err
	}
	return ns, nil
}
