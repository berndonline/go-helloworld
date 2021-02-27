package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func Test_InstrumentHandler(t *testing.T) {

	r := prometheus.NewRegistry()
	r.MustRegister(httpRequestDuration)
	r.MustRegister(httpRequestsTotal)
	r.MustRegister(httpRequestsResponseTime)
	r.MustRegister(httpRequestSizeBytes)
	r.MustRegister(httpResponseSizeBytes)
	r.MustRegister(version)

	router := mux.NewRouter()
	router.Path("/metrics").Handler(promhttp.HandlerFor(r, promhttp.HandlerOpts{}))
	router.HandleFunc("/", handler)
	router.Use(InstrumentHandler)

	rr := httptest.NewRecorder()

	req1, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Error(err)
	}
	req2, err := http.NewRequest("GET", "/metrics", nil)
	if err != nil {
		t.Error(err)
	}

	router.ServeHTTP(rr, req1)
	router.ServeHTTP(rr, req2)
	body := rr.Body.String()

	t.Run("Check http_request_duration_seconds", func(t *testing.T) {
		if !strings.Contains(body, "http_request_duration_seconds") {
			t.Errorf("body does not contain request total entry '%s'", "http_request_duration_seconds")
		}
	})
	t.Run("Check http_requests_total", func(t *testing.T) {
		if !strings.Contains(body, "http_requests_total") {
			t.Errorf("body does not contain request duration entry '%s'", "http_requests_total")
		}
	})
	t.Run("Check http_request_size_bytes", func(t *testing.T) {
		if !strings.Contains(body, "http_request_size_bytes") {
			t.Errorf("body does not contain request total entry '%s'", "http_request_size_bytes")
		}
	})
	t.Run("Check http_response_size_bytes", func(t *testing.T) {
		if !strings.Contains(body, "http_response_size_bytes") {
			t.Errorf("body does not contain request duration entry '%s'", "http_response_size_bytes")
		}
	})
	t.Run("Check version", func(t *testing.T) {
		if !strings.Contains(body, "version") {
			t.Errorf("body does not contain request duration entry '%s'", "versions")
		}
	})
}
