package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
)

var (
	InfoLogger  *log.Logger
	ErrorLogger *log.Logger
)

func init() {
	InfoLogger = log.New(os.Stdout, "INFO: ", log.Ldate|log.Ltime|log.Lshortfile)
	ErrorLogger = log.New(os.Stdout, "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile)
}

func main() {
	wh := newWebHook("ssm", 8443, 777)

	cert, key := createSelfSignedCert(&wh)
	injectCAInMutatingWebhook(&wh, cert)
	certPath, keyPath := writeCertsToHomeFolder(cert, key)

	http.HandleFunc("/", wh.server)
	InfoLogger.Print("WebHook listening on port ", wh.Port)
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
