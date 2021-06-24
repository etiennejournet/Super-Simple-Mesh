package main

import (
	"encoding/json"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"net/http"
	"strconv"
)

func main() {
	ns, err := getCurrentNamespace()
	if err != nil {
		log.Fatal(err)
	}
	client, err := kubClient()
	if err != nil {
		log.Fatal(err)
	}
	wh := newWebHook("ssm", ns, 8443, 777, client)

	clientSet, err := wh.createKubernetesClientSet()
	if err != nil {
		log.Fatal(err)
	}

	cert, key, err := createSelfSignedCert(&wh)
	if err != nil {
		log.Fatal(err)
	}
	err = injectCAInMutatingWebhook(clientSet, wh.Name, cert)
	if err != nil {
		log.Fatal(err)
	}
	certPath, keyPath := writeCertsToHomeFolder(cert, key)

	http.HandleFunc("/", wh.server)
	log.Print("WebHook listening on port ", wh.Port)
	err = http.ListenAndServeTLS(":"+strconv.Itoa(wh.Port), certPath, keyPath, nil)
	if err != nil {
		log.Fatal(err)
	}
}

func (wh *webHook) server(w http.ResponseWriter, req *http.Request) {
	switch req.Method {
	case "GET":
		w.WriteHeader(http.StatusMethodNotAllowed)
		w.Write([]byte("POST requests accepted only \n"))

	case "POST":
		var jsonResponse []byte
		body, _ := ioutil.ReadAll(req.Body)
		if len(body) != 0 {
			jsonResponse, _ = json.Marshal(parseAndResolveInjectionDemand(body, wh))
		} else {
			w.WriteHeader(http.StatusBadRequest)
		}
		w.Write(jsonResponse)
	}
}
