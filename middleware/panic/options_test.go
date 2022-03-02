// Copyright (c) 2021 rookie-ninja
//
// Use of this source code is governed by an Apache-style
// license that can be found in the LICENSE file.

package rkmidpanic

import (
	"github.com/rookie-ninja/rk-entry/entry"
	rkerror "github.com/rookie-ninja/rk-entry/error"
	"github.com/stretchr/testify/assert"
	"reflect"
	"testing"
)

func TestNewOptionSet(t *testing.T) {
	// without option
	set := NewOptionSet()
	assert.NotNil(t, set)

	// with option
	set = NewOptionSet(WithEntryNameAndType("name", "type"))
	assert.Equal(t, "name", set.GetEntryName())
	assert.Equal(t, "type", set.GetEntryType())
}

func TestOptionSet_BeforeCtx(t *testing.T) {
	event := rkentry.EventEntryNoop.EventFactory.CreateEventNoop()
	logger := rkentry.LoggerEntryNoop.Logger
	handler := func(*rkerror.ErrorResp) {}

	set := NewOptionSet()
	ctx := set.BeforeCtx(event, logger, handler)
	assert.Equal(t, event, ctx.Input.Event)
	assert.Equal(t, logger, ctx.Input.Logger)
	assert.Equal(t, reflect.ValueOf(handler).Pointer(), reflect.ValueOf(ctx.Input.PanicHandler).Pointer())
}

func TestOptionSet_Before(t *testing.T) {
	event := rkentry.EventEntryNoop.EventFactory.CreateEventNoop()
	logger := rkentry.LoggerEntryNoop.Logger
	handler := func(*rkerror.ErrorResp) {}

	set := NewOptionSet()
	ctx := set.BeforeCtx(event, logger, handler)
	set.Before(ctx)

	assert.NotNil(t, ctx.Output.DeferFunc)
	ctx.Output.DeferFunc()
}

func TestNewOptionSetMock(t *testing.T) {
	mock := NewOptionSetMock(NewBeforeCtx())
	assert.NotEmpty(t, mock.GetEntryName())
	assert.NotEmpty(t, mock.GetEntryType())
	assert.NotNil(t, mock.BeforeCtx(nil, nil, nil))
	mock.Before(nil)
}
