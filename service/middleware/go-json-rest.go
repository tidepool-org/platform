package middleware

import (
	"fmt"
	"net/http"
	"sync"

	"github.com/ant0ine/go-json-rest/rest"
	otelcontrib "go.opentelemetry.io/contrib"
	"go.opentelemetry.io/otel"

	otelglobal "go.opentelemetry.io/otel/api/global"
	oteltrace "go.opentelemetry.io/otel/api/trace"
	"go.opentelemetry.io/otel/semconv"
)

type config struct {
	TracerProvider oteltrace.TracerProvider
	Propagators    otel.TextMapPropagator
}

// WithPropagators specifies propagators to use for extracting
// information from the HTTP requests. If none are specified, global
// ones will be used.
func WithPropagators(propagators otel.TextMapPropagator) Option {
	return func(cfg *config) {
		cfg.Propagators = propagators
	}
}

// WithTracerProvider specifies a tracer provider to use for creating a tracer.
// If none is specified, the global provider is used.
func WithTracerProvider(provider oteltrace.TracerProvider) Option {
	return func(cfg *config) {
		cfg.TracerProvider = provider
	}
}

//Option is a an option for a open telemetry configuration
type Option func(*config)

const (
	tracerName = "github.com/tidepool-org/platform/service/middleware/go-json-rest"
)

//OtelTracing implements middleware to support OpenTelemetry tracing
type OtelTracing struct {
	service     string
	tracer      oteltrace.Tracer
	propagators otel.TextMapPropagator
}

// NewOtelTracing creates middleware for go-json-rest that supports OpenTelemetry
func NewOtelTracing(service string, opts ...Option) (*OtelTracing, error) {

	cfg := config{}
	for _, opt := range opts {
		opt(&cfg)
	}
	if cfg.TracerProvider == nil {
		cfg.TracerProvider = otelglobal.TracerProvider()
	}
	if cfg.TracerProvider == nil {
		return nil, fmt.Errorf("no tracer provider configured")
	}
	tracer := cfg.TracerProvider.Tracer(
		tracerName,
		oteltrace.WithInstrumentationVersion(otelcontrib.SemVersion()),
	)
	if cfg.Propagators == nil {
		cfg.Propagators = otelglobal.TextMapPropagator()
	}

	return &OtelTracing{
		service,
		tracer,
		cfg.Propagators,
	}, nil
}

//MiddlewareFunc adds tracing to incoming requests
func (tw OtelTracing) MiddlewareFunc(handler rest.HandlerFunc) rest.HandlerFunc {
	return func(res rest.ResponseWriter, r *rest.Request) {

		ctx := tw.propagators.Extract(r.Request.Context(), r.Request.Header)
		spanName := r.Request.URL.Path
		routeStr := spanName
		opts := []oteltrace.SpanOption{
			oteltrace.WithAttributes(semconv.NetAttributesFromHTTPRequest("tcp", r.Request)...),
			oteltrace.WithAttributes(semconv.EndUserAttributesFromHTTPRequest(r.Request)...),
			oteltrace.WithAttributes(semconv.HTTPServerAttributesFromHTTPRequest(tw.service, routeStr, r.Request)...),
			oteltrace.WithSpanKind(oteltrace.SpanKindServer),
		}
		ctx, span := tw.tracer.Start(ctx, spanName, opts...)
		defer span.End()
		r2 := &rest.Request{
			Request:    r.WithContext(ctx),
			PathParams: r.PathParams,
			Env:        r.Env,
		}

		rrw := getRRW(res)
		defer putRRW(rrw)

		handler(rrw, r2)

		attrs := semconv.HTTPAttributesFromHTTPStatusCode(rrw.status)
		spanStatus, spanMessage := semconv.SpanStatusFromHTTPStatusCode(rrw.status)
		span.SetAttributes(attrs...)
		span.SetStatus(spanStatus, spanMessage)
	}
}

type recordingResponseWriter struct {
	writer  rest.ResponseWriter
	written bool
	status  int
}

var rrwPool = &sync.Pool{
	New: func() interface{} {
		return &recordingResponseWriter{}
	},
}

func (rrw *recordingResponseWriter) Header() http.Header {
	return rrw.writer.Header()
}

func (rrw *recordingResponseWriter) WriteJson(v interface{}) error {
	return rrw.writer.WriteJson(v)
}

func (rrw *recordingResponseWriter) EncodeJson(v interface{}) ([]byte, error) {
	b, e := rrw.writer.EncodeJson(v)
	return b, e
}

func (rrw *recordingResponseWriter) WriteHeader(statusCode int) {
	if !rrw.written {
		rrw.written = true
		rrw.status = statusCode
	}
	rrw.writer.WriteHeader(statusCode)
}

func getRRW(writer rest.ResponseWriter) *recordingResponseWriter {
	rrw := rrwPool.Get().(*recordingResponseWriter)
	rrw.written = false
	rrw.status = 0
	rrw.writer = writer
	return rrw
}

func putRRW(rrw *recordingResponseWriter) {
	rrw.writer = nil
	rrwPool.Put(rrw)
}
