// Copyright (c) 2021 rookie-ninja
//
// Use of this source code is governed by an Apache-style
// license that can be found in the LICENSE file.

package rkmidlog

import (
	"github.com/rookie-ninja/rk-entry/entry"
	"github.com/rookie-ninja/rk-query"
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
	assert.Equal(t, "/ut-path", ctx.Input.UrlPath)
}

func TestOptionSet_AfterCtx(t *testing.T) {
	set := NewOptionSet()

	ctx := set.AfterCtx("reqId", "traceId", "resCode")
	assert.Equal(t, "reqId", ctx.Input.RequestId)
	assert.Equal(t, "traceId", ctx.Input.TraceId)
	assert.Equal(t, "resCode", ctx.Input.ResCode)
}

func TestOptionSet_ignore(t *testing.T) {
	set := NewOptionSet(WithIgnorePrefix("/ut-path")).(*optionSet)
	assert.True(t, set.ignore("/ut-path"))
	assert.False(t, set.ignore("/"))
}

func TestOptionSet_createEvent(t *testing.T) {
	defer assertNotPanic(t)

	// with ignore url
	set := NewOptionSet(WithIgnorePrefix("/ut-ignore")).(*optionSet)
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
	assert.NotEmpty(t, ctx.Input.UrlPath)
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
		Enabled:                false,
		ZapLoggerEncoding:      json,
		ZapLoggerOutputPaths:   []string{},
		EventLoggerEncoding:    json,
		EventLoggerOutputPaths: []string{},
	}

	// with disabled
	assert.Empty(t, ToOptions(config, "", "", nil, nil))

	// with enabled
	config.Enabled = true
	assert.NotEmpty(t, ToOptions(config, "", "", nil, nil))
}

func TestNewOptionSet(t *testing.T) {
	// without options
	set := NewOptionSet().(*optionSet)
	assert.NotEmpty(t, set.GetEntryName())
	assert.NotNil(t, set.ZapLoggerEntry())
	assert.NotNil(t, set.EventLoggerEntry())
	assert.NotNil(t, set.zapLogger)
	assert.NotNil(t, set.zapLoggerOutputPath)
	assert.NotNil(t, set.eventLoggerOutputPath)
	assert.Empty(t, set.ignorePrefix)
}

func TestWithEntryNameAndType(t *testing.T) {
	set := NewOptionSet(
		WithEntryNameAndType("ut-entry", "ut-type")).(*optionSet)

	assert.Equal(t, "ut-entry", set.entryName)
	assert.Equal(t, "ut-type", set.entryType)
}

func TestWithZapLoggerEntry(t *testing.T) {
	entry := rkentry.NoopZapLoggerEntry()
	set := NewOptionSet(
		WithZapLoggerEntry(entry)).(*optionSet)
	assert.Equal(t, entry, set.zapLoggerEntry)
}

func TestWithEventLoggerEntry(t *testing.T) {
	entry := rkentry.NoopEventLoggerEntry()
	set := NewOptionSet(
		WithEventLoggerEntry(entry)).(*optionSet)
	assert.Equal(t, entry, set.eventLoggerEntry)
}

func TestWithZapLoggerEncoding(t *testing.T) {
	set := NewOptionSet(
		WithZapLoggerEncoding(json)).(*optionSet)

	assert.Equal(t, json, set.zapLoggerEncoding)
}

func TestWithZapLoggerOutputPaths(t *testing.T) {
	set := NewOptionSet(
		WithZapLoggerOutputPaths("ut-path")).(*optionSet)

	assert.Contains(t, set.zapLoggerOutputPath, "ut-path")
}

func TestWithEventLoggerEncoding(t *testing.T) {
	// Test with console encoding
	set := NewOptionSet(
		WithEventLoggerEncoding(console)).(*optionSet)
	assert.Equal(t, rkquery.CONSOLE, set.eventLoggerEncoding)

	// Test with json encoding
	set = NewOptionSet(
		WithEventLoggerEncoding(json)).(*optionSet)
	assert.Equal(t, rkquery.JSON, set.eventLoggerEncoding)

	// Test with non console and json
	set = NewOptionSet(
		WithEventLoggerEncoding("invalid")).(*optionSet)
	assert.Equal(t, rkquery.CONSOLE, set.eventLoggerEncoding)
}

func TestWithEventLoggerOutputPaths(t *testing.T) {
	set := NewOptionSet(
		WithEventLoggerOutputPaths("ut-path")).(*optionSet)
	assert.Contains(t, set.eventLoggerOutputPath, "ut-path")
}

func TestNewOptionSetMock(t *testing.T) {
	mock := NewOptionSetMock(NewBeforeCtx(), NewAfterCtx())
	assert.NotEmpty(t, mock.GetEntryName())
	assert.NotEmpty(t, mock.GetEntryType())
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
