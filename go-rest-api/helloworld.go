package main

import (
	"crypto/subtle"
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"io/ioutil"
	"log"
	"net/http"
	"os"
  "strings"
	"time"
)

const (
	ADMIN_USER     = "admin"
	ADMIN_PASSWORD = "password"
)

type api struct {
	ID          string `json:"ID"`
	Name        string `json:"Name"`
}

type allContent []api

var contents = allContent{
	{
		ID:          "1",
		Name:        "Content 1",
	},
	{
		ID:          "2",
		Name:        "Content 2",
	},
}

var (
	appVersion string
	version = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "version",
		Help: "Version information about this binary",
		ConstLabels: map[string]string{
			"version": appVersion,
		},
	})
	httpRequestDuration = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name: "http_request_duration_seconds",
		Help: "Duration of all HTTP requests",
		Buckets: prometheus.LinearBuckets(0.01, 0.05, 10),
	}, []string{"path", "method"})
	httpRequestsResponseTime = prometheus.NewSummary(prometheus.SummaryOpts{
		Namespace: "http",
		Name:      "response_time_seconds",
		Help:      "Request response times",
  })
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

func createContent(w http.ResponseWriter, r *http.Request) {
	var newContent api
	reqBody, err := ioutil.ReadAll(r.Body)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		log.Print("helloworld-api: failed createContent")
	}
	json.Unmarshal(reqBody, &newContent)
	contents = append(contents, newContent)
	log.Print("helloworld-api: createContent received a request")
	respondWithJson(w, http.StatusOK, newContent)
}

func getOneContent(w http.ResponseWriter, r *http.Request) {
	contentID := mux.Vars(r)["id"]
	for _, singleContent := range contents {
    if singleContent.ID == contentID {
			log.Print("helloworld-api: getOneContent received a request")
			respondWithJson(w, http.StatusOK, singleContent)
      return
		}
	}
	log.Print("helloworld-api: invalid getOneContent")
	respondWithError(w, http.StatusNotFound, "Invalid ID")
}

func getAllContent(w http.ResponseWriter, r *http.Request) {
	log.Print("helloworld-api: getAllContent received a request")
	respondWithJson(w, http.StatusOK, contents)
}

func updateContent(w http.ResponseWriter, r *http.Request) {
	contentID := mux.Vars(r)["id"]
	var updatedContent api
	reqBody, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Print("helloworld-api: failed updateContent")
		respondWithError(w, http.StatusInternalServerError, err.Error())
	}
	json.Unmarshal(reqBody, &updatedContent)

	for i, singleContent := range contents {
		if singleContent.ID == contentID {
			singleContent.Name = updatedContent.Name
			contents = append(contents[:i], singleContent)
			log.Print("helloworld-api: updateContent received a request")
			respondWithJson(w, http.StatusOK, singleContent)
		}
	}
}

func deleteContent(w http.ResponseWriter, r *http.Request) {
	contentID := mux.Vars(r)["id"]

	for i, singleContent := range contents {
		if singleContent.ID == contentID {
			contents = append(contents[:i], contents[i+1:]...)
			log.Print("helloworld-api: deleteContent received a request")
			respondWithJson(w, http.StatusOK, "The content with has been deleted successfully")
		}
	}
}

func respondWithError(w http.ResponseWriter, code int, msg string) {
	respondWithJson(w, code, map[string]string{"error": msg})
}

func respondWithJson(w http.ResponseWriter, code int, payload interface{}) {
	response, _ := json.Marshal(payload)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(response)
}

func prometheusMiddleware(next http.Handler) http.Handler {
  return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
    route := mux.CurrentRoute(r)
    path, _ := route.GetPathTemplate()
		method := sanitizeMethod(r.Method)
    timer := prometheus.NewTimer(httpRequestDuration.WithLabelValues(path, method))

    next.ServeHTTP(w, r)

    timer.ObserveDuration()
		httpRequestsResponseTime.Observe(float64(time.Since(start).Seconds()))
	})
}

func sanitizeMethod(m string) string {
	return strings.ToLower(m)
}

func main() {
	version.Set(1)

	r := prometheus.NewRegistry()
	r.MustRegister(httpRequestDuration)
	r.MustRegister(httpRequestsResponseTime)
	r.MustRegister(version)

	log.Print("helloworld: is starting...")
	router := mux.NewRouter().StrictSlash(true)
	router.Use(prometheusMiddleware)
	routerInternal := mux.NewRouter().StrictSlash(true)
	routerInternal.Path("/metrics").Handler(promhttp.HandlerFor(r, promhttp.HandlerOpts{}))
	router.HandleFunc("/", handler)
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
