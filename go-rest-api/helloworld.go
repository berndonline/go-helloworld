package main

import (
	"crypto/subtle"
	"net/http"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus"
  "github.com/prometheus/client_golang/prometheus/promhttp"
	"log"
	"os"
	"io"
)

const (
	ADMIN_USER     = "admin"
	ADMIN_PASSWORD = "password"
)

func handler(w http.ResponseWriter, r *http.Request) {
	log.Print("helloworld: defaultHandler received a request")
	response := os.Getenv("RESPONSE")
	if response == "" {
		response = "Hello, World - REST API!"
	}
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, response+"\n"+os.Getenv("HOSTNAME"))
}

func handlerHealth(w http.ResponseWriter, r *http.Request) {
  w.Header().Set("Content-Type", "application/json")
  w.WriteHeader(http.StatusOK)

  io.WriteString(w, `{"alive": true}`)
}

func BasicAuth(handler http.HandlerFunc, realm string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, pass, ok := r.BasicAuth()
		if !ok || subtle.ConstantTimeCompare([]byte(user),
			[]byte(ADMIN_USER)) != 1 || subtle.ConstantTimeCompare([]byte(pass),
			[]byte(ADMIN_PASSWORD)) != 1 {
			w.Header().Set("WWW-Authenticate", `Basic realm="`+realm+`"`)
			w.WriteHeader(401)
			w.Write([]byte("You are Unauthorized to access the application.\n"))
			log.Print("helloworld-api: authentication failed")
			return
		}
		handler(w, r)
	}
}

func main() {
	version.Set(0.1)

	r := prometheus.NewRegistry()
	r.MustRegister(httpRequestDuration)
  r.MustRegister(httpRequestsTotal)
	r.MustRegister(httpRequestsResponseTime)
	r.MustRegister(version)

	log.Print("helloworld: is starting...")
	routerInternal := mux.NewRouter().StrictSlash(true)
	routerInternal.Path("/metrics").Handler(promhttp.HandlerFor(r, promhttp.HandlerOpts{}))
	router := mux.NewRouter().StrictSlash(true)
	router.Use(prometheusMiddleware)
	router.HandleFunc("/", handler)
	router.HandleFunc("/healthz", handlerHealth)
	router.HandleFunc("/api/v1/content", BasicAuth(getAllContent, "Please enter your username and password")).Methods("GET")
	router.HandleFunc("/api/v1/content", BasicAuth(createContent, "Please enter your username and password")).Methods("POST")
	router.HandleFunc("/api/v1/content/{id}", BasicAuth(getOneContent, "Please enter your username and password")).Methods("GET")
	router.HandleFunc("/api/v1/content/{id}", BasicAuth(updateContent, "Please enter your username and password")).Methods("PUT")
	router.HandleFunc("/api/v1/content/{id}", BasicAuth(deleteContent, "Please enter your username and password")).Methods("DELETE")
	port := os.Getenv("PORT")
  metricsPort := os.Getenv("METRICSPORT")

	if port == "" {
		port = "8080"
	}
	if metricsPort == "" {
		metricsPort = "9100"
	}

	go func() {
		log.Printf("metrics: listening on port %s", metricsPort)
		err := http.ListenAndServe(fmt.Sprintf(":%s", metricsPort), routerInternal)
		if err != nil {
			log.Fatal("error starting metrics http server : ", err)
			return
		}
	}()

	log.Printf("helloworld: listening on port %s", port)
	err := http.ListenAndServe(fmt.Sprintf(":%s", port), router)
	if err != nil {
		log.Fatal("error starting http server : ", err)
		return
	}
}
