package main

import (
	"fmt"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	opentracing "github.com/opentracing/opentracing-go"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	mgo "gopkg.in/mgo.v2"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
)

// all constant variables
const (
	contentRoot = "/content/"
	contentID   = "/content/{id}"
)

// default variables and data access object
var (
	serviceName string
	response    = os.Getenv("RESPONSE")
	httpPort    = os.Getenv("PORT")
	metricsPort = os.Getenv("METRICSPORT")
	mongodb, _  = strconv.ParseBool(os.Getenv("MONGODB"))
	server      = os.Getenv("SERVER")
	database    = os.Getenv("DATABASE")
	dao         = contentsDAO{}
	db          *mgo.Database
)

// init function to popluate variables or initiate mongodb connection if enabled
func init() {
	if mongodb != false {
		if serviceName == "" {
			serviceName = "helloworld-mongodb"
		}
		if server == "" {
			server = "mongodb"
		}
		if database == "" {
			database = "contents_db"
		}
		dao.Server = server
		dao.Database = database
		dao.Connect()
	}
	if serviceName == "" {
		serviceName = "helloworld"
	}
	if response == "" {
		response = "Hello, World - REST API!"
	}
	if httpPort == "" {
		httpPort = "8080"
	}
	if metricsPort == "" {
		metricsPort = "9100"
	}
}

// default http response handler
func handler(w http.ResponseWriter, r *http.Request) {
	log.Print("helloworld: defaultHandler received a request - " + getIPAddress(r))
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, response+"\n"+os.Getenv("HOSTNAME"))
}

// http health handler to use with deployments readiness probe
func healthz(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	io.WriteString(w, `{"alive": true}`)
}

// function to get IP address from http header - example below for custom CloudFlare X-Forwarder-For header
func getIPAddress(r *http.Request) string {
	ip := strings.TrimSpace(r.Header.Get("CF-Connecting-IP"))
	return ip
}

func main() {
	// initialize tracer and servicename
	tracer, closer := initTracer(serviceName)
	opentracing.SetGlobalTracer(tracer)
	defer closer.Close()
	// application version displayed in prometheus
	version.Set(0.1)
	log.Print("helloworld: is starting...")
	// prometheus registry filtering the exported metrics
	registry := prometheus.NewRegistry()
	registry.MustRegister(httpRequestDuration)
	registry.MustRegister(httpRequestsTotal)
	registry.MustRegister(httpRequestsResponseTime)
	registry.MustRegister(httpRequestSizeBytes)
	registry.MustRegister(httpResponseSizeBytes)
	registry.MustRegister(version)
	// http request router for /metrics path to be not exposed through main root path
	routerInternal := mux.NewRouter().StrictSlash(true)
	routerInternal.Path("/metrics").Handler(promhttp.HandlerFor(registry, promhttp.HandlerOpts{}))
	// main request router for rest-api
	router := mux.NewRouter().StrictSlash(true)
	// prometheus middleware handlers to capture application metrics
	router.Use(InstrumentHandler)
	// default response and health handler
	router.HandleFunc("/", handler)
	router.HandleFunc("/healthz", healthz)
	// enable reverse proxy config
	for _, conf := range configuration {
		proxy := generateProxy(conf)
		router.HandleFunc(conf.Path, func(w http.ResponseWriter, r *http.Request) {
			proxy.ServeHTTP(w, r)
		})
	}
	// api root path defined as subrouter
	var api = router.PathPrefix("/api").Subrouter()
	api.NotFoundHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	})
	// version 1 of the api using basicAuth
	var v1 = api.PathPrefix("/v1").Subrouter()
	v1.Handle(contentRoot, tracingHandler(basicAuth(getIndexContent))).Methods("GET")
	v1.Handle(contentRoot, tracingHandler(basicAuth(createContent))).Methods("POST")
	v1.Handle(contentID, tracingHandler(basicAuth(getSingleContent))).Methods("GET")
	v1.Handle(contentID, tracingHandler(basicAuth(updateContent))).Methods("PUT")
	v1.Handle(contentID, tracingHandler(basicAuth(deleteContent))).Methods("DELETE")
	v1.NotFoundHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	})
	// version 2 of the api using json web token (JWT) authentication
	var v2 = api.PathPrefix("/v2").Subrouter()
	v2.HandleFunc("/login", jwtLogin).Methods("POST")
	v2.HandleFunc("/logout", jwtLogout).Methods("POST")
	v2.HandleFunc("/refresh", jwtAuth(jwtRefresh)).Methods("POST")
	v2.Handle(contentRoot, tracingHandler(jwtAuth(getIndexContent))).Methods("GET")
	v2.Handle(contentRoot, tracingHandler(jwtAuth(createContent))).Methods("POST")
	v2.Handle(contentID, tracingHandler(jwtAuth(getSingleContent))).Methods("GET")
	v2.Handle(contentID, tracingHandler(jwtAuth(updateContent))).Methods("PUT")
	v2.Handle(contentID, tracingHandler(jwtAuth(deleteContent))).Methods("DELETE")
	v2.NotFoundHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	})
	// function to start /metrics request router on port TCP 9100 (default)
	go func() {
		log.Printf("helloworld: metrics listening on port %s", metricsPort)
		if err := http.ListenAndServe(fmt.Sprintf(":%s", metricsPort), routerInternal); err != nil {
			log.Fatal("error starting metrics http server : ", err)
			return
		}
	}()
	// enable mux request logging handler for default router
	loggingRouter := handlers.CombinedLoggingHandler(os.Stdout, router)
	// main request router to expose default handlers and api versions on port TCP 8080 (default)
	log.Printf("helloworld: listening on port %s", httpPort)
	if err := http.ListenAndServe(fmt.Sprintf(":%s", httpPort), loggingRouter); err != nil {
		log.Fatal("error starting http server : ", err)
		return
	}
}
