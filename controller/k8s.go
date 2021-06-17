package main

import (
	"flag"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"os"
	"strings"
)

func kubClient() *rest.Config {
	//TODO: refacto this
	kubeconfig := flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	flag.Parse()

	if *kubeconfig == "" {
		log.Info("Detected in cluster launch")
		// creates the in-cluster config
		config, err := rest.InClusterConfig()
		if err != nil {
			log.Fatal(err)
		}
		return config
	}

	// use the current context in kubeconfig
	log.Info("Detected Kubeconfig flag")
	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		log.Fatal(err)
	}
	return config
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
