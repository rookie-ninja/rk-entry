// Copyright (c) 2021 rookie-ninja
//
// Use of this source code is governed by an Apache-style
// license that can be found in the LICENSE file.

package logging

import (
	"github.com/rookie-ninja/rk-entry/v3/entry"
	"github.com/rookie-ninja/rk-query/v2"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestOptionSet_BeforeCtx(t *testing.T) {
	set := NewOptionSet()

	// without request
	ctx := set.BeforeCtx(nil)
	assert.NotNil(t, ctx)

	// with request
	req := httptest.NewRequest(http.MethodGet, "/ut-path", nil)
	ctx = set.BeforeCtx(req)
	assert.NotNil(t, ctx)
	assert.Equal(t, "/ut-path", ctx.Input.Path)
}

func TestOptionSet_AfterCtx(t *testing.T) {
	set := NewOptionSet()

	ctx := set.AfterCtx("reqId", "traceId", "resCode")
	assert.Equal(t, "reqId", ctx.Input.RequestId)
	assert.Equal(t, "traceId", ctx.Input.TraceId)
	assert.Equal(t, "resCode", ctx.Input.ResCode)
}

func TestOptionSet_ignore(t *testing.T) {
	set := NewOptionSet(WithPathToIgnore("/ut-path")).(*optionSet)
	assert.True(t, set.ShouldIgnore("/ut-path"))
	assert.False(t, set.ShouldIgnore("/"))
}

func TestOptionSet_createEvent(t *testing.T) {
	defer assertNotPanic(t)

	// with ignore url
	set := NewOptionSet(WithPathToIgnore("/ut-ignore")).(*optionSet)
	assert.NotNil(t, set.createEvent("/ut-ignore", true))

	// with thread safe
	assert.NotNil(t, set.createEvent("/", true))

	// with non-thread safe
	assert.NotNil(t, set.createEvent("/", false))
}

func TestOptionSet_Before(t *testing.T) {
	defer assertNotPanic(t)

	// with nil input
	set := NewOptionSet()
	set.Before(nil)

	// happy case
	ctx := set.BeforeCtx(httptest.NewRequest(http.MethodGet, "/ut-path", nil))
	set.Before(ctx)
	assert.NotNil(t, ctx.Output.Event)
	assert.NotNil(t, ctx.Output.Logger)
	assert.NotEmpty(t, ctx.Input.Path)
}

func TestOptionSet_After(t *testing.T) {
	defer assertNotPanic(t)

	// with nil input
	set := NewOptionSet()
	set.After(nil, nil)

	// happy case
	before := set.BeforeCtx(httptest.NewRequest(http.MethodGet, "/ut-path", nil))
	set.Before(before)
	after := set.AfterCtx("reqId", "traceId", "resCode")
	set.After(before, after)
}

func TestToOptions(t *testing.T) {
	config := &BootConfig{
		Enabled: false,
	}
	config.Zap.Encoding = json
	config.Zap.OutputPaths = []string{}
	config.Event.Encoding = json
	config.Event.OutputPaths = []string{}

	// with disabled
	assert.Empty(t, ToOptions(config, "", "", nil, nil))

	// with enabled
	config.Enabled = true
	assert.NotEmpty(t, ToOptions(config, "", "", nil, nil))
}

func TestNewOptionSet(t *testing.T) {
	// without options
	set := NewOptionSet().(*optionSet)
	assert.NotEmpty(t, set.EntryName())
	assert.NotNil(t, set.ZapEntry())
	assert.NotNil(t, set.EventEntry())
	assert.NotNil(t, set.zap.outputPaths)
	assert.NotNil(t, set.event.outputPaths)
	assert.Empty(t, set.pathToIgnore)
}

func TestWithEntryNameAndType(t *testing.T) {
	set := NewOptionSet(
		WithEntryNameAndKind("ut-entry", "ut-kind")).(*optionSet)

	assert.Equal(t, "ut-entry", set.entryName)
	assert.Equal(t, "ut-kind", set.entryKind)
}

func TestWithLoggerEntry(t *testing.T) {
	entry := rk.ZapEntryNoop
	set := NewOptionSet(
		WithZapEntry(entry)).(*optionSet)
	assert.Equal(t, entry, set.ZapEntry())
}

func TestWithEventLoggerEntry(t *testing.T) {
	entry := rk.EventEntryNoop
	set := NewOptionSet(
		WithEventEntry(entry)).(*optionSet)
	assert.Equal(t, entry, set.EventEntry())
}

func TestWithLoggerEncoding(t *testing.T) {
	set := NewOptionSet(
		WithZapEncoding(json)).(*optionSet)

	assert.Equal(t, json, set.zap.encoding)
}

func TestWithLoggerOutputPaths(t *testing.T) {
	set := NewOptionSet(
		WithZapOutputPaths("ut-path")).(*optionSet)

	assert.Contains(t, set.zap.outputPaths, "ut-path")
}

func TestWithEventLoggerEncoding(t *testing.T) {
	// Test with console encoding
	set := NewOptionSet(
		WithEventEncoding(console)).(*optionSet)
	assert.Equal(t, rkquery.CONSOLE, set.event.encoding)

	// Test with json encoding
	set = NewOptionSet(
		WithEventEncoding(json)).(*optionSet)
	assert.Equal(t, rkquery.JSON, set.event.encoding)

	// Test with non console and json
	set = NewOptionSet(
		WithEventEncoding("invalid")).(*optionSet)
	assert.Equal(t, rkquery.CONSOLE, set.event.encoding)
}

func TestWithEventLoggerOutputPaths(t *testing.T) {
	set := NewOptionSet(
		WithEventOutputPaths("ut-path")).(*optionSet)
	assert.Contains(t, set.event.outputPaths, "ut-path")
}

func TestNewOptionSetMock(t *testing.T) {
	mock := NewOptionSetMock(NewBeforeCtx(), NewAfterCtx())
	assert.NotEmpty(t, mock.EntryName())
	assert.NotEmpty(t, mock.EntryKind())
	assert.NotNil(t, mock.BeforeCtx(nil))
	assert.NotNil(t, mock.AfterCtx("", "", ""))
	mock.Before(nil)
	mock.After(nil, nil)
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
