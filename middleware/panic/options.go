// Copyright (c) 2021 rookie-ninja
//
// Use of this source code is governed by an Apache-style
// license that can be found in the LICENSE file.

// Package rkmidpanic provide options
package rkmidpanic

import (
	"fmt"
	"github.com/rookie-ninja/rk-entry/v2/error"
	rkmid "github.com/rookie-ninja/rk-entry/v2/middleware"
	"github.com/rookie-ninja/rk-query"
	"go.uber.org/zap"
	"net/http"
	"runtime/debug"
)

// ***************** OptionSet Interface *****************

// OptionSetInterface mainly for testing purpose
type OptionSetInterface interface {
	GetEntryName() string

	GetEntryType() string

	Before(*BeforeCtx)

	BeforeCtx(event rkquery.Event, logger *zap.Logger, handler handlerFunc) *BeforeCtx
}

// ***************** OptionSet Implementation *****************

// optionSet which is used for middleware implementation
type optionSet struct {
	entryName string
	entryType string
	mock      OptionSetInterface
}

// NewOptionSet Create new optionSet with options.
func NewOptionSet(opts ...Option) OptionSetInterface {
	set := &optionSet{
		entryName: "fake-entry",
		entryType: "",
	}

	for i := range opts {
		opts[i](set)
	}

	if set.mock != nil {
		return set.mock
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
				res = rkmid.GetErrorBuilder().New(http.StatusInternalServerError, "Panic occurs", recv)
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

// GetEntryName returns entry name
func (mock *optionSetMock) GetEntryName() string {
	return "mock"
}

// GetEntryType returns entry type
func (mock *optionSetMock) GetEntryType() string {
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

// WithEntryNameAndType Provide entry name and entry type.
func WithEntryNameAndType(entryName, entryType string) Option {
	return func(opt *optionSet) {
		opt.entryName = entryName
		opt.entryType = entryType
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
