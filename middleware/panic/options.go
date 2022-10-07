// Copyright (c) 2021 rookie-ninja
//
// Use of this source code is governed by an Apache-style
// license that can be found in the LICENSE file.

// Package panic provide options
package panic

import (
	"fmt"
	"github.com/rookie-ninja/rk-entry/v3/error"
	"github.com/rookie-ninja/rk-entry/v3/middleware"
	"github.com/rookie-ninja/rk-query/v2"
	"go.uber.org/zap"
	"net/http"
	"runtime/debug"
)

// ***************** OptionSet Interface *****************

// OptionSetInterface mainly for testing purpose
type OptionSetInterface interface {
	EntryName() string

	EntryKind() string

	Before(*BeforeCtx)

	BeforeCtx(event rkquery.Event, logger *zap.Logger, handler handlerFunc) *BeforeCtx
}

// ***************** OptionSet Implementation *****************

// optionSet which is used for middleware implementation
type optionSet struct {
	entryName string
	entryKind string
	mock      OptionSetInterface
}

// NewOptionSet Create new optionSet with options.
func NewOptionSet(opts ...Option) OptionSetInterface {
	set := &optionSet{
		entryName: "fake-entry",
		entryKind: "",
	}

	for i := range opts {
		opts[i](set)
	}

	if set.mock != nil {
		return set.mock
	}

	return set
}

// EntryName returns entry name
func (set *optionSet) EntryName() string {
	return set.entryName
}

// EntryKind returns entry kind
func (set *optionSet) EntryKind() string {
	return set.entryKind
}

// BeforeCtx should be created before Before()
func (set *optionSet) BeforeCtx(event rkquery.Event, logger *zap.Logger, handler handlerFunc) *BeforeCtx {
	ctx := NewBeforeCtx()
	ctx.Input.Event = event
	ctx.Input.Logger = logger
	ctx.Input.PanicHandler = handler

	return ctx
}

// Before should run before user handler
func (set *optionSet) Before(ctx *BeforeCtx) {
	if ctx == nil {
		return
	}

	ctx.Output.DeferFunc = func() {
		if recv := recover(); recv != nil {
			var res rkerror.ErrorInterface

			if se, ok := recv.(rkerror.ErrorInterface); ok {
				res = se
			} else {
				res = rkm.GetErrorBuilder().New(http.StatusInternalServerError, "Panic occurs", recv)
			}

			if ctx.Input.Event != nil {
				ctx.Input.Event.SetCounter("panic", 1)
				ctx.Input.Event.AddErr(res)
			}

			if ctx.Input.Logger != nil {
				ctx.Input.Logger.Error(fmt.Sprintf("panic occurs:\n%s", string(debug.Stack())), zap.Error(res))
			}

			if ctx.Input.PanicHandler != nil {
				ctx.Input.PanicHandler(res)
			}
		}
	}
}

// ***************** OptionSet Mock *****************

// NewOptionSetMock for testing purpose
func NewOptionSetMock(before *BeforeCtx) OptionSetInterface {
	return &optionSetMock{
		before: before,
	}
}

type optionSetMock struct {
	before *BeforeCtx
}

// EntryName returns entry name
func (mock *optionSetMock) EntryName() string {
	return "mock"
}

// EntryKind returns entry kind
func (mock *optionSetMock) EntryKind() string {
	return "mock"
}

// BeforeCtx should be created before Before()
func (mock *optionSetMock) BeforeCtx(event rkquery.Event, logger *zap.Logger, handler handlerFunc) *BeforeCtx {
	return mock.before
}

// Before should run before user handler
func (mock *optionSetMock) Before(ctx *BeforeCtx) {
	return
}

// ***************** Context *****************

// NewBeforeCtx create new BeforeCtx with fields initialized
func NewBeforeCtx() *BeforeCtx {
	ctx := &BeforeCtx{}
	return ctx
}

// BeforeCtx context for Before() function
type BeforeCtx struct {
	Input struct {
		Event        rkquery.Event
		Logger       *zap.Logger
		PanicHandler func(resp rkerror.ErrorInterface)
	}
	Output struct {
		DeferFunc func()
	}
}

// ***************** Option *****************

// Option is for middleware while creating
type Option func(*optionSet)

// WithEntryNameAndKind Provide entry name and entry type.
func WithEntryNameAndKind(name, kind string) Option {
	return func(opt *optionSet) {
		opt.entryName = name
		opt.entryKind = kind
	}
}

// WithMockOptionSet provide mock OptionSetInterface
func WithMockOptionSet(mock OptionSetInterface) Option {
	return func(set *optionSet) {
		set.mock = mock
	}
}

// User provided handler fun
type handlerFunc func(resp rkerror.ErrorInterface)
