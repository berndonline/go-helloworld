package app

import (
    "fmt"
    "github.com/gorilla/handlers"
    "github.com/gorilla/mux"
    opentracing "github.com/opentracing/opentracing-go"
    "github.com/prometheus/client_golang/prometheus"
    "github.com/prometheus/client_golang/prometheus/promhttp"
    "log"
    "net/http"
    "os"
)

// all constant variables
const (
    contentRoot = "/content"
    contentID   = "/content/{id}"
)

// default variables and data access object
var (
    // open tracing service name
    serviceName = os.Getenv("SERVICENAME")
    // http handler response
    response = os.Getenv("RESPONSE")
    // http ports
    httpPort    = os.Getenv("PORT")
    metricsPort = os.Getenv("METRICSPORT")
)

// init function to populate runtime variables and defaults
func init() {
    // set default open tracing service name
    if serviceName == "" {
        serviceName = "helloworld"
    }
    // set default response
    if response == "" {
        response = "Hello, World - REST API!"
    }
    // set default http port
    if httpPort == "" {
        httpPort = "8080"
    }
    // set default metrics port
    if metricsPort == "" {
        metricsPort = "9100"
    }
}

func Run() {
    // initialize tracer and servicename
    tracer, closer := initTracer(serviceName)
    opentracing.SetGlobalTracer(tracer)
    defer closer.Close()
    // application version displayed in prometheus
    version.Set(0.1)
    cleanupPublisher := configureContentPublisher()
    defer cleanupPublisher()
    log.Print("helloworld: is starting...")
    // log the running UID/GID for visibility in non-root environments
    log.Printf("helloworld: running as uid=%d gid=%d", os.Getuid(), os.Getgid())
    // prometheus registry filtering the exported metrics
    registry := prometheus.NewRegistry()
    registry.MustRegister(httpRequestDuration)
    registry.MustRegister(httpRequestsTotal)
    registry.MustRegister(httpRequestsResponseTime)
    registry.MustRegister(httpRequestSizeBytes)
    registry.MustRegister(httpResponseSizeBytes)
    registry.MustRegister(version)
    // http request router for /metrics path to be not exposed through main root path
    routerInternal := mux.NewRouter()
    routerInternal.Path("/metrics").Handler(promhttp.HandlerFor(registry, promhttp.HandlerOpts{}))
    // exposing healthz and readyz handlers via internal router
    routerInternal.HandleFunc("/healthz", healthz)
    routerInternal.HandleFunc("/readyz", readyz)
    // main request router for rest-api
    router := mux.NewRouter()
    // prometheus middleware handlers to capture application metrics
    router.Use(InstrumentHandler)
    // default response handler
    router.HandleFunc("/", handler)
    // static file http handler (served from /static inside container)
    router.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("/static/"))))
    // reverse proxy server
    var proxy = router.PathPrefix("/proxy").Subrouter()
    for _, conf := range configuration {
        proxyConf := generateProxy(conf)
        proxy.Handle(conf.Path, tracingHandler(func(w http.ResponseWriter, r *http.Request) {
            proxyConf.ServeHTTP(w, r)
        }))
    }
    proxy.NotFoundHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.WriteHeader(http.StatusNotFound)
    })
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
    // function to start internal request router on port TCP 9100 (default)
    go func() {
        log.Printf("helloworld: metrics listening on port %s", metricsPort)
        if err := http.ListenAndServe(fmt.Sprintf(":%s", metricsPort), routerInternal); err != nil {
            log.Fatal("error starting metrics http server : ", err)
            return
        }
    }()
    // enable mux request logging handler for external request router
    loggingRouter := handlers.CombinedLoggingHandler(os.Stdout, router)
    // main request router to expose default handlers and api versions on port TCP 8080 (default)
    log.Printf("helloworld: listening on port %s", httpPort)
    if err := http.ListenAndServe(fmt.Sprintf(":%s", httpPort), loggingRouter); err != nil {
        log.Fatal("error starting http server : ", err)
        return
    }
}
