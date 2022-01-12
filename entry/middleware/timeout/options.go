// Copyright (c) 2021 rookie-ninja
//
// Use of this source code is governed by an Apache-style
// license that can be found in the LICENSE file.

// package rkmidtimeout provide options
package rkmidtimeout

import (
	"github.com/rookie-ninja/rk-common/error"
	"github.com/rookie-ninja/rk-entry/entry"
	"github.com/rookie-ninja/rk-query"
	"github.com/rs/xid"
	"net/http"
	"strings"
	"time"
)

const global = "rk-global"

var (
	defaultErrResp = rkerror.New(
		rkerror.WithHttpCode(http.StatusRequestTimeout),
		rkerror.WithMessage("Request timed out!"))
	defaultTimeout = 10 * time.Second
)

// ***************** OptionSet Interface *****************

// OptionSetInterface mainly for testing purpose
type OptionSetInterface interface {
	GetEntryName() string

	GetEntryType() string

	BeforeCtx(*http.Request, rkquery.Event) *BeforeCtx

	Before(*BeforeCtx)
}

// ***************** OptionSet Implementation *****************

// Options which is used while initializing extension interceptor
type optionSet struct {
	entryName string
	entryType string
	timeouts  map[string]time.Duration
	mock      OptionSetInterface
}

// NewOptionSet Create new optionSet with options.
func NewOptionSet(opts ...Option) OptionSetInterface {
	set := &optionSet{
		entryName: xid.New().String(),
		entryType: "",
		timeouts:  make(map[string]time.Duration),
	}

	for i := range opts {
		opts[i](set)
	}

	if set.mock != nil {
		return set.mock
	}

	// add global timeout
	set.timeouts[global] = defaultTimeout

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
func (set *optionSet) BeforeCtx(req *http.Request, event rkquery.Event) *BeforeCtx {
	ctx := NewBeforeCtx()

	if event != nil {
		ctx.Input.Event = event
	}

	if req != nil && req.URL != nil {
		ctx.Input.UrlPath = req.URL.Path
	}

	ctx.Output.TimeoutErrResp = defaultErrResp

	return ctx
}

// Before should run before user handler
func (set *optionSet) Before(ctx *BeforeCtx) {
	if ctx == nil {
		return
	}

	// 1: get timeout
	timeoutDuration := set.getTimeout(ctx.Input.UrlPath)

	// 2: create three channels
	//
	// finishChan: triggered while request has been handled successfully
	// panicChan: triggered while panic occurs
	// timeoutChan: triggered while timing out
	finishChan := make(chan struct{}, 1)
	panicChan := make(chan interface{}, 1)
	timeoutChan := time.After(timeoutDuration)

	// 3: call init function from user
	ctx.Input.InitHandler()

	// 4: waiting function
	ctx.Output.WaitFunc = func() {
		go func() {
			defer func() {
				if recv := recover(); recv != nil {
					panicChan <- recv
				}
			}()

			ctx.Input.NextHandler()
			finishChan <- struct{}{}
		}()

		select {
		// 5.1: switch to original writer and panic
		case recv := <-panicChan:
			ctx.Input.PanicHandler()
			panic(recv)
		// 5.2: call user finish handler
		case <-finishChan:
			ctx.Input.FinishHandler()
		// 5.3: call user timeout handler
		case <-timeoutChan:
			ctx.Input.Event.SetCounter("timeout", 1)
			ctx.Input.TimeoutHandler()
		}
	}
}

// Get timeout instance with path.
// Global one will be returned if no not found.
func (set *optionSet) getTimeout(path string) time.Duration {
	if v, ok := set.timeouts[path]; ok {
		return v
	}

	return set.timeouts[global]
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
func (mock *optionSetMock) BeforeCtx(request *http.Request, event rkquery.Event) *BeforeCtx {
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
	ctx.Input.Event = rkentry.NoopEventLoggerEntry().GetEventFactory().CreateEventNoop()
	ctx.Input.TimeoutHandler = func() {}
	ctx.Input.FinishHandler = func() {}
	ctx.Input.InitHandler = func() {}
	ctx.Input.NextHandler = func() {}
	ctx.Input.PanicHandler = func() {}

	return ctx
}

// BeforeCtx context for Before() function
type BeforeCtx struct {
	Input struct {
		UrlPath        string
		InitHandler    func()
		NextHandler    func()
		PanicHandler   func()
		FinishHandler  func()
		TimeoutHandler func()
		Event          rkquery.Event
	}
	Output struct {
		WaitFunc       func()
		TimeoutErrResp *rkerror.ErrorResp
	}
}

// ***************** BootConfig *****************

// BootConfig for YAML
type BootConfig struct {
	Enabled   bool `yaml:"enabled" json:"enabled"`
	TimeoutMs int  `yaml:"timeoutMs" json:"timeoutMs"`
	Paths     []struct {
		Path      string `yaml:"path" json:"path"`
		TimeoutMs int    `yaml:"timeoutMs" json:"timeoutMs"`
	} `yaml:"paths" json:"paths"`
}

// ToOptions convert BootConfig into Option list
func ToOptions(config *BootConfig, entryName, entryType string) []Option {
	opts := make([]Option, 0)

	if config.Enabled {
		opts = append(opts, WithEntryNameAndType(entryName, entryType))

		timeout := time.Duration(config.TimeoutMs) * time.Millisecond
		opts = append(opts, WithTimeout(timeout))

		for i := range config.Paths {
			e := config.Paths[i]
			timeout := time.Duration(e.TimeoutMs) * time.Millisecond
			opts = append(opts, WithTimeoutByPath(e.Path, timeout))
		}
	}

	return opts
}

// ***************** Option *****************

// Option options provided to Interceptor or optionsSet while creating
type Option func(*optionSet)

// WithEntryNameAndType provide entry name and entry type.
func WithEntryNameAndType(entryName, entryType string) Option {
	return func(opt *optionSet) {
		opt.entryName = entryName
		opt.entryType = entryType
	}
}

// WithTimeoutAndResp Provide global timeout and response handler.
// If response is nil, default globalResponse will be assigned
func WithTimeout(timeout time.Duration) Option {
	return func(set *optionSet) {
		if timeout == 0 {
			timeout = defaultTimeout
		}

		defaultTimeout = timeout
	}
}

// WithTimeoutAndRespByPath Provide timeout and response handler by path.
// If response is nil, default globalResponse will be assigned
func WithTimeoutByPath(path string, timeout time.Duration) Option {
	return func(set *optionSet) {
		if !strings.HasPrefix(path, "/") {
			path = "/" + path
		}

		if timeout == 0 {
			timeout = defaultTimeout
		}

		set.timeouts[path] = timeout
	}
}

// WithMockOptionSet provide mock OptionSetInterface
func WithMockOptionSet(mock OptionSetInterface) Option {
	return func(set *optionSet) {
		set.mock = mock
	}
}
