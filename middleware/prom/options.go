// Copyright (c) 2021 rookie-ninja
//
// Use of this source code is governed by an Apache-style
// license that can be found in the LICENSE file.

// Package rkmidprom provide options
package rkmidprom

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/rookie-ninja/rk-entry/entry"
	"github.com/rookie-ninja/rk-entry/middleware"
	"net/http"
	"strings"
	"time"
)

var (
	// labelKeysHttp are default labels for prometheus metrics
	labelKeysHttp = []string{
		"entryName",
		"entryType",
		"realm",
		"region",
		"az",
		"domain",
		"instance",
		"appVersion",
		"appName",
		"restMethod",
		"restPath",
		"resCode",
	}

	// labelKeysGrpc are default labels for prometheus metrics
	labelKeysGrpc = []string{
		"entryName",
		"entryType",
		"realm",
		"region",
		"az",
		"domain",
		"instance",
		"appVersion",
		"appName",
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
	GetEntryName() string

	GetEntryType() string

	BeforeCtx(*http.Request) *BeforeCtx

	Before(*BeforeCtx)

	AfterCtx(string) *AfterCtx

	After(before *BeforeCtx, after *AfterCtx)
}

// ***************** OptionSet Implementation *****************

// optionSet which is used for middleware implementation
type optionSet struct {
	entryName    string
	entryType    string
	registerer   prometheus.Registerer
	labelerType  string
	ignorePrefix []string
	metricsSet   *MetricsSet
	mock         OptionSetInterface
}

// NewOptionSet Create new optionSet with options.
func NewOptionSet(opts ...Option) OptionSetInterface {
	set := &optionSet{
		entryName:    "fake-entry",
		entryType:    "",
		registerer:   prometheus.DefaultRegisterer,
		ignorePrefix: []string{},
		labelerType:  LabelerTypeHttp,
	}

	for i := range opts {
		opts[i](set)
	}

	if set.mock != nil {
		return set.mock
	}

	namespace := strings.ReplaceAll(rkentry.GlobalAppCtx.GetAppInfoEntry().AppName, "-", "_")
	subSystem := strings.ReplaceAll(set.entryName, "-", "_")
	set.metricsSet = NewMetricsSet(
		namespace,
		subSystem,
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

	if set.ignore(before.Input.RestPath) {
		return
	}

	var l labeler

	switch set.labelerType {
	case LabelerTypeGrpc:
		l = &labelerGrpc{
			entryName:   set.entryName,
			entryType:   set.entryType,
			realm:       rkmid.Realm.String,
			region:      rkmid.Region.String,
			az:          rkmid.AZ.String,
			domain:      rkmid.Domain.String,
			instance:    rkmid.LocalHostname.String,
			appVersion:  rkentry.GlobalAppCtx.GetAppInfoEntry().Version,
			appName:     rkentry.GlobalAppCtx.GetAppInfoEntry().AppName,
			restPath:    before.Input.RestPath,
			restMethod:  before.Input.RestMethod,
			grpcType:    before.Input.GrpcType,
			grpcService: before.Input.GrpcService,
			grpcMethod:  before.Input.GrpcMethod,
			resCode:     after.Input.ResCode,
		}
	case LabelerTypeHttp:
		l = &labelerHttp{
			entryName:  set.entryName,
			entryType:  set.entryType,
			realm:      rkmid.Realm.String,
			region:     rkmid.Region.String,
			az:         rkmid.AZ.String,
			domain:     rkmid.Domain.String,
			instance:   rkmid.LocalHostname.String,
			appVersion: rkentry.GlobalAppCtx.GetAppInfoEntry().Version,
			appName:    rkentry.GlobalAppCtx.GetAppInfoEntry().AppName,
			method:     before.Input.RestMethod,
			path:       before.Input.RestPath,
			resCode:    after.Input.ResCode,
		}
	default:
		l = &labelerHttp{
			entryName:  set.entryName,
			entryType:  set.entryType,
			realm:      rkmid.Realm.String,
			region:     rkmid.Region.String,
			az:         rkmid.AZ.String,
			domain:     rkmid.Domain.String,
			instance:   rkmid.LocalHostname.String,
			appVersion: rkentry.GlobalAppCtx.GetAppInfoEntry().Version,
			appName:    rkentry.GlobalAppCtx.GetAppInfoEntry().AppName,
			method:     before.Input.RestMethod,
			path:       before.Input.RestPath,
			resCode:    after.Input.ResCode,
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
func (mock *optionSetMock) AfterCtx(string) *AfterCtx {
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
	Enabled      bool     `yaml:"enabled" json:"enabled"`
	IgnorePrefix []string `yaml:"ignorePrefix" json:"ignorePrefix"`
}

// ToOptions convert BootConfig into Option list
func ToOptions(config *BootConfig,
	entryName, entryType string,
	reg *prometheus.Registry, labelerType string) []Option {
	opts := make([]Option, 0)

	if config.Enabled {
		opts = append(opts,
			WithEntryNameAndType(entryName, entryType),
			WithRegisterer(reg),
			WithLabelerType(labelerType),
			WithIgnorePrefix(config.IgnorePrefix...))
	}

	return opts
}

// ***************** Option *****************

// Option options provided to Interceptor or optionsSet while creating
type Option func(*optionSet)

// WithEntryNameAndType provide entry name and entry type.
func WithEntryNameAndType(entryName, entryType string) Option {
	return func(opt *optionSet) {
		if len(entryName) > 0 {
			opt.entryName = entryName
		}

		if len(entryType) > 0 {
			opt.entryType = entryType
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

// WithIgnorePrefix provide paths prefix that will ignore.
// Mainly used for swagger main page and RK TV entry.
func WithIgnorePrefix(paths ...string) Option {
	return func(set *optionSet) {
		set.ignorePrefix = append(set.ignorePrefix, paths...)
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
	entryName  string
	entryType  string
	realm      string
	region     string
	az         string
	domain     string
	instance   string
	appVersion string
	appName    string
	method     string
	path       string
	resCode    string
}

// Keys returns key set
func (l *labelerHttp) Keys() []string {
	return labelKeysHttp
}

// Values return value set
func (l *labelerHttp) Values() []string {
	return []string{
		l.entryName,
		l.entryType,
		l.realm,
		l.region,
		l.az,
		l.domain,
		l.instance,
		l.appVersion,
		l.appName,
		l.method,
		l.path,
		l.resCode,
	}
}

// Implementation of labeler for gRPC
type labelerGrpc struct {
	entryName   string
	entryType   string
	realm       string
	region      string
	az          string
	domain      string
	instance    string
	appVersion  string
	appName     string
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
		l.entryType,
		l.realm,
		l.region,
		l.az,
		l.domain,
		l.instance,
		l.appVersion,
		l.appName,
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
