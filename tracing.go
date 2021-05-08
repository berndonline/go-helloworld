package main

import (
	"github.com/gorilla/mux"
	opentracing "github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	"github.com/uber/jaeger-client-go"
	jaegercfg "github.com/uber/jaeger-client-go/config"
	// jaegerlog "github.com/uber/jaeger-client-go/log"
	"github.com/uber/jaeger-client-go/zipkin"
	// "github.com/uber/jaeger-lib/metrics"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
)

// opentracing service configuration
func initTracer(service string) (opentracing.Tracer, io.Closer) {
	cfg := jaegercfg.Configuration{
		ServiceName: service,
		Sampler: &jaegercfg.SamplerConfig{
			Type:  jaeger.SamplerTypeConst,
			Param: 1,
		},
		Reporter: &jaegercfg.ReporterConfig{
			LogSpans: true,
		},
	}
	// jLogger := jaegerlog.StdLogger
	// jMetricsFactory := metrics.NullFactory
	
	// https://github.com/jaegertracing/jaeger-client-go/blob/master/zipkin/README.md
	zipkinPropagator := zipkin.NewZipkinB3HTTPHeaderPropagator()

	tracer, closer, err := cfg.NewTracer(
		// jaegercfg.Logger(jLogger),
		// jaegercfg.Metrics(jMetricsFactory),

		// https://github.com/jaegertracing/jaeger-client-go/blob/master/zipkin/README.md
		jaegercfg.Injector(opentracing.HTTPHeaders, zipkinPropagator),
		jaegercfg.Extractor(opentracing.HTTPHeaders, zipkinPropagator),
		jaegercfg.ZipkinSharedRPCSpan(true),
	)
	if err != nil {
		log.Fatal(service+": cannot initialize Jaeger Tracer", err)
	}

	return tracer, closer
}

// tracing handler to start root span
func tracingHandler(handler http.HandlerFunc) http.HandlerFunc {
	incomingHeaders := []string{
		"x-request-id",
		"x-b3-traceid",
		"x-b3-spanid",
		"x-b3-parentspanid",
		"x-b3-sampled",
		"x-b3-flags",
		"x-ot-span-context",
	}
	return func(w http.ResponseWriter, r *http.Request) {
		route := mux.CurrentRoute(r)
		path, _ := route.GetPathTemplate()
		method := strings.ToLower(r.Method)

		tracer := opentracing.GlobalTracer()
		span := StartSpanFromRequest(path, tracer, r)

		defer span.Finish()
		ext.HTTPMethod.Set(span, method)
		ext.PeerHostIPv4.SetString(span, getIPAddress(r))
		span.SetTag("hostname", os.Getenv("HOSTNAME"))

		for _, th := range incomingHeaders {
			span.SetTag(th, r.Header.Get(th))
			w.Header().Set(th, r.Header.Get(th))
		}

		traceID, _ := span.Context().(jaeger.SpanContext)
		log.Print("rootSpan:", traceID)
		Inject(span, r)
		handler(w, r)
	}
}

func StartSpanFromRequest(spanName string, tracer opentracing.Tracer, r *http.Request) opentracing.Span {
	spanCtx, _ := Extract(tracer, r)
	return tracer.StartSpan(spanName, opentracing.ChildOf(spanCtx))
}

func Extract(tracer opentracing.Tracer, r *http.Request) (opentracing.SpanContext, error) {
	return tracer.Extract(
		opentracing.HTTPHeaders,
		opentracing.HTTPHeadersCarrier(r.Header))
}

func Inject(span opentracing.Span, r *http.Request) error {
	return span.Tracer().Inject(
		span.Context(),
		opentracing.HTTPHeaders,
		opentracing.HTTPHeadersCarrier(r.Header))
}
