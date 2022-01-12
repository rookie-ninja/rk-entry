// Copyright (c) 2021 rookie-ninja
//
// Use of this source code is governed by an Apache-style
// license that can be found in the LICENSE file.

// package rkmidlog provide options
package rkmidlog

import (
	"github.com/rookie-ninja/rk-common/common"
	"github.com/rookie-ninja/rk-entry/entry"
	"github.com/rookie-ninja/rk-entry/entry/middleware"
	"github.com/rookie-ninja/rk-logger"
	"github.com/rookie-ninja/rk-query"
	"github.com/rs/xid"
	"go.uber.org/zap"
	"net/http"
	"os"
	"path"
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
}

// ***************** OptionSet Implementation *****************

// optionSet which is used for middleware implementation
type optionSet struct {
	entryName             string
	entryType             string
	zapLoggerEntry        *rkentry.ZapLoggerEntry
	eventLoggerEntry      *rkentry.EventLoggerEntry
	zapLogger             *zap.Logger
	zapLoggerEncoding     string
	eventLoggerEncoding   rkquery.Encoding
	zapLoggerOutputPath   []string
	eventLoggerOutputPath []string
	eventLoggerOverride   *zap.Logger
	ignorePrefix          []string
	mock                  OptionSetInterface
}

// NewOptionSet Create new optionSet with options.
func NewOptionSet(opts ...Option) OptionSetInterface {
	set := &optionSet{
		entryName:             xid.New().String(),
		entryType:             "",
		zapLoggerEntry:        rkentry.GlobalAppCtx.GetZapLoggerEntryDefault(),
		eventLoggerEntry:      rkentry.GlobalAppCtx.GetEventLoggerEntryDefault(),
		zapLogger:             rkentry.GlobalAppCtx.GetZapLoggerEntryDefault().GetLogger(),
		zapLoggerOutputPath:   make([]string, 0),
		eventLoggerOutputPath: make([]string, 0),
		ignorePrefix:          []string{},
	}

	for i := range opts {
		opts[i](set)
	}

	if set.mock != nil {
		return set.mock
	}

	set.zapLogger = set.zapLoggerEntry.GetLogger()

	// Override zap logger encoding and output path if provided by user
	// Override encoding type
	if set.zapLoggerEncoding == json || len(set.zapLoggerOutputPath) > 0 {
		if set.zapLoggerEncoding == json {
			set.zapLoggerEntry.LoggerConfig.Encoding = "json"
		}

		if len(set.zapLoggerOutputPath) > 0 {
			set.zapLoggerEntry.LoggerConfig.OutputPaths = toAbsPath(set.zapLoggerOutputPath...)
		}

		if set.zapLoggerEntry.LumberjackConfig == nil {
			set.zapLoggerEntry.LumberjackConfig = rklogger.NewLumberjackConfigDefault()
		}

		if logger, err := rklogger.NewZapLoggerWithConf(set.zapLoggerEntry.LoggerConfig, set.zapLoggerEntry.LumberjackConfig); err != nil {
			rkcommon.ShutdownWithError(err)
		} else {
			set.zapLogger = logger
		}
	}

	// Override event logger output path if provided by user
	if len(set.eventLoggerOutputPath) > 0 {
		set.eventLoggerEntry.LoggerConfig.OutputPaths = toAbsPath(set.eventLoggerOutputPath...)
		if set.eventLoggerEntry.LumberjackConfig == nil {
			set.eventLoggerEntry.LumberjackConfig = rklogger.NewLumberjackConfigDefault()
		}
		if logger, err := rklogger.NewZapLoggerWithConf(set.eventLoggerEntry.LoggerConfig, set.eventLoggerEntry.LumberjackConfig); err != nil {
			rkcommon.ShutdownWithError(err)
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

// EventLoggerEntry returns rkentry.EventLoggerEntry
func (set *optionSet) EventLoggerEntry() *rkentry.EventLoggerEntry {
	return set.eventLoggerEntry
}

// ZapLoggerEntry returns rkentry.ZapLoggerEntry
func (set *optionSet) ZapLoggerEntry() *rkentry.ZapLoggerEntry {
	return set.zapLoggerEntry
}

// CreateEvent create event based on urlPath
func (set *optionSet) createEvent(urlPath string, threadSafe bool) rkquery.Event {
	if set.ignore(urlPath) {
		return set.EventLoggerEntry().EventFactory.CreateEventNoop()
	}

	var event rkquery.Event
	if threadSafe {
		event = set.eventLoggerEntry.GetEventFactory().CreateEventThreadSafe(
			rkquery.WithZapLogger(set.eventLoggerOverride),
			rkquery.WithEncoding(set.eventLoggerEncoding),
			rkquery.WithAppName(rkentry.GlobalAppCtx.GetAppInfoEntry().AppName),
			rkquery.WithAppVersion(rkentry.GlobalAppCtx.GetAppInfoEntry().Version),
			rkquery.WithEntryName(set.GetEntryName()),
			rkquery.WithEntryType(set.GetEntryType()))
	} else {
		event = set.eventLoggerEntry.GetEventFactory().CreateEvent(
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

// Ignore determine whether auth should be ignored based on path
func (set *optionSet) ignore(path string) bool {
	for i := range set.ignorePrefix {
		if strings.HasPrefix(path, set.ignorePrefix[i]) {
			return true
		}
	}

	return rkmid.IgnorePrefixGlobal(path)
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

// ***************** Context *****************

// NewBeforeCtx create new BeforeCtx with fields initialized
func NewBeforeCtx() *BeforeCtx {
	ctx := &BeforeCtx{}
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
	Enabled                bool     `yaml:"enabled" json:"enabled"`
	ZapLoggerEncoding      string   `yaml:"zapLoggerEncoding" json:"zapLoggerEncoding"`
	ZapLoggerOutputPaths   []string `yaml:"zapLoggerOutputPaths" json:"zapLoggerOutputPaths"`
	EventLoggerEncoding    string   `yaml:"eventLoggerEncoding" json:"eventLoggerEncoding"`
	EventLoggerOutputPaths []string `yaml:"eventLoggerOutputPaths" json:"eventLoggerOutputPaths"`
	IgnorePrefix           []string `yaml:"ignorePrefix" json:"ignorePrefix"`
}

// ToOptions convert BootConfig into Option list
func ToOptions(config *BootConfig,
	entryName, entryType string,
	zapLoggerEntry *rkentry.ZapLoggerEntry,
	eventLoggerEntry *rkentry.EventLoggerEntry) []Option {
	opts := make([]Option, 0)

	if config.Enabled {
		opts = append(opts,
			WithEntryNameAndType(entryName, entryType),
			WithEventLoggerEntry(eventLoggerEntry),
			WithZapLoggerEntry(zapLoggerEntry),
			WithZapLoggerEncoding(config.ZapLoggerEncoding),
			WithEventLoggerEncoding(config.EventLoggerEncoding),
			WithZapLoggerOutputPaths(config.ZapLoggerOutputPaths...),
			WithEventLoggerOutputPaths(config.EventLoggerOutputPaths...),
			WithIgnorePrefix(config.IgnorePrefix...))
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

// WithZapLoggerEntry provide rkentry.ZapLoggerEntry.
func WithZapLoggerEntry(zapLoggerEntry *rkentry.ZapLoggerEntry) Option {
	return func(set *optionSet) {
		if zapLoggerEntry != nil {
			set.zapLoggerEntry = zapLoggerEntry
		}
	}
}

// WithEventLoggerEntry provide rkentry.EventLoggerEntry.
func WithEventLoggerEntry(eventLoggerEntry *rkentry.EventLoggerEntry) Option {
	return func(set *optionSet) {
		if eventLoggerEntry != nil {
			set.eventLoggerEntry = eventLoggerEntry
		}
	}
}

// WithZapLoggerEncoding provide ZapLoggerEncodingType.
// json or console is supported.
func WithZapLoggerEncoding(ec string) Option {
	return func(set *optionSet) {
		set.zapLoggerEncoding = strings.ToLower(ec)
	}
}

// WithZapLoggerOutputPaths provide ZapLogger Output Path.
// Multiple output path could be supported including stdout.
func WithZapLoggerOutputPaths(path ...string) Option {
	return func(set *optionSet) {
		set.zapLoggerOutputPath = append(set.zapLoggerOutputPath, path...)
	}
}

// WithEventLoggerEncoding provide ZapLoggerEncodingType.
// Console or Json is supported.
func WithEventLoggerEncoding(ec string) Option {
	return func(set *optionSet) {
		switch strings.ToLower(ec) {
		case console:
			set.eventLoggerEncoding = rkquery.CONSOLE
		case json:
			set.eventLoggerEncoding = rkquery.JSON
		default:
			set.eventLoggerEncoding = rkquery.CONSOLE
		}
	}
}

// WithEventLoggerOutputPaths provide EventLogger Output Path.
// Multiple output path could be supported including stdout.
func WithEventLoggerOutputPaths(path ...string) Option {
	return func(set *optionSet) {
		set.eventLoggerOutputPath = append(set.eventLoggerOutputPath, path...)
	}
}

// WithIgnorePrefix provide paths prefix that will ignore.
// Mainly used for swagger main page and RK TV entry.
func WithIgnorePrefix(paths ...string) Option {
	return func(set *optionSet) {
		set.ignorePrefix = append(set.ignorePrefix, paths...)
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
