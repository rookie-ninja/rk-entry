// Copyright (c) 2021 rookie-ninja
//
// Use of this source code is governed by an Apache-style
// license that can be found in the LICENSE file.

// Package meta is a middleware for metadata
package meta

import (
	"fmt"
	"github.com/rookie-ninja/rk-entry/v3/entry"
	"github.com/rookie-ninja/rk-entry/v3/middleware"
	"github.com/rookie-ninja/rk-query/v2"
	"net/http"
	"strings"
	"time"
)

// ***************** OptionSet Interface *****************

// OptionSetInterface mainly for testing purpose
type OptionSetInterface interface {
	EntryName() string

	EntryKind() string

	Before(*BeforeCtx)

	BeforeCtx(*http.Request, rkquery.Event) *BeforeCtx

	ShouldIgnore(string) bool
}

// ***************** OptionSet Implementation *****************

// optionSet which is used for middleware implementation
type optionSet struct {
	entryName       string
	entryKind       string
	prefix          string
	appNameKey      string
	appVersionKey   string
	appUnixTimeKey  string
	receivedTimeKey string
	pathToIgnore    []string
	mock            OptionSetInterface
}

// NewOptionSet Create new optionSet with options.
func NewOptionSet(opts ...Option) OptionSetInterface {
	set := &optionSet{
		entryName:    "fake-entry",
		entryKind:    "",
		prefix:       "RK",
		pathToIgnore: []string{},
	}

	for i := range opts {
		opts[i](set)
	}

	if set.mock != nil {
		return set.mock
	}

	if len(set.prefix) < 1 {
		set.prefix = "RK"
	}

	set.appNameKey = fmt.Sprintf("X-%s-App-Name", set.prefix)
	set.appVersionKey = fmt.Sprintf("X-%s-App-Version", set.prefix)
	set.appUnixTimeKey = fmt.Sprintf("X-%s-App-Unix-Time", set.prefix)
	set.receivedTimeKey = fmt.Sprintf("X-%s-Received-Time", set.prefix)

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

	ctx.Input.Request = req
	if req != nil && req.URL != nil {
		ctx.Input.UrlPath = req.URL.Path
	}

	ctx.Output.HeadersToReturn = make(map[string]string)
	ctx.Input.Event = event
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

	reqId := rkm.GenerateRequestId()
	now := time.Now().Format(time.RFC3339Nano)

	if ctx.Input.Event != nil {
		ctx.Input.Event.SetRequestId(reqId)
		ctx.Input.Event.SetEventId(reqId)
	}

	ctx.Output.RequestId = reqId

	ctx.Output.HeadersToReturn[rkm.HeaderRequestId] = reqId
	ctx.Output.HeadersToReturn[fmt.Sprintf("X-%s-Service-Name", set.prefix)] = rk.Registry.ServiceName()
	ctx.Output.HeadersToReturn[fmt.Sprintf("X-%s-Service-Version", set.prefix)] = rk.Registry.ServiceVersion()
	ctx.Output.HeadersToReturn[fmt.Sprintf("X-%s-Service-Unix-Time", set.prefix)] = now
	ctx.Output.HeadersToReturn[fmt.Sprintf("X-%s-Received-Time", set.prefix)] = now
	ctx.Output.HeadersToReturn[fmt.Sprintf("X-%s-Service-Domain", set.prefix)] = rkm.Domain.String
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
func (mock *optionSetMock) BeforeCtx(req *http.Request, event rkquery.Event) *BeforeCtx {
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
	ctx.Output.HeadersToReturn = make(map[string]string)
	return ctx
}

// BeforeCtx context for Before() function
type BeforeCtx struct {
	Input struct {
		UrlPath string
		Request *http.Request
		Event   rkquery.Event
	}
	Output struct {
		RequestId       string
		HeadersToReturn map[string]string
	}
}

// ***************** BootConfig *****************

// BootConfig for YAML
type BootConfig struct {
	Enabled bool     `yaml:"enabled" json:"enabled"`
	Prefix  string   `yaml:"prefix" json:"prefix"`
	Ignore  []string `yaml:"ignore" json:"ignore"`
}

// ToOptions convert BootConfig into Option list
func ToOptions(config *BootConfig, name, kind string) []Option {
	opts := make([]Option, 0)

	if config.Enabled {
		opts = append(opts,
			WithEntryNameAndKind(name, kind),
			WithPrefix(config.Prefix),
			WithPathToIgnore(config.Ignore...))
	}

	return opts
}

// ***************** Option *****************

// Option if for middleware options while creating middleware
type Option func(*optionSet)

// WithEntryNameAndKind provide entry name and entry kind.
func WithEntryNameAndKind(name, kind string) Option {
	return func(opt *optionSet) {
		opt.entryName = name
		opt.entryKind = kind
	}
}

// WithPrefix provide prefix.
func WithPrefix(prefix string) Option {
	return func(opt *optionSet) {
		opt.prefix = prefix
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
