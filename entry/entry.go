package rk

import (
	"context"
	"embed"
	"github.com/golang-jwt/jwt/v4"
	"github.com/prometheus/client_golang/prometheus"
	"go.opentelemetry.io/contrib"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/propagation"
	sdkresource "go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
	oteltrace "go.opentelemetry.io/otel/trace"
	"net/http"
)

type EntryConfigHeader struct {
	Kind       string `yaml:"kind"`
	ApiVersion string `yaml:"apiVersion"`
	Metadata   struct {
		Name    string `yaml:"name"`
		Version string `yaml:"version"`
		Domain  string `yaml:"domain"`
		Enabled bool   `yaml:"enabled"`
		Default bool   `yaml:"default"`
	} `yaml:"metadata"`
}

type EntryConfig interface {
	JSON() string

	YAML() string

	Header() *EntryConfigHeader

	Register() (Entry, error)
}

type Entry interface {
	Category() string

	Kind() string

	Name() string

	Config() EntryConfig

	Bootstrap(context.Context)

	Interrupt(context.Context)

	Monitor() *Monitor

	FS() *embed.FS

	Apis() []*BuiltinApi
}

// SignerJwt interface which must be implemented for JWT signer
type SignerJwt interface {
	Entry

	// SignJwt sign jwt.Token
	SignJwt(claim jwt.Claims) (string, error)

	// VerifyJwt verify jwt.Token
	VerifyJwt(token string) (*jwt.Token, error)

	// PubKey get public key
	PubKey() []byte

	// Algorithms supported algorithms
	Algorithms() []string
}

type Crypto interface {
	Entry

	Encrypt(plaintext []byte) ([]byte, error)

	Decrypt(plaintext []byte) ([]byte, error)
}

type BuiltinApi struct {
	Handler http.HandlerFunc
	Method  string
	Path    string
}

// ********** Monitor **********

func NewMonitorStd() *Monitor {
	return &Monitor{
		Prometheus: &Prometheus{
			Collectors: []prometheus.Collector{},
		},
		Otel: NewOtel(&NoopExporter{}, "", "", ""),
	}
}

type Monitor struct {
	Prometheus *Prometheus
	Otel       *Otel
}

// ********** Prometheus **********

type Prometheus struct {
	Collectors []prometheus.Collector
}

// ********** OTEL **********

func NewOtel(exporter sdktrace.SpanExporter, serviceName, serviceVersion, tracerName string) *Otel {
	res := &Otel{
		Exporter: exporter,
	}

	res.Processor = sdktrace.NewBatchSpanProcessor(exporter)
	res.Provider = sdktrace.NewTracerProvider(
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
		sdktrace.WithSpanProcessor(res.Processor),
		sdktrace.WithResource(
			sdkresource.NewWithAttributes(
				semconv.SchemaURL,
				attribute.String("service.name", serviceName),
				attribute.String("service.version", serviceVersion),
			)),
	)

	res.Tracer = res.Provider.Tracer(tracerName, oteltrace.WithInstrumentationVersion(contrib.SemVersion()))
	res.Propagator = propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{})

	return res
}

type Otel struct {
	Exporter   sdktrace.SpanExporter
	Processor  sdktrace.SpanProcessor
	Tracer     oteltrace.Tracer
	Provider   *sdktrace.TracerProvider
	Propagator propagation.TextMapPropagator
}

// NoopExporter noop
type NoopExporter struct{}

// ExportSpans handles export of SpanSnapshots by dropping them.
func (nsb *NoopExporter) ExportSpans(context.Context, []sdktrace.ReadOnlySpan) error {
	return nil
}

// Shutdown stops the exporter by doing nothing.
func (nsb *NoopExporter) Shutdown(context.Context) error {
	return nil
}
