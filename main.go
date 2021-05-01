package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
)

func main() {
	wh := newWebHook("ssm", 8443, 777)

	cert, key := wh.createCert()

	ioutil.WriteFile("cert.key", key, 0644)
	ioutil.WriteFile("cert.pem", cert, 0644)
	wh.alterMutatingWebhook(cert)

	http.HandleFunc("/", wh.server)

	log.Print("Listening on port ", wh.Port)
	err := http.ListenAndServeTLS(":"+strconv.Itoa(wh.Port), "cert.pem", "cert.key", nil)
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
		jsonResponse, _ := json.Marshal(injectSidecar(body, wh))
		w.Write(jsonResponse)
	}
}
