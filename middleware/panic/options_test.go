// Copyright (c) 2021 rookie-ninja
//
// Use of this source code is governed by an Apache-style
// license that can be found in the LICENSE file.

package panic

import (
	"github.com/rookie-ninja/rk-entry/v3/entry"
	"github.com/rookie-ninja/rk-entry/v3/error"
	"github.com/stretchr/testify/assert"
	"reflect"
	"testing"
)

func TestNewOptionSet(t *testing.T) {
	// without option
	set := NewOptionSet()
	assert.NotNil(t, set)

	// with option
	set = NewOptionSet(WithEntryNameAndKind("name", "type"))
	assert.Equal(t, "name", set.EntryName())
	assert.Equal(t, "type", set.EntryKind())
}

func TestOptionSet_BeforeCtx(t *testing.T) {
	event := rk.EventEntryNoop.EventFactory.CreateEventNoop()
	logger := rk.ZapEntryNoop.Logger
	handler := func(rkerror.ErrorInterface) {}

	set := NewOptionSet()
	ctx := set.BeforeCtx(event, logger, handler)
	assert.Equal(t, event, ctx.Input.Event)
	assert.Equal(t, logger, ctx.Input.Logger)
	assert.Equal(t, reflect.ValueOf(handler).Pointer(), reflect.ValueOf(ctx.Input.PanicHandler).Pointer())
}

func TestOptionSet_Before(t *testing.T) {
	event := rk.EventEntryNoop.EventFactory.CreateEventNoop()
	logger := rk.ZapEntryNoop.Logger
	handler := func(rkerror.ErrorInterface) {}

	set := NewOptionSet()
	ctx := set.BeforeCtx(event, logger, handler)
	set.Before(ctx)

	assert.NotNil(t, ctx.Output.DeferFunc)
	ctx.Output.DeferFunc()
}

func TestNewOptionSetMock(t *testing.T) {
	mock := NewOptionSetMock(NewBeforeCtx())
	assert.NotEmpty(t, mock.EntryName())
	assert.NotEmpty(t, mock.EntryKind())
	assert.NotNil(t, mock.BeforeCtx(nil, nil, nil))
	mock.Before(nil)
}
