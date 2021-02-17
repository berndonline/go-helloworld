package main

import (
  "strings"
  "time"
  "strconv"
	"net/http"
  "github.com/prometheus/client_golang/prometheus"
  "github.com/gorilla/mux"
)

var (
	appName = string("helloworld")
	version = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "version",
		Help: "Version information about this binary",
		ConstLabels: map[string]string{"appName": appName},
	})
	httpRequestDuration = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name: "http_request_duration_seconds",
		Help: "Duration of all HTTP requests",
		Buckets: prometheus.LinearBuckets(0.01, 0.05, 10),
	  },
		[]string{"method", "code", "path"})
	httpRequestsResponseTime = prometheus.NewSummary(prometheus.SummaryOpts{
		Namespace: "http",
		Name:      "response_time_seconds",
		Help:      "Request response times",
  })
	httpRequestsTotal = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "http_requests_total",
		Help: "How many HTTP requests processed, partitioned by status code, method and HTTP path.",
		},
		[]string{"code", "method", "path"})
)

func prometheusMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		delegate := &responseWriterDelegator{ResponseWriter: w}
		rw := delegate

		next.ServeHTTP(rw, r) // call original

		route := mux.CurrentRoute(r)
		path, _ := route.GetPathTemplate()

		code := strconv.Itoa(delegate.status)
		method := strings.ToLower(r.Method)

		httpRequestsResponseTime.Observe(float64(time.Since(start).Seconds()))
		httpRequestsTotal.WithLabelValues(code,method,path).Inc()
		timer := prometheus.NewTimer(httpRequestDuration.WithLabelValues(method, code, path))
		timer.ObserveDuration()

	})
}

type responseWriterDelegator struct {
	http.ResponseWriter
	status      int
	written     int64
	wroteHeader bool
}

func (r *responseWriterDelegator) WriteHeader(code int) {
	r.status = code
	r.wroteHeader = true
	r.ResponseWriter.WriteHeader(code)
}

func (r *responseWriterDelegator) Write(b []byte) (int, error) {
	if !r.wroteHeader {
		r.WriteHeader(http.StatusOK)
	}
	n, err := r.ResponseWriter.Write(b)
	r.written += int64(n)
	return n, err
}
