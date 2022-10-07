// Copyright (c) 2021 rookie-ninja
//
// Use of this source code is governed by an Apache-style
// license that can be found in the LICENSE file.

// Package timeout provide options
package timeout

import (
	"github.com/rookie-ninja/rk-entry/v3/entry"
	"github.com/rookie-ninja/rk-entry/v3/error"
	"github.com/rookie-ninja/rk-entry/v3/middleware"
	"github.com/rookie-ninja/rk-query/v2"
	"net/http"
	"strings"
	"time"
)

const global = "rk-global"

var (
	defaultErrResp = rkm.GetErrorBuilder().New(http.StatusRequestTimeout, "")
	defaultTimeout = 10 * time.Second
)

// ***************** OptionSet Interface *****************

// OptionSetInterface mainly for testing purpose
type OptionSetInterface interface {
	EntryName() string

	EntryKind() string

	BeforeCtx(*http.Request, rkquery.Event) *BeforeCtx

	Before(*BeforeCtx)

	ShouldIgnore(string) bool
}

// ***************** OptionSet Implementation *****************

// Options which is used while initializing extension interceptor
type optionSet struct {
	entryName    string
	entryKind    string
	pathToIgnore []string
	timeouts     map[string]time.Duration
	mock         OptionSetInterface
}

// NewOptionSet Create new optionSet with options.
func NewOptionSet(opts ...Option) OptionSetInterface {
	set := &optionSet{
		entryName:    "fake-entry",
		entryKind:    "",
		pathToIgnore: []string{},
		timeouts:     make(map[string]time.Duration),
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

// EntryName returns entry name
func (set *optionSet) EntryName() string {
	return set.entryName
}

// EntryKind returns entry kind
func (set *optionSet) EntryKind() string {
	return set.entryKind
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

	// case 0: ignore path
	if set.ShouldIgnore(ctx.Input.UrlPath) {
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

// ShouldIgnore determine whether auth should be ignored based on path
func (set *optionSet) ShouldIgnore(path string) bool {
	for i := range set.pathToIgnore {
		if strings.HasPrefix(path, set.pathToIgnore[i]) {
			return true
		}
	}

	return rkm.ShouldIgnoreGlobal(path)
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
func (mock *optionSetMock) BeforeCtx(request *http.Request, event rkquery.Event) *BeforeCtx {
	return mock.before
}

// Before should run before user handler
func (mock *optionSetMock) Before(ctx *BeforeCtx) {
	return
}

// ShouldIgnore should run before user handler
func (mock *optionSetMock) ShouldIgnore(string) bool {
	return false
}

// ***************** Context *****************

// NewBeforeCtx create new BeforeCtx with fields initialized
func NewBeforeCtx() *BeforeCtx {
	ctx := &BeforeCtx{}
	ctx.Input.Event = rk.EventEntryNoop.CreateEventNoop()
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
		TimeoutErrResp rkerror.ErrorInterface
	}
}

// ***************** BootConfig *****************

// BootConfig for YAML
type BootConfig struct {
	Enabled   bool     `yaml:"enabled" json:"enabled"`
	TimeoutMs int      `yaml:"timeoutMs" json:"timeoutMs"`
	Ignore    []string `yaml:"ignore" json:"ignore"`
	Paths     []struct {
		Path      string `yaml:"path" json:"path"`
		TimeoutMs int    `yaml:"timeoutMs" json:"timeoutMs"`
	} `yaml:"paths" json:"paths"`
}

// ToOptions convert BootConfig into Option list
func ToOptions(config *BootConfig, entryName, entryType string) []Option {
	opts := make([]Option, 0)

	if config.Enabled {
		opts = append(opts, WithEntryNameAndKind(entryName, entryType))

		timeout := time.Duration(config.TimeoutMs) * time.Millisecond
		opts = append(opts, WithTimeout(timeout))

		for i := range config.Paths {
			e := config.Paths[i]
			timeout := time.Duration(e.TimeoutMs) * time.Millisecond
			opts = append(opts, WithTimeoutByPath(e.Path, timeout))
		}

		opts = append(opts, WithPathToIgnore(config.Ignore...))
	}

	return opts
}

// ***************** Option *****************

// Option options provided to Interceptor or optionsSet while creating
type Option func(*optionSet)

// WithEntryNameAndKind provide entry name and entry kind.
func WithEntryNameAndKind(name, kind string) Option {
	return func(opt *optionSet) {
		opt.entryName = name
		opt.entryKind = kind
	}
}

// WithTimeout Provide global timeout and response handler.
// If response is nil, default globalResponse will be assigned
func WithTimeout(timeout time.Duration) Option {
	return func(set *optionSet) {
		if timeout == 0 {
			timeout = defaultTimeout
		}

		defaultTimeout = timeout
	}
}

// WithTimeoutByPath Provide timeout and response handler by path.
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

// WithPathToIgnore provide paths prefix that will ignore.
func WithPathToIgnore(paths ...string) Option {
	return func(set *optionSet) {
		for i := range paths {
			if len(paths[i]) > 0 {
				set.pathToIgnore = append(set.pathToIgnore, paths[i])
			}
		}
	}
}

// WithMockOptionSet provide mock OptionSetInterface
func WithMockOptionSet(mock OptionSetInterface) Option {
	return func(set *optionSet) {
		set.mock = mock
	}
}
