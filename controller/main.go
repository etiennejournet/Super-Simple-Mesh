package main

import (
	"encoding/json"
	"io/ioutil"
  "os"
	"log"
	"net/http"
	"strconv"
)

func main() {
	wh := newWebHook("ssm", 8443, 777)

	cert, key := wh.createCert()

  userHomeDir, err := os.UserHomeDir()
  if err != nil {
    log.Fatal(err)
  }
	err = ioutil.WriteFile(userHomeDir+"/cert.key", key, 0644)
  if err != nil {
    log.Fatal(err)
  }
	err = ioutil.WriteFile(userHomeDir+"/cert.pem", cert, 0644)
  if err != nil {
    log.Fatal(err)
  }
	wh.alterMutatingWebhook(cert)

	http.HandleFunc("/", wh.server)

	log.Print("Listening on port ", wh.Port)
	err = http.ListenAndServeTLS(":"+strconv.Itoa(wh.Port), userHomeDir+"/cert.pem", userHomeDir+"/cert.key", nil)
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
