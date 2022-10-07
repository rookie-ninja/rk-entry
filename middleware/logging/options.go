// Copyright (c) 2021 rookie-ninja
//
// Use of this source code is governed by an Apache-style
// license that can be found in the LICENSE file.

// Package logging provide options
package logging

import (
	"github.com/rookie-ninja/rk-entry/v3/entry"
	"github.com/rookie-ninja/rk-entry/v3/middleware"
	"github.com/rookie-ninja/rk-entry/v3/util"
	"github.com/rookie-ninja/rk-logger"
	"github.com/rookie-ninja/rk-query/v2"
	"go.uber.org/zap"
	"net/http"
	"os"
	"path"
	"strings"
	"time"
)

const (
	// Console encoding style of logging
	console = "console"
	// Json encoding style of logging
	json = "json"
)

// ***************** OptionSet Interface *****************

// OptionSetInterface mainly for testing purpose
type OptionSetInterface interface {
	EntryName() string

	EntryKind() string

	BeforeCtx(*http.Request) *BeforeCtx

	Before(*BeforeCtx)

	AfterCtx(reqId, traceId, resCode string) *AfterCtx

	After(before *BeforeCtx, after *AfterCtx)

	ShouldIgnore(string) bool
}

// ***************** OptionSet Implementation *****************

// optionSet which is used for middleware implementation
type optionSet struct {
	entryName    string
	entryKind    string
	pathToIgnore []string
	mock         OptionSetInterface
	zap          struct {
		entry       *rk.ZapEntry
		encoding    string
		outputPaths []string
	}
	event struct {
		entry       *rk.EventEntry
		encoding    rkquery.Encoding
		outputPaths []string
		baseLogger  *zap.Logger
	}
}

