package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
)

// default http response handler
func handler(w http.ResponseWriter, r *http.Request) {
	log.Print("helloworld: defaultHandler received a request - " + getIPAddress(r))
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, response+"\n"+os.Getenv("HOSTNAME"))
}

// http health handler to use with deployment liveness probe
func healthz(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	io.WriteString(w, `ok`)
}

// http ready handler to use with deployment readiness probe
func readyz(w http.ResponseWriter, r *http.Request) {
	if mongodb != false {
		// mongodb connectivity to display status content
		readyz, err := dao.Readyz()
		if err != nil {
			log.Print("helloworld: readiness - mongodb not ok")
			w.WriteHeader(http.StatusInternalServerError)
			io.WriteString(w, `notok`)
			return
		}
		// check if status match expected response
		readyzJson, _ := json.Marshal(readyz)
		response := string(readyzJson)
		expected := `[{"id":"606077e5e1e6bd09812a1098","status":"ok"}]`
		if response != expected {
			log.Print("helloworld: readiness - status not ok")
			w.WriteHeader(http.StatusInternalServerError)
			io.WriteString(w, `notok`)
			return
		}
		w.WriteHeader(http.StatusOK)
		io.WriteString(w, `ok`)

	} else {
		w.WriteHeader(http.StatusOK)
		io.WriteString(w, `ok`)
	}
}
