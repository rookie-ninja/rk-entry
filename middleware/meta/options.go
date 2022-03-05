// Copyright (c) 2021 rookie-ninja
//
// Use of this source code is governed by an Apache-style
// license that can be found in the LICENSE file.

// Package rkmidmeta is a middleware for metadata
package rkmidmeta

import (
	"fmt"
	"github.com/rookie-ninja/rk-entry/v2/entry"
	"github.com/rookie-ninja/rk-entry/v2/middleware"
	"github.com/rookie-ninja/rk-query"
	"net/http"
	"strings"
	"time"
)

// ***************** OptionSet Interface *****************

// OptionSetInterface mainly for testing purpose
type OptionSetInterface interface {
	GetEntryName() string

	GetEntryType() string

	Before(*BeforeCtx)

	BeforeCtx(*http.Request, rkquery.Event) *BeforeCtx

	ShouldIgnore(string) bool
}

// ***************** OptionSet Implementation *****************

// optionSet which is used for middleware implementation
type optionSet struct {
	entryName       string
	entryType       string
	prefix          string
	localeKey       string
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
		entryType:    "",
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
	set.localeKey = fmt.Sprintf("X-%s-Locale", set.prefix)

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

	reqId := rkmid.GenerateRequestId()
	now := time.Now().Format(time.RFC3339Nano)

	if ctx.Input.Event != nil {
		ctx.Input.Event.SetRequestId(reqId)
		ctx.Input.Event.SetEventId(reqId)
	}

	ctx.Output.RequestId = reqId

	ctx.Output.HeadersToReturn[rkmid.HeaderRequestId] = reqId
	ctx.Output.HeadersToReturn[fmt.Sprintf("X-%s-App-Name", set.prefix)] = rkentry.GlobalAppCtx.GetAppInfoEntry().AppName
	ctx.Output.HeadersToReturn[fmt.Sprintf("X-%s-App-Version", set.prefix)] = rkentry.GlobalAppCtx.GetAppInfoEntry().Version
	ctx.Output.HeadersToReturn[fmt.Sprintf("X-%s-App-Unix-Time", set.prefix)] = now
	ctx.Output.HeadersToReturn[fmt.Sprintf("X-%s-Received-Time", set.prefix)] = now
	ctx.Output.HeadersToReturn[fmt.Sprintf("X-%s-App-Locale", set.prefix)] = strings.Join([]string{
		rkmid.Realm.String, rkmid.Region.String, rkmid.AZ.String, rkmid.Domain.String,
	}, "::")
}

// ShouldIgnore determine whether auth should be ignored based on path
func (set *optionSet) ShouldIgnore(path string) bool {
	for i := range set.pathToIgnore {
		if strings.HasPrefix(path, set.pathToIgnore[i]) {
			return true
		}
	}

	return rkmid.ShouldIgnoreGlobal(path)
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
func ToOptions(config *BootConfig, entryName, entryType string) []Option {
	opts := make([]Option, 0)

	if config.Enabled {
		opts = append(opts,
			WithEntryNameAndType(entryName, entryType),
			WithPrefix(config.Prefix),
			WithPathToIgnore(config.Ignore...))
	}

	return opts
}

// ***************** Option *****************

// Option if for middleware options while creating middleware
type Option func(*optionSet)

// WithEntryNameAndType provide entry name and entry type.
func WithEntryNameAndType(entryName, entryType string) Option {
	return func(opt *optionSet) {
		opt.entryName = entryName
		opt.entryType = entryType
	}
}

// WithPrefix provide prefix.
func WithPrefix(prefix string) Option {
	return func(opt *optionSet) {
		opt.prefix = prefix
	}
}

// WithPathToIgnore provide paths prefix that will ignore.
// Mainly used for swagger main page and RK TV entry.
func WithPathToIgnore(paths ...string) Option {
	return func(set *optionSet) {
		set.pathToIgnore = append(set.pathToIgnore, paths...)
	}
}

// WithMockOptionSet provide mock OptionSetInterface
func WithMockOptionSet(mock OptionSetInterface) Option {
	return func(set *optionSet) {
		set.mock = mock
	}
}
