package main

import (
	"encoding/json"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"net/http"
	"strconv"
)

func main() {
	wh := newWebHook("ssm", getNamespace(), 8443, 777, kubClient())

	cert, key := createSelfSignedCert(&wh)
	injectCAInMutatingWebhook(createKubernetesClientSet(wh.KubernetesClient), wh.Name, cert)
	certPath, keyPath := writeCertsToHomeFolder(cert, key)

	http.HandleFunc("/", wh.server)
	log.Print("WebHook listening on port ", wh.Port)
	err := http.ListenAndServeTLS(":"+strconv.Itoa(wh.Port), certPath, keyPath, nil)
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
		body, _ := ioutil.ReadAll(req.Body)
		jsonResponse, _ := json.Marshal(parseAndResolveInjectionDemand(body, wh))
		w.Write(jsonResponse)
	}
}
