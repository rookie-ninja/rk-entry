// Copyright (c) 2021 rookie-ninja
//
// Use of this source code is governed by an Apache-style
// license that can be found in the LICENSE file.

// Package rkmidlog provide options
package rkmidlog

import (
	"github.com/rookie-ninja/rk-entry/v2/entry"
	"github.com/rookie-ninja/rk-entry/v2/middleware"
	"github.com/rookie-ninja/rk-logger"
	"github.com/rookie-ninja/rk-query"
	"go.uber.org/zap"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const (
	// Console console encoding style of logging
	console = "console"
	// Json console encoding style of logging
	json = "json"
)

// ***************** OptionSet Interface *****************

// OptionSetInterface mainly for testing purpose
type OptionSetInterface interface {
	GetEntryName() string

	GetEntryType() string

	BeforeCtx(*http.Request) *BeforeCtx

	Before(*BeforeCtx)

	AfterCtx(reqId, traceId, resCode string) *AfterCtx

	After(before *BeforeCtx, after *AfterCtx)

	ShouldIgnore(string) bool
}

// ***************** OptionSet Implementation *****************

// optionSet which is used for middleware implementation
type optionSet struct {
	entryName             string
	entryType             string
	loggerEntry           *rkentry.LoggerEntry
	eventEntry            *rkentry.EventEntry
	zapLogger             *zap.Logger
	zapLoggerEncoding     string
	eventLoggerEncoding   rkquery.Encoding
	zapLoggerOutputPath   []string
	eventLoggerOutputPath []string
	eventLoggerOverride   *zap.Logger
	pathToIgnore          []string
	mock                  OptionSetInterface
}

// NewOptionSet Create new optionSet with options.
func NewOptionSet(opts ...Option) OptionSetInterface {
	set := &optionSet{
		entryName:             "fake-entry",
		entryType:             "",
		loggerEntry:           rkentry.LoggerEntryStdout,
		eventEntry:            rkentry.EventEntryStdout,
		zapLogger:             rkentry.LoggerEntryStdout.Logger,
		zapLoggerOutputPath:   make([]string, 0),
		eventLoggerOutputPath: make([]string, 0),
		pathToIgnore:          []string{},
	}

	for i := range opts {
		opts[i](set)
	}

	if set.mock != nil {
		return set.mock
	}

	set.zapLogger = set.loggerEntry.Logger

	// Override zap logger encoding and output path if provided by user
	// Override encoding type
	if set.zapLoggerEncoding == json || len(set.zapLoggerOutputPath) > 0 {
		if set.zapLoggerEncoding == json {
			set.loggerEntry.LoggerConfig.Encoding = "json"
		}

		if len(set.zapLoggerOutputPath) > 0 {
			set.loggerEntry.LoggerConfig.OutputPaths = toAbsPath(set.zapLoggerOutputPath...)
		}

		if set.loggerEntry.LumberjackConfig == nil {
			set.loggerEntry.LumberjackConfig = rklogger.NewLumberjackConfigDefault()
		}

		if logger, err := rklogger.NewZapLoggerWithConf(set.loggerEntry.LoggerConfig, set.loggerEntry.LumberjackConfig); err != nil {
			rkentry.ShutdownWithError(err)
		} else {
			set.zapLogger = logger.WithOptions(zap.WithCaller(true))
		}
	}

	// Override event logger output path if provided by user
	if len(set.eventLoggerOutputPath) > 0 {
		set.eventEntry.LoggerConfig.OutputPaths = toAbsPath(set.eventLoggerOutputPath...)
		if set.eventEntry.LumberjackConfig == nil {
			set.eventEntry.LumberjackConfig = rklogger.NewLumberjackConfigDefault()
		}
		if logger, err := rklogger.NewZapLoggerWithConf(set.eventEntry.LoggerConfig, set.eventEntry.LumberjackConfig); err != nil {
			rkentry.ShutdownWithError(err)
		} else {
			set.eventLoggerOverride = logger
		}
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
func (set *optionSet) BeforeCtx(req *http.Request) *BeforeCtx {
	ctx := NewBeforeCtx()

	if req != nil && req.URL != nil && req.Header != nil {
		remoteIp, remotePort := rkmid.GetRemoteAddressSet(req)
		ctx.Input.RemoteAddr = remoteIp + ":" + remotePort

		ctx.Input.UrlPath = req.URL.Path
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

	ctx.Output.Event = set.createEvent(ctx.Input.UrlPath, true)
	ctx.Output.Logger = set.zapLogger

	ctx.Output.Event.SetRemoteAddr(ctx.Input.RemoteAddr)

	ctx.Output.Event.AddPayloads([]zap.Field{
		zap.String("apiPath", ctx.Input.UrlPath),
		zap.String("apiMethod", ctx.Input.Method),
		zap.String("apiQuery", ctx.Input.RawQuery),
		zap.String("apiProtocol", ctx.Input.Protocol),
		zap.String("userAgent", ctx.Input.UserAgent),
	}...)

	ctx.Output.Event.AddPayloads(ctx.Input.Fields...)

	ctx.Output.Event.SetOperation(ctx.Input.UrlPath)
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

// EventEntry returns rkentry.EventEntry
func (set *optionSet) EventEntry() *rkentry.EventEntry {
	return set.eventEntry
}

// ZapEntry returns rkentry.ZapEntry
func (set *optionSet) ZapEntry() *rkentry.LoggerEntry {
	return set.loggerEntry
}

// CreateEvent create event based on urlPath
func (set *optionSet) createEvent(urlPath string, threadSafe bool) rkquery.Event {
	if set.ShouldIgnore(urlPath) {
		return set.EventEntry().EventFactory.CreateEventNoop()
	}

	var event rkquery.Event
	if threadSafe {
		event = set.eventEntry.EventFactory.CreateEventThreadSafe(
			rkquery.WithZapLogger(set.eventLoggerOverride),
			rkquery.WithEncoding(set.eventLoggerEncoding),
			rkquery.WithAppName(rkentry.GlobalAppCtx.GetAppInfoEntry().AppName),
			rkquery.WithAppVersion(rkentry.GlobalAppCtx.GetAppInfoEntry().Version),
			rkquery.WithEntryName(set.GetEntryName()),
			rkquery.WithEntryType(set.GetEntryType()))
	} else {
		event = set.eventEntry.EventFactory.CreateEvent(
			rkquery.WithZapLogger(set.eventLoggerOverride),
			rkquery.WithEncoding(set.eventLoggerEncoding),
			rkquery.WithAppName(rkentry.GlobalAppCtx.GetAppInfoEntry().AppName),
			rkquery.WithAppVersion(rkentry.GlobalAppCtx.GetAppInfoEntry().Version),
			rkquery.WithEntryName(set.GetEntryName()),
			rkquery.WithEntryType(set.GetEntryType()))
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

	return rkmid.ShouldIgnoreGlobal(path)
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

// GetEntryName returns entry name
func (mock *optionSetMock) GetEntryName() string {
	return "mock"
}

// GetEntryType returns entry type
func (mock *optionSetMock) GetEntryType() string {
	return "mock"
}

// BeforeCtx should be created before Before()
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
		UrlPath    string
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
	Enabled           bool     `yaml:"enabled" json:"enabled"`
	LoggerEncoding    string   `yaml:"loggerEncoding" json:"loggerEncoding"`
	LoggerOutputPaths []string `yaml:"loggerOutputPaths" json:"loggerOutputPaths"`
	EventEncoding     string   `yaml:"eventEncoding" json:"eventEncoding"`
	EventOutputPaths  []string `yaml:"eventOutputPaths" json:"eventOutputPaths"`
	Ignore            []string `yaml:"ignore" json:"ignore"`
}

// ToOptions convert BootConfig into Option list
func ToOptions(config *BootConfig,
	entryName, entryType string,
	loggerEntry *rkentry.LoggerEntry,
	eventEntry *rkentry.EventEntry) []Option {
	opts := make([]Option, 0)

	if config.Enabled {
		opts = append(opts,
			WithEntryNameAndType(entryName, entryType),
			WithEventEntry(eventEntry),
			WithLoggerEntry(loggerEntry),
			WithLoggerEncoding(config.LoggerEncoding),
			WithEventEncoding(config.EventEncoding),
			WithLoggerOutputPaths(config.LoggerOutputPaths...),
			WithEventOutputPaths(config.EventOutputPaths...),
			WithPathToIgnore(config.Ignore...))
	}

	return opts
}

// ***************** Option *****************

// Option
type Option func(*optionSet)

// WithEntryNameAndType provide entry name and entry type.
func WithEntryNameAndType(entryName, entryType string) Option {
	return func(set *optionSet) {
		set.entryName = entryName
		set.entryType = entryType
	}
}

// WithLoggerEntry provide rkentry.LoggerEntry.
func WithLoggerEntry(loggerEntry *rkentry.LoggerEntry) Option {
	return func(set *optionSet) {
		if loggerEntry != nil {
			set.loggerEntry = loggerEntry
		}
	}
}

// WithEventEntry provide rkentry.EventEntry.
func WithEventEntry(eventEntry *rkentry.EventEntry) Option {
	return func(set *optionSet) {
		if eventEntry != nil {
			set.eventEntry = eventEntry
		}
	}
}

// WithLoggerEncoding provide ZapLoggerEncodingType.
// json or console is supported.
func WithLoggerEncoding(ec string) Option {
	return func(set *optionSet) {
		set.zapLoggerEncoding = strings.ToLower(ec)
	}
}

// WithLoggerOutputPaths provide ZapLogger Output Path.
// Multiple output path could be supported including stdout.
func WithLoggerOutputPaths(path ...string) Option {
	return func(set *optionSet) {
		set.zapLoggerOutputPath = append(set.zapLoggerOutputPath, path...)
	}
}

// WithEventEncoding provide ZapLoggerEncodingType.
// Console or Json is supported.
func WithEventEncoding(ec string) Option {
	return func(set *optionSet) {
		set.eventLoggerEncoding = rkquery.ToEncoding(ec)
	}
}

// WithEventOutputPaths provide EventLogger Output Path.
// Multiple output path could be supported including stdout.
func WithEventOutputPaths(path ...string) Option {
	return func(set *optionSet) {
		set.eventLoggerOutputPath = append(set.eventLoggerOutputPath, path...)
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
		// get file name
		_, fileName := filepath.Split(p[i])

		// empty file name, ignore
		if len(fileName) < 1 {
			continue
		}

		if filepath.IsAbs(p[i]) || p[i] == "stdout" || p[i] == "stderr" {
			res = append(res, p[i])
			continue
		}
		wd, _ := os.Getwd()
		res = append(res, filepath.ToSlash(filepath.Join(wd, p[i])))
	}

	return res
}
