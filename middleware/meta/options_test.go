// Copyright (c) 2021 rookie-ninja
//
// Use of this source code is governed by an Apache-style
// license that can be found in the LICENSE file.

package rkmidmeta

import (
	"github.com/rookie-ninja/rk-entry/v2/entry"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
)

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

func TestNewOptionSet(t *testing.T) {
	// with empty prefix
	set := NewOptionSet().(*optionSet)
	assert.Equal(t, "RK", set.prefix)

	// with prefix
	set = NewOptionSet(WithPrefix("ut")).(*optionSet)
	assert.Equal(t, "ut", set.prefix)
}

func TestOptionSet_BeforeCtx(t *testing.T) {
	set := NewOptionSet().(*optionSet)
	event := rkentry.EventEntryNoop.EventFactory.CreateEventNoop()
	req := httptest.NewRequest(http.MethodGet, "/ut", nil)
	ctx := set.BeforeCtx(req, event)

	assert.Equal(t, event, ctx.Input.Event)
	assert.Equal(t, req, ctx.Input.Request)

	assert.NotNil(t, ctx.Output.HeadersToReturn)
}

func TestOptionSet_Before(t *testing.T) {
	defer assertNotPanic(t)

	// with nil ctx
	set := NewOptionSet()
	set.Before(nil)

	req := httptest.NewRequest(http.MethodGet, "/ut", nil)

	ctx := set.BeforeCtx(req, rkentry.EventEntryNoop.EventFactory.CreateEventNoop())
	set.Before(ctx)
	assert.NotEmpty(t, ctx.Output.HeadersToReturn)
}

func TestNewOptionSetMock(t *testing.T) {
	mock := NewOptionSetMock(NewBeforeCtx())
	assert.NotEmpty(t, mock.GetEntryName())
	assert.NotEmpty(t, mock.GetEntryType())
	assert.NotNil(t, mock.BeforeCtx(nil, nil))
	mock.Before(nil)
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
