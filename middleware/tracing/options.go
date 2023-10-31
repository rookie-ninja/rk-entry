// Copyright (c) 2021 rookie-ninja
//
// Use of this source code is governed by an Apache-style
// license that can be found in the LICENSE file.

// Package rkmidtrace is a middleware for recording tracing
package rkmidtrace

import (
	"context"
	"github.com/rookie-ninja/rk-entry/v2/entry"
	"github.com/rookie-ninja/rk-entry/v2/middleware"
	"github.com/rookie-ninja/rk-logger"
	"go.opentelemetry.io/contrib"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	otexporterotlp "go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	otexporterzipkin "go.opentelemetry.io/otel/exporters/zipkin"
	"go.opentelemetry.io/otel/propagation"
	sdkresource "go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.12.0"
	oteltrace "go.opentelemetry.io/otel/trace"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// ***************** OptionSet Interface *****************

// OptionSetInterface mainly for testing purpose
type OptionSetInterface interface {
	GetEntryName() string

	GetEntryType() string

	BeforeCtx(req *http.Request, isClient bool, attrs ...attribute.KeyValue) *BeforeCtx

	Before(*BeforeCtx)

	AfterCtx(resCode int, resMsg string, attrs ...attribute.KeyValue) *AfterCtx

	After(before *BeforeCtx, after *AfterCtx)

	GetTracer() oteltrace.Tracer

	GetProvider() *sdktrace.TracerProvider

	GetPropagator() propagation.TextMapPropagator

	ShouldIgnore(string) bool
}

// ***************** OptionSet Implementation *****************

// optionSet which is used for middleware implementation
type optionSet struct {
	entryName    string
	entryType    string
	exporter     sdktrace.SpanExporter
	processor    sdktrace.SpanProcessor
	provider     *sdktrace.TracerProvider
	propagator   propagation.TextMapPropagator
	tracer       oteltrace.Tracer
	pathToIgnore []string
	mock         OptionSetInterface
}

// NewOptionSet Create new optionSet with options.
func NewOptionSet(opts ...Option) OptionSetInterface {
	set := &optionSet{
		entryName:    "fake-entry",
		entryType:    "",
		pathToIgnore: []string{},
	}

	for i := range opts {
		opts[i](set)
	}

	if set.mock != nil {
		return set.mock
	}

	if set.exporter == nil {
		set.exporter = NewNoopExporter()
	}

	if set.processor == nil {
		set.processor = sdktrace.NewBatchSpanProcessor(set.exporter)
	}

	if set.provider == nil {
		res, _ := sdkresource.New(context.Background(),
			sdkresource.WithFromEnv(),
			sdkresource.WithProcess(),
			sdkresource.WithTelemetrySDK(),
			sdkresource.WithHost(),
			sdkresource.WithAttributes(
				semconv.ServiceNameKey.String(rkentry.GlobalAppCtx.GetAppInfoEntry().AppName),
				semconv.ServiceVersionKey.String(rkentry.GlobalAppCtx.GetAppInfoEntry().Version),
				attribute.String("service.entryName", set.entryName),
				attribute.String("service.entryType", set.entryType),
				semconv.TelemetrySDKLanguageGo,
			),
		)
		set.provider = sdktrace.NewTracerProvider(
			sdktrace.WithSampler(sdktrace.AlwaysSample()),
			sdktrace.WithSpanProcessor(set.processor),
			sdktrace.WithResource(res),
		)
	}

	set.tracer = set.provider.Tracer(set.entryName, oteltrace.WithInstrumentationVersion(contrib.SemVersion()))

	if set.propagator == nil {
		set.propagator = propagation.NewCompositeTextMapPropagator(
			propagation.TraceContext{},
			propagation.Baggage{})
	}

	return set
}

// GetEntryName returns entry name
func (set *optionSet) GetEntryName() string {
	return set.entryName
}

// GetEntryType returns entry type
func (set *optionSet) GetEntryType() string {
	return set.entryType
}

// GetTracer returns oteltrace.Tracer
func (set *optionSet) GetTracer() oteltrace.Tracer {
	return set.tracer
}

// GetProvider returns sdktrace.TracerProvider
func (set *optionSet) GetProvider() *sdktrace.TracerProvider {
	return set.provider
}

// GetPropagator returns propagation.TextMapPropagator
func (set *optionSet) GetPropagator() propagation.TextMapPropagator {
	return set.propagator
}

// BeforeCtx create beforeCtx based on http.Request
func (set *optionSet) BeforeCtx(req *http.Request, isClient bool, attrs ...attribute.KeyValue) *BeforeCtx {
	ctx := NewBeforeCtx()
	ctx.Input.IsClient = isClient
	ctx.Input.Attributes = append(ctx.Input.Attributes,
		attribute.String(rkmid.Domain.Key, rkmid.Domain.String))

	ctx.Input.Attributes = append(ctx.Input.Attributes, attrs...)

	if req != nil && req.URL != nil {
		ctx.Input.Attributes = append(ctx.Input.Attributes, semconv.NetAttributesFromHTTPRequest("tcp", req)...)
		ctx.Input.Attributes = append(ctx.Input.Attributes, semconv.EndUserAttributesFromHTTPRequest(req)...)
		ctx.Input.Attributes = append(ctx.Input.Attributes, semconv.HTTPServerAttributesFromHTTPRequest(
			rkentry.GlobalAppCtx.GetAppInfoEntry().AppName, req.URL.Path, req)...)
		ctx.Input.SpanName = req.URL.Path

		ctx.Input.RequestCtx = req.Context()
		ctx.Input.Carrier = propagation.HeaderCarrier(req.Header)
		ctx.Input.UrlPath = req.URL.Path
		// assign NewCtx for safety
		ctx.Output.NewCtx = req.Context()
	}

	return ctx
}

// Before should run before user handler
func (set *optionSet) Before(ctx *BeforeCtx) {
	if ctx == nil {
		return
	}

	if set.ShouldIgnore(ctx.Input.UrlPath) {
		ctx.Output.NewCtx = ctx.Input.RequestCtx
		return
	}

	opts := []oteltrace.SpanStartOption{
		oteltrace.WithAttributes(ctx.Input.Attributes...),
	}

	if ctx.Input.IsClient {
		opts = append(opts, oteltrace.WithSpanKind(oteltrace.SpanKindClient))
	} else {
		opts = append(opts, oteltrace.WithSpanKind(oteltrace.SpanKindServer))
	}

	// 1: extract tracing info from request header
	spanCtx := oteltrace.SpanContextFromContext(set.propagator.Extract(ctx.Input.RequestCtx, ctx.Input.Carrier))

	// 2: start new span
	ctx.Output.NewCtx, ctx.Output.Span = set.tracer.Start(
		oteltrace.ContextWithRemoteSpanContext(ctx.Input.RequestCtx, spanCtx),
		ctx.Input.SpanName, opts...)
}

// AfterCtx should be created before After()
func (set *optionSet) AfterCtx(resCode int, resMsg string, attrs ...attribute.KeyValue) *AfterCtx {
	ctx := NewAfterCtx()
	ctx.Input.ResCode = resCode
	ctx.Input.ResMsg = resMsg

	ctx.Input.Attributes = append(ctx.Input.Attributes, attrs...)

	return ctx
}

// After should run after user handler
func (set *optionSet) After(before *BeforeCtx, after *AfterCtx) {
	if before == nil || after == nil {
		return
	}

	if set.ShouldIgnore(before.Input.UrlPath) {
		return
	}

	if after.Input.ResCode >= 0 {
		before.Output.Span.SetAttributes(semconv.HTTPAttributesFromHTTPStatusCode(after.Input.ResCode)...)
	}

	code, _ := semconv.SpanStatusFromHTTPStatusCode(after.Input.ResCode)
	if code == codes.Unset {
		code = codes.Ok
	}

	before.Output.Span.SetStatus(code, after.Input.ResMsg)
	before.Output.Span.SetAttributes(after.Input.Attributes...)
	before.Output.Span.End()
}

// ShouldIgnore determine whether auth should be ignored based on path
func (set *optionSet) ShouldIgnore(path string) bool {
	for i := range set.pathToIgnore {
		if strings.HasPrefix(path, set.pathToIgnore[i]) {
			return true
		}
	}

	return rkmid.ShouldIgnoreGlobal(path)
}

// ***************** OptionSet Mock *****************

// NewOptionSetMock for testing purpose
func NewOptionSetMock(before *BeforeCtx, after *AfterCtx,
	tracer oteltrace.Tracer,
	provider *sdktrace.TracerProvider,
	propagator propagation.TextMapPropagator) OptionSetInterface {
	return &optionSetMock{
		before:     before,
		after:      after,
		tracer:     tracer,
		provider:   provider,
		propagator: propagator,
	}
}

type optionSetMock struct {
	before     *BeforeCtx
	after      *AfterCtx
	tracer     oteltrace.Tracer
	provider   *sdktrace.TracerProvider
	propagator propagation.TextMapPropagator
}

func (mock *optionSetMock) GetTracer() oteltrace.Tracer {
	return mock.tracer
}

func (mock *optionSetMock) GetProvider() *sdktrace.TracerProvider {
	return mock.provider
}

func (mock *optionSetMock) GetPropagator() propagation.TextMapPropagator {
	return mock.propagator
}

// GetEntryName returns entry name
func (mock *optionSetMock) GetEntryName() string {
	return "mock"
}

// GetEntryType returns entry type
func (mock *optionSetMock) GetEntryType() string {
	return "mock"
}

// BeforeCtx should be created before Before()
func (mock *optionSetMock) BeforeCtx(req *http.Request, isClient bool, attrs ...attribute.KeyValue) *BeforeCtx {
	return mock.before
}

// Before should run before user handler
func (mock *optionSetMock) Before(ctx *BeforeCtx) {
	return
}

// AfterCtx should be created before After()
func (mock *optionSetMock) AfterCtx(resCode int, resMsg string, attrs ...attribute.KeyValue) *AfterCtx {
	return mock.after
}

// After should run after user handler
func (mock *optionSetMock) After(before *BeforeCtx, after *AfterCtx) {
	return
}

// ShouldIgnore should run before user handler
func (mock *optionSetMock) ShouldIgnore(string) bool {
	return false
}

// ***************** Context *****************

// NewBeforeCtx create new BeforeCtx with fields initialized
func NewBeforeCtx() *BeforeCtx {
	ctx := &BeforeCtx{}
	ctx.Input.Attributes = make([]attribute.KeyValue, 0)

	return ctx
}

// NewAfterCtx create new AfterCtx with fields initialized
func NewAfterCtx() *AfterCtx {
	ctx := &AfterCtx{}
	ctx.Input.Attributes = make([]attribute.KeyValue, 0)
	return ctx
}

// BeforeCtx context for Before() function
type BeforeCtx struct {
	Input struct {
		UrlPath    string
		SpanName   string
		IsClient   bool
		Attributes []attribute.KeyValue
		RequestCtx context.Context
		Carrier    propagation.TextMapCarrier
	}
	Output struct {
		NewCtx context.Context
		Span   oteltrace.Span
	}
}

// AfterCtx context for After() function
type AfterCtx struct {
	Input struct {
		ResCode    int
		ResMsg     string
		Attributes []attribute.KeyValue
	}
	Output struct{}
}

// ***************** BootConfig *****************

// BootConfig for YAML
type BootConfig struct {
	Enabled  bool     `yaml:"enabled" json:"enabled"`
	Ignore   []string `yaml:"ignore" json:"ignore"`
	Exporter struct {
		File struct {
			Enabled    bool   `yaml:"enabled" json:"enabled"`
			OutputPath string `yaml:"outputPath" json:"outputPath"`
		} `yaml:"file" json:"file"`
		Otlp struct {
			Enabled  bool   `yaml:"enabled" json:"enabled"`
			Endpoint string `yaml:"endpoint" json:"endpoint"`
		} `yaml:"otlp" json:"otlp"`
		Zipkin struct {
			Enabled  bool   `yaml:"enabled" json:"enabled"`
			Endpoint string `yaml:"endpoint" json:"endpoint"`
		} `yaml:"zipkin" json:"zipkin"`
	} `yaml:"exporter" json:"exporter"`
}

// ToOptions convert BootConfig into Option list
func ToOptions(config *BootConfig, entryName, entryType string) []Option {
	opts := make([]Option, 0)

	if config.Enabled {
		var exporter sdktrace.SpanExporter

		if config.Exporter.File.Enabled {
			exporter = NewFileExporter(config.Exporter.File.OutputPath)
		}
		if config.Exporter.Otlp.Enabled {
			opts := make([]otlptracegrpc.Option, 0)
			client := otlptracegrpc.NewClient(opts...)
			if len(config.Exporter.Otlp.Endpoint) > 0 {
				opts := []otlptracegrpc.Option{
					otlptracegrpc.WithInsecure(),
					otlptracegrpc.WithEndpoint(config.Exporter.Otlp.Endpoint),
					otlptracegrpc.WithReconnectionPeriod(50 * time.Millisecond),
				}
				client = otlptracegrpc.NewClient(opts...)
			}

			exporter = NewOTLPTraceExporter(client)
		}
		if config.Exporter.Zipkin.Enabled {
			var url string
			if len(config.Exporter.Zipkin.Endpoint) > 0 {
				exporter = NewZipkinExporter(config.Exporter.Zipkin.Endpoint)
			}

			exporter = NewZipkinExporter(url)
		}
		opts = append(opts,
			WithEntryNameAndType(entryName, entryType),
			WithExporter(exporter),
			WithPathToIgnore(config.Ignore...))
	}

	return opts
}

// ***************** Option *****************

// Option is used while creating middleware as param
type Option func(*optionSet)

// WithExporter Provide sdktrace.SpanExporter.
func WithExporter(exporter sdktrace.SpanExporter) Option {
	return func(opt *optionSet) {
		if exporter != nil {
			opt.exporter = exporter
		}
	}
}

// WithSpanProcessor provide sdktrace.SpanProcessor.
func WithSpanProcessor(processor sdktrace.SpanProcessor) Option {
	return func(opt *optionSet) {
		if processor != nil {
			opt.processor = processor
		}
	}
}

// WithTracerProvider provide *sdktrace.TracerProvider.
func WithTracerProvider(provider *sdktrace.TracerProvider) Option {
	return func(opt *optionSet) {
		if provider != nil {
			opt.provider = provider
		}
	}
}

// WithPropagator provide propagation.TextMapPropagator.
func WithPropagator(propagator propagation.TextMapPropagator) Option {
	return func(opt *optionSet) {
		if propagator != nil {
			opt.propagator = propagator
		}
	}
}

// WithEntryNameAndType provide entry name and entry type.
func WithEntryNameAndType(entryName, entryType string) Option {
	return func(opt *optionSet) {
		opt.entryName = entryName
		opt.entryType = entryType
	}
}

// WithPathToIgnore provide paths prefix that will ignore.
func WithPathToIgnore(paths ...string) Option {
	return func(set *optionSet) {
		for i := range paths {
			if len(paths[i]) > 0 {
				set.pathToIgnore = append(set.pathToIgnore, paths[i])
			}
		}
	}
}

// WithMockOptionSet provide mock OptionSetInterface
func WithMockOptionSet(mock OptionSetInterface) Option {
	return func(set *optionSet) {
		set.mock = mock
	}
}

// ***************** Global *****************

// NoopExporter noop
type NoopExporter struct{}

// ExportSpans handles export of SpanSnapshots by dropping them.
func (nsb *NoopExporter) ExportSpans(context.Context, []sdktrace.ReadOnlySpan) error { return nil }

// Shutdown stops the exporter by doing nothing.
func (nsb *NoopExporter) Shutdown(context.Context) error { return nil }

// NewNoopExporter create a noop exporter
func NewNoopExporter() sdktrace.SpanExporter {
	return &NoopExporter{}
}

// NewFileExporter create a file exporter whose default output is stdout.
func NewFileExporter(outputPath string, opts ...stdouttrace.Option) sdktrace.SpanExporter {
	if opts == nil {
		opts = make([]stdouttrace.Option, 0)
	}

	if outputPath == "" {
		outputPath = "stdout"
	}

	if outputPath == "stdout" {
		opts = append(opts, stdouttrace.WithPrettyPrint())
	} else {
		// init lumberjack logger
		writer := rklogger.NewLumberjackConfigDefault()
		if !filepath.IsAbs(outputPath) {
			wd, _ := os.Getwd()
			outputPath = filepath.ToSlash(filepath.Join(wd, outputPath))
		}

		writer.Filename = outputPath

		opts = append(opts, stdouttrace.WithWriter(writer))
	}

	exporter, _ := stdouttrace.New(opts...)

	return exporter
}

func NewOTLPTraceExporter(client otexporterotlp.Client) sdktrace.SpanExporter {
	// Assign default otlp endpoint which is localhost:4317
	if client == nil {
		addr := "localhost:4317"
		client = otlptracegrpc.NewClient(
			otlptracegrpc.WithInsecure(),
			otlptracegrpc.WithEndpoint(addr),
			otlptracegrpc.WithReconnectionPeriod(50*time.Millisecond),
		)
	}
	exporter, err := otexporterotlp.New(context.Background(), client)

	if err != nil {
		rkentry.ShutdownWithError(err)
	}

	return exporter
}
func NewZipkinExporter(url string) sdktrace.SpanExporter {
	// Assign default zipkin endpoint which is localhost:9411
	if url == "" {
		url = "http://localhost:9411/api/v2/spans"
	}
	exporter, err := otexporterzipkin.New(url, otexporterzipkin.WithLogger(nil))
	if err != nil {
		rkentry.ShutdownWithError(err)
	}

	return exporter
}
