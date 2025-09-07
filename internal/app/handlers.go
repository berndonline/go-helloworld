package app

import (
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
    w.WriteHeader(http.StatusOK)
    io.WriteString(w, `ok`)
}
