// Copyright (c) 2021 rookie-ninja
//
// Use of this source code is governed by an Apache-style
// license that can be found in the LICENSE file.

// Package prom provide options
package prom

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/rookie-ninja/rk-entry/v3/middleware"
	"net/http"
	"strings"
	"time"
)

var (
	// labelKeysHttp are default labels for prometheus metrics
	labelKeysHttp = []string{
		"entryName",
		"entryKind",
		"domain",
		"instance",
		"restMethod",
		"restPath",
		"resCode",
	}

	// labelKeysGrpc are default labels for prometheus metrics
	labelKeysGrpc = []string{
		"entryName",
		"entryKind",
		"domain",
		"instance",
		"grpcService",
		"grpcMethod",
		"grpcType",
		"restMethod",
		"restPath",
		"resCode",
	}
)

const (
	// LabelerTypeHttp type of labeler
	LabelerTypeHttp = "http"
	// LabelerTypeGrpc type of labeler
	LabelerTypeGrpc = "grpc"
)

// ***************** OptionSet Interface *****************

// OptionSetInterface mainly for testing purpose
type OptionSetInterface interface {
	EntryName() string

	EntryKind() string

	BeforeCtx(*http.Request) *BeforeCtx

	Before(*BeforeCtx)

	AfterCtx(string) *AfterCtx

	After(before *BeforeCtx, after *AfterCtx)

	ShouldIgnore(string) bool
}

// ***************** OptionSet Implementation *****************

// optionSet which is used for middleware implementation
type optionSet struct {
	entryName    string
	entryKind    string
	registerer   prometheus.Registerer
	labelerType  string
	pathToIgnore []string
	metricsSet   *MetricsSet
	mock         OptionSetInterface
}