// NewOptionSet Create new optionSet with options.
func NewOptionSet(opts ...Option) OptionSetInterface {
	set := &optionSet{
		entryName:    "fake-entry",
		entryKind:    "",
		pathToIgnore: []string{},
	}
	set.zap.entry = rk.NewZapEntryStdout()
	set.zap.outputPaths = make([]string, 0)
	set.event.entry = rk.NewEventEntryStdout()
	set.event.outputPaths = make([]string, 0)

	for i := range opts {
		opts[i](set)
	}

	if set.mock != nil {
		return set.mock
	}

	// Override zap logger encoding and output path if provided by user
	// Override encoding type
	if set.zap.encoding == json || len(set.zap.outputPaths) > 0 {
		loggerConfig := rklogger.NewZapStdoutConfig()
		lumberjackConfig := rklogger.NewLumberjackConfigDefault()

		if set.zap.encoding == json {
			loggerConfig.Encoding = "json"
		}

		if len(set.zap.outputPaths) > 0 {
			loggerConfig.OutputPaths = toAbsPath(set.zap.outputPaths...)
		}

		if logger, err := rklogger.NewZapLoggerWithConf(loggerConfig, lumberjackConfig); err != nil {
			rku.ShutdownWithError(err)
		} else {
			set.zap.entry.Logger = logger
			set.zap.entry.LoggerConfig = loggerConfig
			set.zap.entry.Rotator = lumberjackConfig
		}
	}

	// Override event logger output path if provided by user
	if len(set.event.outputPaths) > 0 {
		loggerConfig := rklogger.NewZapStdoutConfig()
		lumberjackConfig := rklogger.NewLumberjackConfigDefault()

		loggerConfig.OutputPaths = toAbsPath(set.event.outputPaths...)

		if logger, err := rklogger.NewZapLoggerWithConf(loggerConfig, lumberjackConfig); err != nil {
			rku.ShutdownWithError(err)
		} else {
			set.event.baseLogger = logger
		}
	} else {
		set.event.baseLogger = set.event.entry.BaseLogger
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

// BeforeCtx should be created before this
func (set *optionSet) BeforeCtx(req *http.Request) *BeforeCtx {
	ctx := NewBeforeCtx()

	if req != nil && req.URL != nil && req.Header != nil {
		remoteIp, remotePort := rkm.GetRemoteAddressSet(req)
		ctx.Input.RemoteAddr = remoteIp + ":" + remotePort

		ctx.Input.Path = req.URL.Path
		ctx.Input.Method = req.Method
		ctx.Input.RawQuery = req.URL.RawQuery
		ctx.Input.Protocol = req.Proto
		ctx.Input.UserAgent = req.UserAgent()
	}

	return ctx
}

// Before should run before user handler
func (set *optionSet) Before(ctx *BeforeCtx) {
	if ctx == nil {
		return
	}

	ctx.Output.Event = set.createEvent(ctx.Input.Path, true)
	ctx.Output.Logger = set.zap.entry.Logger

	ctx.Output.Event.SetRemoteAddr(ctx.Input.RemoteAddr)

	ctx.Output.Event.AddPayloads([]zap.Field{
		zap.String("path", ctx.Input.Path),
		zap.String("method", ctx.Input.Method),
		zap.String("query", ctx.Input.RawQuery),
		zap.String("protocol", ctx.Input.Protocol),
		zap.String("userAgent", ctx.Input.UserAgent),
	}...)

	ctx.Output.Event.AddPayloads(ctx.Input.Fields...)

	ctx.Output.Event.SetOperation(ctx.Input.Path)
}

// AfterCtx should be created before After()
func (set *optionSet) AfterCtx(reqId, traceId, resCode string) *AfterCtx {
	ctx := NewAfterCtx()

	ctx.Input.RequestId = reqId
	ctx.Input.TraceId = traceId
	ctx.Input.ResCode = resCode

	return ctx
}

// After should run after user handler
func (set *optionSet) After(before *BeforeCtx, after *AfterCtx) {
	if before == nil || after == nil {
		return
	}

	event := before.Output.Event

	if len(after.Input.RequestId) > 0 {
		event.SetEventId(after.Input.RequestId)
		event.SetRequestId(after.Input.RequestId)
	}

	if len(after.Input.TraceId) > 0 {
		event.SetTraceId(after.Input.TraceId)
	}

	event.SetResCode(after.Input.ResCode)
	event.SetEndTime(time.Now())
	event.Finish()
}

// EventEntry returns rk.EventEntry
func (set *optionSet) EventEntry() *rk.EventEntry {
	return set.event.entry
}

// ZapEntry returns rk.ZapEntry
func (set *optionSet) ZapEntry() *rk.ZapEntry {
	return set.zap.entry
}

// CreateEvent create event based on urlPath
func (set *optionSet) createEvent(urlPath string, threadSafe bool) rkquery.Event {
	if set.ShouldIgnore(urlPath) {
		return set.EventEntry().EventFactory.CreateEventNoop()
	}

	var event rkquery.Event
	if threadSafe {
		event = set.event.entry.CreateEventThreadSafe(
			rkquery.WithZapLogger(set.event.baseLogger),
			rkquery.WithEncoding(set.event.encoding),
			rkquery.WithServiceName(rk.Registry.ServiceName()),
			rkquery.WithServiceVersion(rk.Registry.ServiceVersion()),
			rkquery.WithEntryName(set.EntryName()),
			rkquery.WithEntryKind(set.EntryKind()))
	} else {
		event = set.event.entry.CreateEvent(
			rkquery.WithZapLogger(set.event.baseLogger),
			rkquery.WithEncoding(set.event.encoding),
			rkquery.WithServiceName(rk.Registry.ServiceName()),
			rkquery.WithServiceVersion(rk.Registry.ServiceVersion()),
			rkquery.WithEntryName(set.EntryName()),
			rkquery.WithEntryKind(set.EntryKind()))
	}

	event.SetStartTime(time.Now())

	return event
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
func NewOptionSetMock(before *BeforeCtx, after *AfterCtx) OptionSetInterface {
	return &optionSetMock{
		before: before,
		after:  after,
	}
}

type optionSetMock struct {
	before *BeforeCtx
	after  *AfterCtx
}

// EntryName returns entry name
func (mock *optionSetMock) EntryName() string {
	return "mock"
}

// EntryKind returns entry kind
func (mock *optionSetMock) EntryKind() string {
	return "mock"
}

// BeforeCtx should be created before this
func (mock *optionSetMock) BeforeCtx(request *http.Request) *BeforeCtx {
	return mock.before
}

// Before should run before user handler
func (mock *optionSetMock) Before(ctx *BeforeCtx) {
	return
}

// AfterCtx should be created before After()
func (mock *optionSetMock) AfterCtx(reqId, traceId, resCode string) *AfterCtx {
	return mock.after
}

// After should run after user handler
func (mock *optionSetMock) After(before *BeforeCtx, after *AfterCtx) {
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
	ctx.Input.Fields = make([]zap.Field, 0)
	return ctx
}

// NewAfterCtx create new AfterCtx with fields initialized
func NewAfterCtx() *AfterCtx {
	ctx := &AfterCtx{}
	return ctx
}

// BeforeCtx context for Before() function
type BeforeCtx struct {
	Input struct {
		Path       string
		RemoteAddr string
		Method     string
		RawQuery   string
		Protocol   string
		UserAgent  string
		Fields     []zap.Field
	}
	Output struct {
		Event  rkquery.Event
		Logger *zap.Logger
	}
}

// AfterCtx context for After() function
type AfterCtx struct {
	Input struct {
		RequestId string
		TraceId   string
		ResCode   string
	}
	Output struct{}
}

// ***************** BootConfig *****************

// BootConfig for YAML
type BootConfig struct {
	Enabled bool `yaml:"enabled"`
	Zap     struct {
		Encoding    string   `yaml:"encoding"`
		OutputPaths []string `yaml:"outputPaths"`
	} `yaml:"zap"`
	Event struct {
		Encoding    string   `yaml:"encoding"`
		OutputPaths []string `yaml:"outputPaths"`
	} `yaml:"event"`
	Ignore []string `yaml:"ignore"`
}

// ToOptions convert BootConfig into Option list
func ToOptions(config *BootConfig, name, kind string,
	zapEntry *rk.ZapEntry, eventEntry *rk.EventEntry) []Option {
	opts := make([]Option, 0)

	if config.Enabled {
		opts = append(opts,
			WithEntryNameAndKind(name, kind),
			WithZapEntry(zapEntry),
			WithZapEncoding(config.Zap.Encoding),
			WithZapOutputPaths(config.Zap.OutputPaths...),
			WithEventEntry(eventEntry),
			WithEventEncoding(config.Event.Encoding),
			WithEventOutputPaths(config.Event.OutputPaths...),
			WithPathToIgnore(config.Ignore...))
	}

	return opts
}

// ***************** Option *****************

type Option func(*optionSet)

// WithEntryNameAndKind provide entry name and entry kind.
func WithEntryNameAndKind(name, kind string) Option {
	return func(set *optionSet) {
		set.entryName = name
		set.entryKind = kind
	}
}

// WithZapEntry provide rk.ZapEntry.
func WithZapEntry(entry *rk.ZapEntry) Option {
	return func(set *optionSet) {
		if entry != nil {
			set.zap.entry = entry
		}
	}
}

// WithEventEntry provide rk.EventEntry.
func WithEventEntry(entry *rk.EventEntry) Option {
	return func(set *optionSet) {
		if entry != nil {
			set.event.entry = entry
		}
	}
}

// WithZapEncoding provide ZapLoggerEncodingType.
// json or console is supported.
func WithZapEncoding(ec string) Option {
	return func(set *optionSet) {
		set.zap.encoding = strings.ToLower(ec)
	}
}

// WithZapOutputPaths provide ZapLogger Output Path.
// Multiple output path could be supported including stdout.
func WithZapOutputPaths(paths ...string) Option {
	return func(set *optionSet) {
		set.zap.outputPaths = append(set.zap.outputPaths, paths...)
	}
}

// WithEventEncoding provide ZapLoggerEncodingType.
// Console or Json is supported.
func WithEventEncoding(ec string) Option {
	return func(set *optionSet) {
		set.event.encoding = rkquery.ToEncoding(ec)
	}
}

// WithEventOutputPaths provide EventLogger Output Path.
// Multiple output path could be supported including stdout.
func WithEventOutputPaths(paths ...string) Option {
	return func(set *optionSet) {
		set.event.outputPaths = append(set.event.outputPaths, paths...)
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

// Make incoming paths to absolute path with current working directory attached as prefix
func toAbsPath(p ...string) []string {
	res := make([]string, 0)

	for i := range p {
		if path.IsAbs(p[i]) {
			res = append(res, p[i])
		}
		wd, _ := os.Getwd()
		res = append(res, path.Join(wd, p[i]))
	}

	return res
}
