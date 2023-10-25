// Copyright (c) 2021 rookie-ninja
//
// Use of this source code is governed by an Apache-style
// license that can be found in the LICENSE file.
package rkmidtrace

import (
	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestWithEntryNameAndType(t *testing.T) {
	set := NewOptionSet(
		WithEntryNameAndType("ut-entry", "ut-type")).(*optionSet)

	assert.Equal(t, "ut-entry", set.GetEntryName())
	assert.Equal(t, "ut-type", set.GetEntryType())
}

func TestWithExporter(t *testing.T) {
	exporter := &NoopExporter{}
	set := NewOptionSet(
		WithExporter(exporter)).(*optionSet)

	assert.Equal(t, exporter, set.exporter)
}

func TestWithSpanProcessor(t *testing.T) {
	processor := sdktrace.NewSimpleSpanProcessor(&NoopExporter{})
	set := NewOptionSet(
		WithSpanProcessor(processor)).(*optionSet)

	assert.Equal(t, processor, set.processor)
}

func TestWithTracerProvider(t *testing.T) {
	provider := sdktrace.NewTracerProvider()
	set := NewOptionSet(
		WithTracerProvider(provider)).(*optionSet)

	assert.Equal(t, provider, set.GetProvider())
}

func TestWithPropagator(t *testing.T) {
	prop := propagation.NewCompositeTextMapPropagator()
	set := NewOptionSet(
		WithPropagator(prop)).(*optionSet)

	assert.Equal(t, prop, set.GetPropagator())
}

func TestNoopExporter_ExportSpans(t *testing.T) {
	exporter := NoopExporter{}
	assert.Nil(t, exporter.ExportSpans(nil, nil))
}

func TestNoopExporter_Shutdown(t *testing.T) {
	exporter := NoopExporter{}
	assert.Nil(t, exporter.Shutdown(nil))
}

func TestCreateNoopExporter(t *testing.T) {
	assert.NotNil(t, NewNoopExporter())
}
func TestCreateOtlpExporter(t *testing.T) {
	defer assertNotPanic(t)

	// without endpoint
	exporter := NewOTLPTraceExporter(nil)
	assert.NotNil(t, exporter)

	// with default otlp collector
	opts := make([]otlptracegrpc.Option, 0)
	client := otlptracegrpc.NewClient(opts...)
	exporter = NewOTLPTraceExporter(client)
	assert.NotNil(t, exporter)
}
func TestCreateZipkinExporter(t *testing.T) {
	defer assertNotPanic(t)

	// without endpoint
	var url string
	exporter := NewZipkinExporter(url)
	assert.NotNil(t, exporter)

	// with default Zipkin
	url = "http://localhost:9411/api/v2/spans"
	exporter = NewZipkinExporter(url)
	assert.NotNil(t, exporter)
}
func TestCreateFileExporter(t *testing.T) {
	defer assertNotPanic(t)

	// with stdout
	exporter := NewFileExporter("")
	assert.NotNil(t, exporter)

	// with non stdout
	exporter = NewFileExporter("stderror")
	assert.NotNil(t, exporter)
}

func TestToOptions(t *testing.T) {
	defer assertNotPanic(t)

	config := &BootConfig{
		Enabled: false,
	}

	// with disabled
	assert.Empty(t, ToOptions(config, "", ""))

	// with enabled
	config.Enabled = true
	assert.NotEmpty(t, ToOptions(config, "", ""))

	// with file exporter
	config.Enabled = true
	config.Exporter.File.Enabled = true
	config.Exporter.File.OutputPath = "output"
	NewOptionSet(ToOptions(config, "", "")...)

	// with otlp collector
	config = &BootConfig{
		Enabled: true,
	}
	config.Exporter.Otlp.Enabled = true
	NewOptionSet(ToOptions(config, "", "")...)
	// with zipkin
	config = &BootConfig{
		Enabled: true,
	}
	config.Exporter.Zipkin.Enabled = true
	NewOptionSet(ToOptions(config, "", "")...)
}

func TestOptionSet_BeforeCtx(t *testing.T) {
	// with nil request
	set := NewOptionSet()
	ctx := set.BeforeCtx(nil, true)
	assert.NotEmpty(t, ctx.Input.Attributes)
	assert.True(t, ctx.Input.IsClient)

	// with request
	req := httptest.NewRequest(http.MethodGet, "/ut", nil)
	ctx = set.BeforeCtx(req, true)
	assert.NotEmpty(t, ctx.Input.Attributes)
	assert.True(t, ctx.Input.IsClient)
	assert.NotNil(t, ctx.Input.RequestCtx)
	assert.NotNil(t, ctx.Input.Carrier)
}

func TestOptionSet_AfterCtx(t *testing.T) {
	set := NewOptionSet()
	ctx := set.AfterCtx(200, "msg")
	assert.Equal(t, 200, ctx.Input.ResCode)
	assert.Equal(t, "msg", ctx.Input.ResMsg)
	assert.NotNil(t, ctx.Input.Attributes)
}

func TestOptionSet_Before(t *testing.T) {
	// with client span
	set := NewOptionSet()
	req := httptest.NewRequest(http.MethodGet, "/ut", nil)
	ctx := set.BeforeCtx(req, true)
	set.Before(ctx)

	assert.NotNil(t, ctx.Output.Span)
	assert.NotNil(t, ctx.Output.NewCtx)

	// with server span
	set = NewOptionSet()
	req = httptest.NewRequest(http.MethodGet, "/ut", nil)
	ctx = set.BeforeCtx(req, false)
	set.Before(ctx)

	assert.NotNil(t, ctx.Output.Span)
	assert.NotNil(t, ctx.Output.NewCtx)
}

func TestOptionSet_After(t *testing.T) {
	defer assertNotPanic(t)

	set := NewOptionSet()
	req := httptest.NewRequest(http.MethodGet, "/ut", nil)

	before := set.BeforeCtx(req, false)
	set.Before(before)

	after := set.AfterCtx(200, "msg")
	set.After(before, after)
}

func TestNewOptionSetMock(t *testing.T) {
	mock := NewOptionSetMock(NewBeforeCtx(), NewAfterCtx(), nil, nil, nil)
	assert.NotEmpty(t, mock.GetEntryName())
	assert.NotEmpty(t, mock.GetEntryType())
	assert.NotNil(t, mock.BeforeCtx(nil, false))
	assert.NotNil(t, mock.AfterCtx(1, ""))
	mock.Before(nil)
	mock.After(nil, nil)
	assert.Nil(t, mock.GetTracer())
	assert.Nil(t, mock.GetProvider())
	assert.Nil(t, mock.GetPropagator())
}

func assertNotPanic(t *testing.T) {
	if r := recover(); r != nil {
		// Expect panic to be called with non nil error
		assert.True(t, false)
	} else {
		// This should never be called in case of a bug
		assert.True(t, true)
	}
}