// NewOptionSet Create new optionSet with options.
func NewOptionSet(opts ...Option) OptionSetInterface {
	set := &optionSet{
		entryName:    "fake-entry",
		entryKind:    "",
		registerer:   prometheus.DefaultRegisterer,
		pathToIgnore: []string{},
		labelerType:  LabelerTypeHttp,
	}

	for i := range opts {
		opts[i](set)
	}

	if set.mock != nil {
		return set.mock
	}

	set.metricsSet = NewMetricsSet(
		"rk",
		"prom",
		set.registerer)

	if _, ok := optionsMap[set.entryName]; !ok {
		optionsMap[set.entryName] = set
	}

	var keys []string

	switch set.labelerType {
	case LabelerTypeHttp:
		keys = labelKeysHttp
	case LabelerTypeGrpc:
		keys = labelKeysGrpc
	default:
		keys = labelKeysHttp
	}

	set.metricsSet.RegisterSummary(MetricsNameElapsedNano, SummaryObjectives, keys...)
	set.metricsSet.RegisterCounter(MetricsNameResCode, keys...)

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
func (set *optionSet) BeforeCtx(req *http.Request) *BeforeCtx {
	ctx := NewBeforeCtx()
	ctx.Output.StartTime = time.Now()

	if req != nil && req.URL != nil {
		ctx.Input.RestMethod = req.Method
		ctx.Input.RestPath = req.URL.Path
	}

	return ctx
}

// Before noop
func (set *optionSet) Before(ctx *BeforeCtx) {}

// AfterCtx should be created before After()
func (set *optionSet) AfterCtx(resCode string) *AfterCtx {
	ctx := NewAfterCtx()
	ctx.Input.ResCode = resCode
	return ctx
}

// After should run after user handler
func (set *optionSet) After(before *BeforeCtx, after *AfterCtx) {
	if before == nil || after == nil {
		return
	}

	if set.ShouldIgnore(before.Input.RestPath) {
		return
	}

	var l labeler

	switch set.labelerType {
	case LabelerTypeGrpc:
		l = &labelerGrpc{
			entryName:   set.entryName,
			entryKind:   set.entryKind,
			domain:      rkm.Domain.String,
			instance:    rkm.LocalHostname.String,
			restPath:    before.Input.RestPath,
			restMethod:  before.Input.RestMethod,
			grpcType:    before.Input.GrpcType,
			grpcService: before.Input.GrpcService,
			grpcMethod:  before.Input.GrpcMethod,
			resCode:     after.Input.ResCode,
		}
	case LabelerTypeHttp:
		l = &labelerHttp{
			entryName: set.entryName,
			entryKind: set.entryKind,
			domain:    rkm.Domain.String,
			instance:  rkm.LocalHostname.String,
			method:    before.Input.RestMethod,
			path:      before.Input.RestPath,
			resCode:   after.Input.ResCode,
		}
	default:
		l = &labelerHttp{
			entryName: set.entryName,
			entryKind: set.entryKind,
			domain:    rkm.Domain.String,
			instance:  rkm.LocalHostname.String,
			method:    before.Input.RestMethod,
			path:      before.Input.RestPath,
			resCode:   after.Input.ResCode,
		}
	}

	elapsed := time.Now().Sub(before.Output.StartTime)

	if durationMetrics := set.getServerDurationMetrics(l); durationMetrics != nil {
		durationMetrics.Observe(float64(elapsed.Nanoseconds()))
	}

	if resCodeMetrics := set.getServerResCodeMetrics(l); resCodeMetrics != nil {
		resCodeMetrics.Inc()
	}
}

// getServerDurationMetrics server request elapsed metrics.
func (set *optionSet) getServerDurationMetrics(l labeler) prometheus.Observer {
	return set.metricsSet.GetSummaryWithValues(MetricsNameElapsedNano, l.Values()...)
}

// getServerResCodeMetrics server response code metrics.
func (set *optionSet) getServerResCodeMetrics(l labeler) prometheus.Counter {
	return set.metricsSet.GetCounterWithValues(MetricsNameResCode, l.Values()...)
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

// BeforeCtx should be created before Before()
func (mock *optionSetMock) BeforeCtx(request *http.Request) *BeforeCtx {
	return mock.before
}

// Before should run before user handler
func (mock *optionSetMock) Before(ctx *BeforeCtx) {
	return
}

// AfterCtx should be created before After()
func (mock *optionSetMock) AfterCtx(string) *AfterCtx {
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
		RestMethod  string
		RestPath    string
		GrpcType    string
		GrpcMethod  string
		GrpcService string
	}
	Output struct {
		StartTime time.Time
	}
}

// AfterCtx context for After() function
type AfterCtx struct {
	Input struct {
		ResCode string
	}
	Output struct{}
}

// ***************** BootConfig *****************

// BootConfig for YAML
type BootConfig struct {
	Enabled bool     `yaml:"enabled" json:"enabled"`
	Ignore  []string `yaml:"ignore" json:"ignore"`
}

// ToOptions convert BootConfig into Option list
func ToOptions(config *BootConfig, name, kind string,
	reg *prometheus.Registry, labelerType string) []Option {
	opts := make([]Option, 0)

	if config.Enabled {
		opts = append(opts,
			WithEntryNameAndKind(name, kind),
			WithRegisterer(reg),
			WithLabelerType(labelerType),
			WithPathToIgnore(config.Ignore...))
	}

	return opts
}

// ***************** Option *****************

// Option options provided to Interceptor or optionsSet while creating
type Option func(*optionSet)

// WithEntryNameAndKind provide entry name and entry kind.
func WithEntryNameAndKind(name, kind string) Option {
	return func(opt *optionSet) {
		if len(name) > 0 {
			opt.entryName = name
		}

		if len(kind) > 0 {
			opt.entryKind = kind
		}
	}
}

// WithRegisterer provide prometheus.Registerer.
func WithRegisterer(registerer prometheus.Registerer) Option {
	return func(opt *optionSet) {
		if registerer != nil {
			opt.registerer = registerer
		}
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

// WithLabelerType provide Labeler which will init metrics based on that
func WithLabelerType(l string) Option {
	return func(opt *optionSet) {
		opt.labelerType = strings.ToLower(l)
	}
}

// WithMockOptionSet provide mock OptionSetInterface
func WithMockOptionSet(mock OptionSetInterface) Option {
	return func(set *optionSet) {
		set.mock = mock
	}
}

// ***************** Labeler *****************

// Interface for labeler
type labeler interface {
	Keys() []string

	Values() []string
}

// Implementation of labeler for http request
type labelerHttp struct {
	entryName string
	entryKind string
	domain    string
	instance  string
	method    string
	path      string
	resCode   string
}

// Keys returns key set
func (l *labelerHttp) Keys() []string {
	return labelKeysHttp
}

// Values return value set
func (l *labelerHttp) Values() []string {
	return []string{
		l.entryName,
		l.entryKind,
		l.domain,
		l.instance,
		l.method,
		l.path,
		l.resCode,
	}
}

// Implementation of labeler for gRPC
type labelerGrpc struct {
	entryName   string
	entryKind   string
	domain      string
	instance    string
	restMethod  string
	restPath    string
	grpcService string
	grpcMethod  string
	grpcType    string
	resCode     string
}

// Keys returns key set
func (l *labelerGrpc) Keys() []string {
	return labelKeysGrpc
}

// Values returns value set
func (l *labelerGrpc) Values() []string {
	return []string{
		l.entryName,
		l.entryKind,
		l.domain,
		l.instance,
		l.grpcService,
		l.grpcMethod,
		l.grpcType,
		l.restMethod,
		l.restPath,
		l.resCode,
	}
}

// ***************** Global functions *****************

const (
	// MetricsNameElapsedNano records RPC duration
	MetricsNameElapsedNano = "elapsedNano"
	// MetricsNameResCode records response code
	MetricsNameResCode = "resCode"
)

// Global map stores metrics sets
// Interceptor would distinguish metrics set based on
var optionsMap = make(map[string]*optionSet)

// GetServerMetricsSet server metrics set.
func GetServerMetricsSet(entryName string) *MetricsSet {
	if set, ok := optionsMap[entryName]; ok {
		return set.metricsSet
	}

	return nil
}

// Internal use only.
func ClearAllMetrics() {
	for _, v := range optionsMap {
		v.metricsSet.UnRegisterSummary(MetricsNameElapsedNano)
		v.metricsSet.UnRegisterCounter(MetricsNameResCode)
	}

	optionsMap = make(map[string]*optionSet)
}
