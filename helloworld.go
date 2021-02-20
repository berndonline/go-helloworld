package main

import (
	"crypto/subtle"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"io"
	"log"
	"net/http"
	"os"
)

var (
	username    = os.Getenv("USERNAME")
	password    = os.Getenv("PASSWORD")
	response    = os.Getenv("RESPONSE")
	port        = os.Getenv("PORT")
	metricsPort = os.Getenv("METRICSPORT")
)

func init() {
	if username == "" {
		username = "admin"
	}
	if password == "" {
		password = "password"
	}
	if response == "" {
		response = "Hello, World - REST API!"
	}
	if port == "" {
		port = "8080"
	}
	if metricsPort == "" {
		metricsPort = "9100"
	}
}

func handler(w http.ResponseWriter, r *http.Request) {
	log.Print("helloworld: defaultHandler received a request")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, response+"\n"+os.Getenv("HOSTNAME"))
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	io.WriteString(w, `{"alive": true}`)
}

func BasicAuth(handler http.HandlerFunc, realm string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, pass, ok := r.BasicAuth()
		if !ok || subtle.ConstantTimeCompare([]byte(user),
			[]byte(username)) != 1 || subtle.ConstantTimeCompare([]byte(pass),
			[]byte(password)) != 1 {
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
	r.MustRegister(httpRequestSizeBytes)
	r.MustRegister(httpResponseSizeBytes)
	r.MustRegister(version)

	log.Print("helloworld: is starting...")

	routerInternal := mux.NewRouter().StrictSlash(true)
	routerInternal.Path("/metrics").Handler(promhttp.HandlerFor(r, promhttp.HandlerOpts{}))
	router := mux.NewRouter().StrictSlash(true)
	router.Use(InstrumentHandler)
	router.HandleFunc("/", handler)
	router.HandleFunc("/healthz", healthHandler)

	var api = router.PathPrefix("/api").Subrouter()
	api.NotFoundHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	})
	var v1 = api.PathPrefix("/v1").Subrouter()
	v1.HandleFunc("/content", BasicAuth(getIndexContent, "Please enter your username and password")).Methods("GET")
	v1.HandleFunc("/content", BasicAuth(createContent, "Please enter your username and password")).Methods("POST")
	v1.HandleFunc("/content/{id}", BasicAuth(getSingleContent, "Please enter your username and password")).Methods("GET")
	v1.HandleFunc("/content/{id}", BasicAuth(updateContent, "Please enter your username and password")).Methods("PUT")
	v1.HandleFunc("/content/{id}", BasicAuth(deleteContent, "Please enter your username and password")).Methods("DELETE")
	v1.NotFoundHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	})

	go func() {
		log.Printf("helloworld: metrics listening on port %s", metricsPort)
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
