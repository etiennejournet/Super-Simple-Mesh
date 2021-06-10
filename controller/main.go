package main

import (
  log "github.com/sirupsen/logrus"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strconv"
)

func main() {
	wh := newWebHook("ssm", 8443, 777)

	cert, key := createSelfSignedCert(&wh)
	injectCAInMutatingWebhook(&wh, cert)
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
