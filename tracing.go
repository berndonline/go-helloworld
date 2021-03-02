package main

import (
	opentracing "github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	"net/http"
  "github.com/gorilla/mux"
  "strings"
)

func TracingHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		delegate := &responseWriterDelegator{ResponseWriter: w}
		rw := delegate
		next.ServeHTTP(rw, r)

		route := mux.CurrentRoute(r)
		path, _ := route.GetPathTemplate()
		method := strings.ToLower(r.Method)
		code := uint16(delegate.status)

		tracer := opentracing.GlobalTracer()
		span := tracer.StartSpan(path)
		defer span.Finish()
		ext.HTTPMethod.Set(span, method)
		ext.HTTPStatusCode.Set(span, code)
		ext.PeerHostIPv4.SetString(span, getIPAddress(r))
		tracer.Inject(span.Context(), opentracing.HTTPHeaders, opentracing.HTTPHeadersCarrier(r.Header))
	})
}

func TracingExtract(r *http.Request, name string) {
	tracer := opentracing.GlobalTracer()
	spanCtx, _ := tracer.Extract(opentracing.HTTPHeaders, opentracing.HTTPHeadersCarrier(r.Header))
	span := tracer.StartSpan(name, ext.RPCServerOption(spanCtx))
	defer span.Finish()
}
