// Copyright (c) 2021 rookie-ninja
//
// Use of this source code is governed by an Apache-style
// license that can be found in the LICENSE file.

package rkmidtimeout

import (
	rkentry "github.com/rookie-ninja/rk-entry/entry"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestNewOptionSet(t *testing.T) {
	// without options
	set := NewOptionSet().(*optionSet)
	assert.NotEmpty(t, set.GetEntryName())
	assert.NotEmpty(t, set.timeouts)

	// with options
	set = NewOptionSet(
		WithEntryNameAndType("name", "type"),
		WithTimeout(1*time.Second),
		WithTimeoutByPath("/ut", 1*time.Second)).(*optionSet)
	assert.Equal(t, "name", set.GetEntryName())
	assert.Equal(t, "type", set.GetEntryType())
	assert.Equal(t, 1*time.Second, set.timeouts[global])
	assert.Equal(t, 1*time.Second, set.getTimeout("/ut"))
}

func TestOptionSet_BeforeCtx(t *testing.T) {
	// with nil input
	set := NewOptionSet()
	ctx := set.BeforeCtx(nil, nil)
	assert.NotNil(t, ctx)

	// with req
	req := httptest.NewRequest(http.MethodGet, "/ut", nil)
	event := rkentry.NoopEventLoggerEntry().GetEventFactory().CreateEventNoop()
	ctx = set.BeforeCtx(req, event)

	assert.Equal(t, "/ut", ctx.Input.UrlPath)
	assert.Equal(t, event, ctx.Input.Event)
}

func TestOptionSet_Before(t *testing.T) {
	var initCall, nextCall, finishCall, panicCall, timeoutCall bool
	initF := func() {
		initCall = true
	}
	nextF := func() {
		nextCall = true
	}
	panicF := func() {
		panicCall = true
	}
	finishF := func() {
		finishCall = true
	}
	timeoutF := func() {
		timeoutCall = true
	}

	// without timeout
	set := NewOptionSet(
		WithTimeout(1 * time.Second))
	req := httptest.NewRequest(http.MethodGet, "/ut", nil)
	event := rkentry.NoopEventLoggerEntry().GetEventFactory().CreateEventNoop()
	ctx := set.BeforeCtx(req, event)

	ctx.Input.InitHandler = initF
	ctx.Input.NextHandler = nextF
	ctx.Input.PanicHandler = panicF
	ctx.Input.FinishHandler = finishF
	ctx.Input.TimeoutHandler = timeoutF

	set.Before(ctx)
	ctx.Output.WaitFunc()
	time.Sleep(1 * time.Second)

	assert.True(t, initCall)
	assert.True(t, nextCall)
	assert.True(t, finishCall)
	assert.False(t, panicCall)
	assert.False(t, timeoutCall)

	initCall = false
	nextCall = false
	finishCall = false

	// with timeout
	nextF = func() {
		nextCall = true
		time.Sleep(2 * time.Second)
	}
	ctx.Input.NextHandler = nextF
	set.Before(ctx)
	ctx.Output.WaitFunc()

	assert.True(t, initCall)
	assert.True(t, nextCall)
	assert.False(t, finishCall)
	assert.False(t, panicCall)
	assert.True(t, timeoutCall)

	// with panic
	nextF = func() {
		nextCall = true
		panic("")
	}
	ctx.Input.NextHandler = nextF
	set.Before(ctx)
	defer assertPanic(t)
	ctx.Output.WaitFunc()

	assert.True(t, initCall)
	assert.True(t, nextCall)
	assert.False(t, finishCall)
	assert.True(t, panicCall)
	assert.False(t, timeoutCall)
}

func TestToOptions(t *testing.T) {
	// with disabled
	config := &BootConfig{
		Enabled: false,
	}
	assert.Empty(t, ToOptions(config, "", ""))

	// with enabled
	config.Enabled = true
	assert.NotEmpty(t, ToOptions(config, "", ""))
}

func TestNewOptionSetMock(t *testing.T) {
	mock := NewOptionSetMock(NewBeforeCtx())
	assert.NotEmpty(t, mock.GetEntryName())
	assert.NotEmpty(t, mock.GetEntryType())
	assert.NotNil(t, mock.BeforeCtx(nil, nil))
	mock.Before(nil)
}

func assertPanic(t *testing.T) {
	if r := recover(); r != nil {
		// Expect panic to be called with non nil error
		assert.True(t, true)
	} else {
		// This should never be called in case of a bug
		assert.True(t, false)
	}
}
