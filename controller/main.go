package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
)

func main() {
	wh := newWebHook("ssm", 8443, 777)

	cert, key := createSelfSignedCert(&wh)
	injectCAInMutatingWebhook(&wh, cert)
	certPath, keyPath := writeCertsToHomeFolder(cert, key)

	http.HandleFunc("/", wh.server)
	log.Print("Listening on port ", wh.Port)
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

func writeCertsToHomeFolder(cert []byte, key []byte) (string, string) {
	userHomeDir, err := os.UserHomeDir()
	if err != nil {
		log.Fatal(err)
	}

	certPath := userHomeDir + "/tls.crt"
	keyPath := userHomeDir + "/tls.key"

	err = ioutil.WriteFile(certPath, cert, 0644)
	if err != nil {
		log.Fatal(err)
	}
	err = ioutil.WriteFile(keyPath, key, 0644)
	if err != nil {
		log.Fatal(err)
	}

	return certPath, keyPath
}
